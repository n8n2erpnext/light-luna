package controlgate

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func reviewKeys(t *testing.T) (ed25519.PublicKey, ed25519.PrivateKey, ed25519.PublicKey, ed25519.PrivateKey) {
	t.Helper()
	agentPub, agentPriv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	reviewPub, reviewPriv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	return agentPub, agentPriv, reviewPub, reviewPriv
}
func offlineFixtures(t *testing.T, agentPriv, reviewPriv ed25519.PrivateKey, now time.Time) (SignedIntent, SignedReviewEvidence) {
	t.Helper()
	p := baseline(t)
	raw := signedFixture(t, bytes.Repeat([]byte{33}, 32), p, now).Intent
	raw.Operation = "APPLY_POLICY"
	signed, err := SignEd25519(agentPriv, raw)
	if err != nil {
		t.Fatal(err)
	}
	digest, err := IntentReviewDigest(raw)
	if err != nil {
		t.Fatal(err)
	}
	e := ReviewEvidence{
		Protocol: OfflineReviewProtocol, Purpose: OfflineReviewPurpose,
		ReviewID: "offline-review-100", ReviewerID: "offline-owner-test",
		SiteID: "synthetic-site", RequestID: raw.RequestID,
		AgentIntentSHA256: digest, Operation: raw.Operation,
		IssuedUTC: now, ExpiresUTC: now.Add(time.Minute),
	}
	return signed, signReview(t, reviewPriv, e)
}
func signReview(t *testing.T, k ed25519.PrivateKey, e ReviewEvidence) SignedReviewEvidence {
	t.Helper()
	b, err := reviewCanonical(e)
	if err != nil {
		t.Fatal(err)
	}
	return SignedReviewEvidence{Evidence: e, SignatureEd25519: hex.EncodeToString(ed25519.Sign(k, append([]byte(offlineReviewDomain), b...)))}
}
func newReviewer(t *testing.T, ap, rp ed25519.PublicKey) *OfflineReviewVerifier {
	t.Helper()
	v, err := NewOfflineReviewVerifier(ap, rp, "synthetic-site", "synthetic-agent", "offline-owner-test")
	if err != nil {
		t.Fatal(err)
	}
	return v
}
func TestSeparateOfflineReviewNeverAuthorizesApply(t *testing.T) {
	now := time.Date(2026, 10, 14, 9, 0, 0, 0, time.UTC)
	ap, ak, rp, rk := reviewKeys(t)
	agent, review := offlineFixtures(t, ak, rk, now)
	assessment := newReviewer(t, ap, rp).Check(agent, review, now)
	if assessment.Status != "CRYPTOGRAPHIC_TEST_MATCH_ONLY" || !assessment.CryptographicMatch ||
		assessment.HumanOwnerAuthenticated || assessment.GrantsApply || assessment.KernelWrites || assessment.DurableReplay {
		t.Fatalf("offline signature must not authorize enforcement %+v", assessment)
	}
	// Even independent test-review evidence is not wired to Gate or able to
	// relax its unconditional signed APPLY denial.
	gate := gateFromPub(t, ap)
	actual := gate.Evaluate(agent, baseline(t), now)
	if actual.Decision != "REJECTED" || actual.KernelWrites || actual.ExecutorStatus != "NOT_AVAILABLE" {
		t.Fatalf("review evidence somehow bypassed denial %+v", actual)
	}
	out, _ := json.Marshal(assessment)
	for _, unsafe := range []string{`"grantsApply":true`, `"humanOwnerAuthenticated":true`, `"kernelWrites":true`} {
		if strings.Contains(string(out), unsafe) {
			t.Fatal("incorrect authority field", unsafe)
		}
	}
}
func TestOfflineReviewReplayAndScopeFailClosed(t *testing.T) {
	now := time.Date(2026, 10, 14, 9, 0, 0, 0, time.UTC)
	ap, ak, rp, rk := reviewKeys(t)
	agent, review := offlineFixtures(t, ak, rk, now)
	v := newReviewer(t, ap, rp)
	if r := v.Check(agent, review, now); r.Status != "CRYPTOGRAPHIC_TEST_MATCH_ONLY" {
		t.Fatal(r)
	}
	if r := v.Check(agent, review, now); r.Status != "REJECTED" || !strings.Contains(r.Reason, "replay") {
		t.Fatal(r)
	}
	// New review ID is NOT a replay escape when the Agent-signed intent
	// is unchanged. The same request must be accounted for independently.
	alternate := review.Evidence
	alternate.ReviewID = "offline-review-different-200"
	if r := v.Check(agent, signReview(t, rk, alternate), now); r.Status != "REJECTED" ||
		!strings.Contains(r.Reason, "replay") {
		t.Fatal("same Agent intent accepted under fresh review ID", r)
	}
	tests := []struct {
		name   string
		modify func(*SignedIntent, *ReviewEvidence)
		clock  time.Time
	}{
		{"review request mismatch", func(a *SignedIntent, e *ReviewEvidence) { e.RequestID = "different-request-id" }, now},
		{"review site mismatch", func(a *SignedIntent, e *ReviewEvidence) { e.SiteID = "foreign-site-test" }, now},
		{"reviewer identity mismatch", func(a *SignedIntent, e *ReviewEvidence) { e.ReviewerID = "different-reviewer" }, now},
		{"cross operation", func(a *SignedIntent, e *ReviewEvidence) { e.Operation = "PREVIEW_POLICY" }, now},
		{"changed nonce with new valid Agent signature", func(a *SignedIntent, e *ReviewEvidence) { a.Intent.Nonce = "changed-nonce-200" }, now},
		{"changed policy hash with new valid Agent signature", func(a *SignedIntent, e *ReviewEvidence) { a.Intent.PolicyHashSHA256 = strings.Repeat("a", 64) }, now},
		{"review future", func(a *SignedIntent, e *ReviewEvidence) {
			e.IssuedUTC = now.Add(time.Minute)
			e.ExpiresUTC = now.Add(90 * time.Second)
		}, now},
		{"review expired", func(a *SignedIntent, e *ReviewEvidence) {}, now.Add(61 * time.Second)},
		{"review starts before Agent", func(a *SignedIntent, e *ReviewEvidence) { e.IssuedUTC = a.Intent.IssuedUTC.Add(-time.Second) }, now},
		{"review exceeds agent expiry", func(a *SignedIntent, e *ReviewEvidence) { e.ExpiresUTC = a.Intent.ExpiresUTC.Add(time.Second) }, now},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ag, rv := offlineFixtures(t, ak, rk, now)
			tc.modify(&ag, &rv.Evidence)
			ag, err := SignEd25519(ak, ag.Intent)
			if err != nil {
				t.Fatal(err)
			}
			rv = signReview(t, rk, rv.Evidence)
			r := newReviewer(t, ap, rp).Check(ag, rv, tc.clock)
			if r.Status != "REJECTED" || r.GrantsApply || r.HumanOwnerAuthenticated || r.KernelWrites {
				t.Fatalf("unsafe review %+v", r)
			}
		})
	}
}
func TestOfflineReviewSignerIsolationAndBadSignatures(t *testing.T) {
	now := time.Date(2026, 10, 14, 9, 0, 0, 0, time.UTC)
	ap, ak, rp, rk := reviewKeys(t)
	agent, review := offlineFixtures(t, ak, rk, now)
	otherPub, otherPriv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewOfflineReviewVerifier(ap, ap, "synthetic-site", "synthetic-agent", "offline-owner-test"); err == nil {
		t.Fatal("same Agent and owner test key accepted")
	}
	if _, err := NewOfflineReviewVerifier(ap, nil, "synthetic-site", "synthetic-agent", "offline-owner-test"); err == nil {
		t.Fatal("empty reviewer key accepted")
	}
	cases := []struct {
		name                string
		agent               SignedIntent
		review              SignedReviewEvidence
		agentPub, reviewPub ed25519.PublicKey
	}{
		{"wrong pinned reviewer", agent, review, ap, otherPub},
		{"wrong pinned agent", agent, review, otherPub, rp},
		{"agent private key cannot sign review", agent, signReview(t, ak, review.Evidence), ap, rp},
		{"other reviewer cannot sign", agent, signReview(t, otherPriv, review.Evidence), ap, rp},
		{"malformed review signature", agent, SignedReviewEvidence{Evidence: review.Evidence, SignatureEd25519: "invalid"}, ap, rp},
	}
	legacy, err := Sign(bytes.Repeat([]byte{42}, 32), agent.Intent)
	if err != nil {
		t.Fatal(err)
	}
	cases = append(cases, struct {
		name                string
		agent               SignedIntent
		review              SignedReviewEvidence
		agentPub, reviewPub ed25519.PublicKey
	}{"HMAC downgrade", legacy, review, ap, rp})
	mixed := agent
	mixed.MACSHA256 = strings.Repeat("1", 64)
	cases = append(cases, struct {
		name                string
		agent               SignedIntent
		review              SignedReviewEvidence
		agentPub, reviewPub ed25519.PublicKey
	}{"mixed Agent auth", mixed, review, ap, rp})
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := newReviewer(t, tc.agentPub, tc.reviewPub).Check(tc.agent, tc.review, now)
			if r.Status != "REJECTED" || r.CryptographicMatch || r.GrantsApply {
				t.Fatalf("unsafe signer boundary %+v", r)
			}
		})
	}
	tampered := review
	tampered.Evidence.ReviewID = "changed-review-uuid"
	if r := newReviewer(t, ap, rp).Check(agent, tampered, now); r.Status != "REJECTED" {
		t.Fatal("review content modified without fresh signature", r)
	}
}

