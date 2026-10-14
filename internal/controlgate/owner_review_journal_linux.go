//go:build linux

package controlgate

// Optional P0 TEST-ONLY restart-persistent replay backing for detached
// cryptographic review checks. This is not an owner-approval ledger,
// production authorization database, or executor receipt store.
import (
	"crypto/ed25519"
	"errors"
	"path/filepath"
	"strings"
	"time"
)

// NewOfflineReviewVerifierWithJournal uses a dedicated, separately named,
// nonroot, private journal; it never touches the Gate's replay.jsonl or
// replay-ed25519.jsonl. It reuses the tested inode/mode/flock/fsync/corruption
// handling of OpenReplayJournal. Caller MUST close the returned journal.
// No ReviewAssessment returned by this verifier grants a firewall operation.
func NewOfflineReviewVerifierWithJournal(
	agentPublic, reviewerPublic ed25519.PublicKey,
	siteID, actorID, reviewerID, path string, now time.Time,
) (*OfflineReviewVerifier, *ReplayJournal, error) {
	v, err := NewOfflineReviewVerifier(agentPublic, reviewerPublic, siteID, actorID, reviewerID)
	if err != nil {
		return nil, nil, err
	}
	// Refuse accidental reuse of the existing Agent/Gate replay files.
	if !filepath.IsAbs(path) || filepath.Base(path) != "offline-review.jsonl" {
		return nil, nil, errors.New("test-only review journal requires its dedicated absolute filename")
	}
	j, seen, err := OpenReplayJournal(path, now)
	if err != nil {
		return nil, nil, err
	}
	ready := false
	defer func() {
		if !ready {
			j.Close()
		}
	}()
	for key, expires := range seen {
		switch {
		case strings.HasPrefix(key, "request:"):
			v.seen[strings.TrimPrefix(key, "request:")] = expires
		case strings.HasPrefix(key, "nonce:"):
			v.seenRequests[strings.TrimPrefix(key, "nonce:")] = expires
		default:
			return nil, nil, errors.New("unexpected review journal replay dimension")
		}
	}
	if len(v.seen) != len(v.seenRequests) ||
		len(v.seen) > MaxOfflineReviewSeen ||
		len(v.seenRequests) > MaxOfflineReviewSeen {
		return nil, nil, errors.New("offline review journal replay capacity or shape invalid")
	}
	v.journal = j
	ready = true
	return v, j, nil
}
