//go:build linux

package controlgate

import (
	"crypto/ed25519"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func openOfflineReviewTestJournal(t *testing.T, ap, rp ed25519.PublicKey, path string, now time.Time) (*OfflineReviewVerifier, *ReplayJournal) {
	t.Helper()
	v, j, err := NewOfflineReviewVerifierWithJournal(ap, rp, "synthetic-site", "synthetic-agent", "offline-owner-test", path, now)
	if err != nil {
		t.Fatal(err)
	}
	return v, j
}

func newReviewForAgent(t *testing.T, previous SignedReviewEvidence, agent SignedIntent, reviewerPrivate ed25519.PrivateKey, reviewID string) SignedReviewEvidence {
	t.Helper()
	digest, err := IntentReviewDigest(agent.Intent)
	if err != nil {
		t.Fatal(err)
	}
	e := previous.Evidence
	e.ReviewID = reviewID
	e.RequestID = agent.Intent.RequestID
	e.AgentIntentSHA256 = digest
	return signReview(t, reviewerPrivate, e)
}

func TestOfflineReviewJournalRestartAndSeparateDimensions(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("journal requires a nonroot service")
	}
	now := time.Date(2026, 10, 14, 9, 0, 0, 0, time.UTC)
	ap, ak, rp, rk := reviewKeys(t)
	agent, review := offlineFixtures(t, ak, rk, now)
	path := filepath.Join(t.TempDir(), "offline-review.jsonl")
	v, j := openOfflineReviewTestJournal(t, ap, rp, path, now)
	accepted := v.Check(agent, review, now)
	if accepted.Status != "CRYPTOGRAPHIC_TEST_MATCH_ONLY" || !accepted.CryptographicMatch ||
		accepted.GrantsApply || accepted.KernelWrites || accepted.HumanOwnerAuthenticated || accepted.DurableReplay {
		t.Fatalf("unexpected authority %+v", accepted)
	}
	info, err := os.Stat(path)
	if err != nil || info.Mode().Perm() != 0600 || !info.Mode().IsRegular() {
		t.Fatalf("bad private journal: %v %v", info, err)
	}
	if err := j.Close(); err != nil {
		t.Fatal(err)
	}
	reloaded, j2 := openOfflineReviewTestJournal(t, ap, rp, path, now.Add(time.Second))
	if res := reloaded.Check(agent, review, now.Add(time.Second)); res.Status != "REJECTED" ||
		!strings.Contains(res.Reason, "replay") {
		t.Fatalf("accepted same review after restart: %+v", res)
	}
	anotherReview := newReviewForAgent(t, review, agent, rk, "offline-review-renamed-200")
	if res := reloaded.Check(agent, anotherReview, now.Add(time.Second)); res.Status != "REJECTED" ||
		!strings.Contains(res.Reason, "replay") {
		t.Fatalf("accepted same Agent intent under changed reviewer ID: %+v", res)
	}
	// A distinct Agent intent with a new request ID and nonce is independent.
	different := agent
	different.Intent.RequestID = "synthetic-agent-second-request"
	different.Intent.Nonce = "synthetic-agent-second-nonce"
	different, err = SignEd25519(ak, different.Intent)
	if err != nil {
		t.Fatal(err)
	}
	distinctReview := newReviewForAgent(t, review, different, rk, "offline-review-new-request-300")
	if res := reloaded.Check(different, distinctReview, now.Add(time.Second)); res.Status != "CRYPTOGRAPHIC_TEST_MATCH_ONLY" ||
		res.GrantsApply || res.DurableReplay {
		t.Fatalf("distinct test-only evidence unexpectedly denied/escalated %+v", res)
	}
	if err := j2.Close(); err != nil {
		t.Fatal(err)
	}
	third, j3 := openOfflineReviewTestJournal(t, ap, rp, path, now.Add(2*time.Second))
	defer j3.Close()
	if res := third.Check(different, distinctReview, now.Add(2*time.Second)); res.Status != "REJECTED" {
		t.Fatalf("repeated second evidence passed third startup %+v", res)
	}
	// Control Gate is still deny-only even when the detached test journal exists.
	actual := gateFromPub(t, ap).Evaluate(agent, baseline(t), now)
	if actual.Decision != "REJECTED" || actual.KernelWrites {
		t.Fatalf("offline journal must not authorize Gate %+v", actual)
	}
}