// A hex spelling alias of a valid Ed25519 signature must not create a
// second representation of one Agent envelope. A rejected alias may not
// consume the review ID or intent replay budget.
func TestOfflineReviewRejectsNoncanonicalAgentSignatureWithoutConsumingReview(t *testing.T) {
	now := time.Date(2026, 10, 14, 9, 0, 0, 0, time.UTC)
	ap, ak, rp, rk := reviewKeys(t)
	agent, review := offlineFixtures(t, ak, rk, now)
	v := newReviewer(t, ap, rp)

	alias := agent
	alias.SignatureEd25519 = strings.ToUpper(agent.SignatureEd25519)
	if alias.SignatureEd25519 == agent.SignatureEd25519 {
		t.Fatal("test signature unexpectedly lacks lowercase hex letters")
	}
	denied := v.Check(alias, review, now)
	if denied.Status != "REJECTED" || denied.CryptographicMatch ||
		denied.GrantsApply || denied.KernelWrites || denied.HumanOwnerAuthenticated ||
		len(v.seen) != 0 || len(v.seenRequests) != 0 {
		t.Fatalf("noncanonical Agent signature accepted/consumed: %+v", denied)
	}
	valid := v.Check(agent, review, now)
	if valid.Status != "CRYPTOGRAPHIC_TEST_MATCH_ONLY" || !valid.CryptographicMatch ||
		valid.GrantsApply || valid.KernelWrites || valid.HumanOwnerAuthenticated || valid.DurableReplay {
		t.Fatalf("canonical signature not accepted as test-only evidence: %+v", valid)
	}
	if replay := v.Check(agent, review, now); replay.Status != "REJECTED" {
		t.Fatalf("canonical Agent replay not rejected: %+v", replay)
	}
}

