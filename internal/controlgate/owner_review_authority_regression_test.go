package controlgate

// P0 regression guardrails ONLY. These tests do not create lifecycle keys,
// revoke an installed user, persist owner consent or expose an APPLY path.
import (
	"crypto/ed25519"
	"crypto/rand"
	"strings"
	"testing"
	"time"
)

func TestOfflineReviewerMissingAndChangedKeyFailsClosed(t *testing.T) {
	now := time.Date(2026, 10, 14, 9, 0, 0, 0, time.UTC)
	agentPub, agentPriv, oldPub, oldPriv := reviewKeys(t)
	agent, review := offlineFixtures(t, agentPriv, oldPriv, now)
	var missing *OfflineReviewVerifier
	result := missing.Check(agent, review, now)
	if result.Status != "REJECTED" || result.CryptographicMatch || result.HumanOwnerAuthenticated || result.GrantsApply || result.KernelWrites || result.DurableReplay {
		t.Fatalf("missing reviewer must fail closed: %+v", result)
	}
	nextPub, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	changed := newReviewer(t, agentPub, nextPub)
	result = changed.Check(agent, review, now)
	if result.Status != "REJECTED" || result.CryptographicMatch || result.GrantsApply || result.KernelWrites {
		t.Fatalf("old signature accepted under a different pinned key: %+v", result)
	}
	// There is no real revocation API today. This test does NOT claim that a
	// revoked persisted key cannot be restored from a stale snapshot.
	current := newReviewer(t, agentPub, oldPub).Check(agent, review, now)
	if current.Status != "CRYPTOGRAPHIC_TEST_MATCH_ONLY" || current.HumanOwnerAuthenticated || current.GrantsApply || current.KernelWrites {
		t.Fatalf("test-only signature check changed authority: %+v", current)
	}
}

func TestOfflineReviewerClockBoundariesNeverGrantApply(t *testing.T) {
	base := time.Date(2026, 10, 14, 9, 0, 0, 0, time.UTC)
	ap, ak, rp, rk := reviewKeys(t)
	a, r := offlineFixtures(t, ak, rk, base)
	checks := []struct {
		name   string
		at     time.Time
		expect string
	}{
		{"before review issued", base.Add(-time.Nanosecond), "REJECTED"},
		{"review issuance instant", base, "CRYPTOGRAPHIC_TEST_MATCH_ONLY"},
		{"inside review window", base.Add(time.Second), "CRYPTOGRAPHIC_TEST_MATCH_ONLY"},
		{"review expiry instant", base.Add(time.Minute), "REJECTED"},
		{"well after Agent expiry", base.Add(10 * time.Minute), "REJECTED"},
		{"clock backwards beyond tolerance", base.Add(-time.Hour), "REJECTED"},
	}
	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			result := newReviewer(t, ap, rp).Check(a, r, c.at)
			if result.Status != c.expect || result.HumanOwnerAuthenticated || result.GrantsApply || result.KernelWrites || result.DurableReplay {
				t.Fatalf("wrong/non-deny-only review at %s: %+v", c.at, result)
			}
		})
	}
}

func TestOfflineReviewNeverTreatsSelfAssertedOwnerRoleAsConsent(t *testing.T) {
	now := time.Date(2026, 10, 14, 9, 0, 0, 0, time.UTC)
	ap, ak, rp, rk := reviewKeys(t)
	a, r := offlineFixtures(t, ak, rk, now)
	if a.Intent.ActorRole != "OWNER" {
		t.Fatal("fixture must exercise self-asserted OWNER role")
	}
	result := newReviewer(t, ap, rp).Check(a, r, now)
	if result.Status != "CRYPTOGRAPHIC_TEST_MATCH_ONLY" || result.HumanOwnerAuthenticated || result.GrantsApply || result.KernelWrites {
		t.Fatalf("Agent owner string promoted into owner consent: %+v", result)
	}
	receipt := gateFromPub(t, ap).Evaluate(a, baseline(t), now)
	if receipt.Decision != "REJECTED" || receipt.ExecutorStatus != "NOT_AVAILABLE" || receipt.KernelWrites {
		t.Fatalf("review evidence bypassed Gate: %+v", receipt)
	}
	if !strings.Contains(receipt.Reason, "no privileged executor") {
		t.Fatalf("unexpected denial reason %q", receipt.Reason)
	}
}
