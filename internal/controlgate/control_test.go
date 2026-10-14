package controlgate

import (
	"bytes"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/n8n2erpnext/light-luna/internal/policy"
)

func baseline(t *testing.T) policy.Bundle {
	t.Helper()
	r, err := policy.Decode(strings.NewReader(`{
 "apiVersion":"lightluna.io/v1alpha1","kind":"PolicyBundle",
 "metadata":{"name":"test-local","revision":4},
 "spec":{"mode":"host","defaultInbound":"accept","emergencyAccess":{"enabled":false,"networks":[]},
 "aliases":[],"rules":[],"autonomous":{"mode":"observe","allowRealBans":false,"maxBanTtlSeconds":0}}
 }`))
	if err != nil {
		t.Fatal(err)
	}
	return r
}
func signedFixture(t *testing.T, key []byte, p policy.Bundle, now time.Time) SignedIntent {
	t.Helper()
	preview, err := policy.MakePreview(p)
	if err != nil {
		t.Fatal(err)
	}
	s, err := Sign(key, Intent{Protocol: Protocol, RequestID: "synthetic-request-100", Nonce: "synthetic-nonce-100",
		SiteID: "synthetic-site", ActorID: "synthetic-agent", ActorRole: "OWNER", Origin: "LOCAL_AGENT",
		PolicyName: p.Metadata.Name, ExpectedRevision: p.Metadata.Revision, PolicyHashSHA256: preview.HashSHA256,
		IssuedUTC: now.Add(-time.Second), ExpiresUTC: now.Add(time.Minute), Operation: "PREVIEW_POLICY"})
	if err != nil {
		t.Fatal(err)
	}
	return s
}
func gateFixture(t *testing.T, key []byte) *Boundary {
	t.Helper()
	b, err := NewBoundary(key, "synthetic-site", "synthetic-agent")
	if err != nil {
		t.Fatal(err)
	}
	return b
}
func TestValidPreviewAndReplay(t *testing.T) {
	key := bytes.Repeat([]byte{41}, 32)
	p := baseline(t)
	now := time.Date(2026, 10, 14, 13, 0, 0, 0, time.UTC)
	msg := signedFixture(t, key, p, now)
	b := gateFixture(t, key)
	first := b.Evaluate(msg, p, now)
	if first.Decision != "ADVISORY_ONLY" || first.KernelWrites || first.ExecutorStatus != "NOT_AVAILABLE" {
		t.Fatal(first)
	}
	second := b.Evaluate(msg, p, now)
	if second.Decision != "REJECTED" || !strings.Contains(second.Reason, "replay") {
		t.Fatal(second)
	}
}
func TestRejectTamperScopeExpiryRevocationAndOrigin(t *testing.T) {
	key := bytes.Repeat([]byte{42}, 32)
	now := time.Date(2026, 10, 14, 13, 0, 0, 0, time.UTC)
	p := baseline(t)
	msg := signedFixture(t, key, p, now)
	tests := []struct {
		name        string
		mutate      func(*SignedIntent)
		when        time.Time
		alterPolicy bool
	}{
		{"tamper-revision", func(m *SignedIntent) { m.Intent.ExpectedRevision++ }, now, false},
		{"tamper-role", func(m *SignedIntent) { m.Intent.ActorRole = "ADMINISTRATOR" }, now, false},
		{"tamper-site", func(m *SignedIntent) { m.Intent.SiteID = "synthetic-other" }, now, false},
		{"tamper-MAC", func(m *SignedIntent) { m.MACSHA256 = strings.Repeat("0", 64) }, now, false},
		{"expired", func(m *SignedIntent) {}, now.Add(2 * time.Minute), false},
		{"premature", func(m *SignedIntent) {}, now.Add(-1 * time.Minute), false},
		{"hash-changed-policy", func(m *SignedIntent) {}, now, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			copyMsg := msg
			tc.mutate(&copyMsg)
			local := p
			if tc.alterPolicy {
				local.Metadata.Revision++
			}
			result := gateFixture(t, key).Evaluate(copyMsg, local, tc.when)
			if result.Decision != "REJECTED" || result.KernelWrites {
				t.Fatalf("unsafe %s: %+v", tc.name, result)
			}
		})
	}
	otherSigned := msg
	otherSigned.Intent.Origin = "CLOUD_RELAY"
	if _, err := Sign(key, otherSigned.Intent); err == nil {
		t.Fatal("cloud origin accepted in P0")
	}
}
func TestSameRequestIDDifferentNonceAndPolicyHashDrift(t *testing.T) {
	key := bytes.Repeat([]byte{45}, 32)
	p := baseline(t)
	now := time.Date(2026, 10, 14, 13, 0, 0, 0, time.UTC)
	original := signedFixture(t, key, p, now)
	b := gateFixture(t, key)
	if r := b.Evaluate(original, p, now); r.Decision != "ADVISORY_ONLY" {
		t.Fatal(r)
	}
	alternate := original.Intent
	alternate.Nonce = "alternate-nonce-300"
	next, err := Sign(key, alternate)
	if err != nil {
		t.Fatal(err)
	}
	if r := b.Evaluate(next, p, now); r.Decision != "REJECTED" || !strings.Contains(r.Reason, "request ID") {
		t.Fatal(r)
	}
	altered := p
	altered.Spec.Autonomous.MaxBanTTLSeconds = 25 // same revision, different policy bytes
	if r := gateFixture(t, key).Evaluate(original, altered, now); r.Decision != "REJECTED" || !strings.Contains(r.Reason, "hash") {
		t.Fatal(r)
	}
}
func TestSignedApplyDenied(t *testing.T) {
	key := bytes.Repeat([]byte{43}, 32)
	p := baseline(t)
	now := time.Date(2026, 10, 14, 13, 0, 0, 0, time.UTC)
	msg := signedFixture(t, key, p, now)
	msg.Intent.Operation = "APPLY_POLICY"
	signed, err := Sign(key, msg.Intent)
	if err != nil {
		t.Fatal(err)
	}
	receipt := gateFixture(t, key).Evaluate(signed, p, now)
	if receipt.Decision != "REJECTED" || receipt.KernelWrites || receipt.ExecutorStatus != "NOT_AVAILABLE" {
		t.Fatal(receipt)
	}
}
func TestNoWeakKeysOrOverlongLease(t *testing.T) {
	p := baseline(t)
	now := time.Now().UTC()
	key := bytes.Repeat([]byte{1}, 32)
	msg := signedFixture(t, key, p, now)
	if _, err := NewBoundary([]byte("short"), "synthetic-site", "synthetic-agent"); err == nil {
		t.Fatal("weak key accepted")
	}
	msg.Intent.ExpiresUTC = msg.Intent.IssuedUTC.Add(MaxLifetime + time.Second)
	if _, err := Sign(key, msg.Intent); err == nil {
		t.Fatal("too long lifetime")
	}
}
func TestConcurrentNonceReplayOnlyOne(t *testing.T) {
	key := bytes.Repeat([]byte{44}, 32)
	p := baseline(t)
	now := time.Date(2026, 10, 14, 13, 0, 0, 0, time.UTC)
	msg := signedFixture(t, key, p, now)
	gate := gateFixture(t, key)
	const parallel = 32
	results := make(chan string, parallel)
	var wg sync.WaitGroup
	for j := 0; j < parallel; j++ {
		wg.Add(1)
		go func() { defer wg.Done(); results <- gate.Evaluate(msg, p, now).Decision }()
	}
	wg.Wait()
	close(results)
	accepted := 0
	for x := range results {
		if x == "ADVISORY_ONLY" {
			accepted++
		} else if x != "REJECTED" {
			t.Fatalf("unexpected decision %s", x)
		}
	}
	if accepted != 1 {
		t.Fatalf("replayed nonce accepted %d times", accepted)
	}
}
func TestSelftestSmoke(t *testing.T) {
	p := baseline(t)
	out, err := SelfTest(p, time.Date(2026, 10, 14, 13, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if out.Preview.Decision != "ADVISORY_ONLY" || out.Replay.Decision != "REJECTED" || out.Tampered.Decision != "REJECTED" || out.ApplyAttempt.Decision != "REJECTED" {
		t.Fatal(out)
	}
}