func TestOfflineReviewJournalExclusiveLockCorruptOrWrongPath(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("journal requires a nonroot service")
	}
	now := time.Date(2026, 10, 14, 9, 0, 0, 0, time.UTC)
	ap, _, rp, _ := reviewKeys(t)
	dir := t.TempDir()
	path := filepath.Join(dir, "offline-review.jsonl")
	if _, _, err := NewOfflineReviewVerifierWithJournal(ap, rp, "synthetic-site", "synthetic-agent", "offline-owner-test", filepath.Join(dir, "replay-ed25519.jsonl"), now); err == nil {
		t.Fatal("accepted production/Agent Gate journal path")
	}
	v, journal := openOfflineReviewTestJournal(t, ap, rp, path, now)
	if v == nil {
		t.Fatal("missing offline verifier")
	}
	if _, _, err := NewOfflineReviewVerifierWithJournal(ap, rp, "synthetic-site", "synthetic-agent", "offline-owner-test", path, now); err == nil {
		t.Fatal("second verifier acquired exclusive review ledger lock")
	}
	if err := journal.Close(); err != nil {
		t.Fatal(err)
	}
	// Strict journal decoder refuses malformed tail; no fallback to RAM.
	if err := os.WriteFile(path, []byte("{invalid-journal-record}\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := NewOfflineReviewVerifierWithJournal(ap, rp, "synthetic-site", "synthetic-agent", "offline-owner-test", path, now); err == nil {
		t.Fatal("corrupt journal silently accepted")
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	// Journal cannot be a symlink or existing non-file object.
	other := filepath.Join(dir, "other")
	if err := os.WriteFile(other, nil, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(other, path); err != nil {
		t.Fatal(err)
	}
	if _, _, err := NewOfflineReviewVerifierWithJournal(ap, rp, "synthetic-site", "synthetic-agent", "offline-owner-test", path, now); err == nil {
		t.Fatal("symlink journal accepted")
	}
}

func TestOfflineReviewJournalWriteErrorFailsClosed(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("journal requires a nonroot service")
	}
	now := time.Date(2026, 10, 14, 9, 0, 0, 0, time.UTC)
	ap, ak, rp, rk := reviewKeys(t)
	agent, review := offlineFixtures(t, ak, rk, now)
	v, journal := openOfflineReviewTestJournal(t, ap, rp, filepath.Join(t.TempDir(), "offline-review.jsonl"), now)
	// Inject a bounded file-descriptor failure inside the test only.
	if err := journal.file.Close(); err != nil {
		t.Fatal(err)
	}
	res := v.Check(agent, review, now)
	if res.Status != "REJECTED" || res.CryptographicMatch ||
		res.HumanOwnerAuthenticated || res.GrantsApply || res.KernelWrites ||
		len(v.seen) != 0 || len(v.seenRequests) != 0 {
		t.Fatalf("failed journal append acknowledged as review %+v", res)
	}
	if again := v.Check(agent, review, now); again.Status != "REJECTED" {
		t.Fatalf("journal silently downgraded to RAM %+v", again)
	}
}

func TestOfflineReviewJournalConcurrentSameIntentOnlyOnce(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("journal requires a nonroot service")
	}
	now := time.Date(2026, 10, 14, 9, 0, 0, 0, time.UTC)
	ap, ak, rp, rk := reviewKeys(t)
	agent, review := offlineFixtures(t, ak, rk, now)
	v, journal := openOfflineReviewTestJournal(t, ap, rp, filepath.Join(t.TempDir(), "offline-review.jsonl"), now)
	defer journal.Close()
	const workers = 32
	samples := make([]SignedReviewEvidence, workers)
	for i := range samples {
		samples[i] = newReviewForAgent(t, review, agent, rk, fmt.Sprintf("offline-review-concurrent-%04d", i))
	}
	var accepted atomic.Int32
	var wg sync.WaitGroup
	for _, sample := range samples {
		wg.Add(1)
		go func(rv SignedReviewEvidence) {
			defer wg.Done()
			answer := v.Check(agent, rv, now)
			if answer.Status == "CRYPTOGRAPHIC_TEST_MATCH_ONLY" {
				accepted.Add(1)
			} else if answer.Status != "REJECTED" {
				t.Errorf("unexpected review verdict %+v", answer)
			}
		}(sample)
	}
	wg.Wait()
	if accepted.Load() != 1 {
		t.Fatalf("concurrent same Agent intent accepted %d times, wanted exactly one", accepted.Load())
	}
}