func TestOfflineReviewValidationAndBudget(t *testing.T) {
	now := time.Date(2026, 10, 14, 9, 0, 0, 0, time.UTC)
	ap, ak, rp, rk := reviewKeys(t)
	agent, review := offlineFixtures(t, ak, rk, now)
	long := review.Evidence
	long.ExpiresUTC = long.IssuedUTC.Add(MaxOfflineReviewLifetime + time.Second)
	if _, err := reviewCanonical(long); err == nil {
		t.Fatal("oversized review TTL")
	}
	wrong := review.Evidence
	wrong.Purpose = "AUTHORIZE_APPLY_POLICY"
	if _, err := reviewCanonical(wrong); err == nil {
		t.Fatal("review purpose escalation accepted")
	}
	wrong = review.Evidence
	wrong.AgentIntentSHA256 = strings.Repeat("A", 64)
	if _, err := reviewCanonical(wrong); err == nil {
		t.Fatal("noncanonical digest accepted")
	}
	v := newReviewer(t, ap, rp)
	makeUnique := func(i int) (SignedIntent, SignedReviewEvidence) {
		t.Helper()
		a := agent
		a.Intent.RequestID = "synthetic-request-" + fmtReviewID(i)
		a.Intent.Nonce = "synthetic-nonce-" + fmtReviewID(i)
		signed, err := SignEd25519(ak, a.Intent)
		if err != nil {
			t.Fatal(err)
		}
		digest, err := IntentReviewDigest(a.Intent)
		if err != nil {
			t.Fatal(err)
		}
		e := review.Evidence
		e.ReviewID = fmtReviewID(i)
		e.RequestID = a.Intent.RequestID
		e.AgentIntentSHA256 = digest
		return signed, signReview(t, rk, e)
	}
	for i := 0; i < MaxOfflineReviewSeen; i++ {
		a, r := makeUnique(i)
		result := v.Check(a, r, now)
		if result.Status != "CRYPTOGRAPHIC_TEST_MATCH_ONLY" {
			t.Fatalf("budget unexpectedly denied at %d: %+v", i, result)
		}
	}
	a, r := makeUnique(MaxOfflineReviewSeen)
	if result := v.Check(a, r, now); result.Status != "REJECTED" ||
		!strings.Contains(result.Reason, "budget") {
		t.Fatalf("replay budget not fail closed %+v", result)
	}
}
func fmtReviewID(i int) string {
	// IDs are deterministic and within safeID for test-only bounded budget.
	b := []byte("offline-review-00000")
	for j := len(b) - 1; i > 0 && j >= len("offline-review-"); j-- {
		b[j] = byte('0' + i%10)
		i /= 10
	}
	return string(b)
}
