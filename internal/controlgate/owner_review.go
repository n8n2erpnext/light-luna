package controlgate

// P0 offline cryptographic separation experiment. The reviewer public key
// here is a TEST key, not proof that a human owner authenticated or consented.
// This module is not wired to Gate, IPC, policy apply, or a privileged executor.
import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"sync"
	"time"
)

const OfflineReviewProtocol = "lightluna.offline-review/v1alpha1"
const OfflineReviewPurpose = "TEST_REVIEW_EVIDENCE_NOT_OWNER_APPROVAL"
const MaxOfflineReviewLifetime = 2 * time.Minute
const offlineReviewDomain = "lightluna.offline-review.ed25519/v1alpha1\x00"
const agentIntentDigestDomain = "lightluna.offline-review.agent-intent/v1alpha1\x00"
const MaxOfflineReviewSeen = 1024

// ReviewEvidence is scoped to one exact, separately Agent-signed intent.
// Neither this type nor a verified signature represents real owner presence.
type ReviewEvidence struct {
	Protocol          string    `json:"protocol"`
	Purpose           string    `json:"purpose"`
	ReviewID          string    `json:"reviewId"`
	ReviewerID        string    `json:"reviewerId"`
	SiteID            string    `json:"siteId"`
	RequestID         string    `json:"requestId"`
	AgentIntentSHA256 string    `json:"agentIntentSha256"`
	Operation         string    `json:"operation"`
	IssuedUTC         time.Time `json:"issuedUtc"`
	ExpiresUTC        time.Time `json:"expiresUtc"`
}
type SignedReviewEvidence struct {
	Evidence         ReviewEvidence `json:"evidence"`
	SignatureEd25519 string         `json:"signatureEd25519"`
}
type ReviewAssessment struct {
	Schema                  string `json:"schema"`
	Status                  string `json:"status"`
	Reason                  string `json:"reason"`
	CryptographicMatch      bool   `json:"cryptographicMatch"`
	HumanOwnerAuthenticated bool   `json:"humanOwnerAuthenticated"`
	GrantsApply             bool   `json:"grantsApply"`
	KernelWrites            bool   `json:"kernelWrites"`
	DurableReplay           bool   `json:"durableReplay"`
}

// OfflineReviewVerifier pins two DIFFERENT keys: Agent signer and an
// experimental reviewer signer. The in-memory anti-replay is deliberately
// not sufficient to authorize a privileged operation after process restart.
type OfflineReviewVerifier struct {
	agentPub                    ed25519.PublicKey
	reviewerPub                 ed25519.PublicKey
	siteID, actorID, reviewerID string
	mu                          sync.Mutex
	seen                        map[string]time.Time
	seenRequests                map[string]time.Time
	journal                     replayWriter
}

func NewOfflineReviewVerifier(agentPub, reviewerPub ed25519.PublicKey, siteID, actorID, reviewerID string) (*OfflineReviewVerifier, error) {
	if len(agentPub) != ed25519.PublicKeySize || len(reviewerPub) != ed25519.PublicKeySize {
		return nil, errors.New("review needs two pinned Ed25519 public keys")
	}
	if string(agentPub) == string(reviewerPub) {
		return nil, errors.New("agent and reviewer cannot share a signing identity")
	}
	if !safeID.MatchString(siteID) || !safeID.MatchString(actorID) || !safeID.MatchString(reviewerID) {
		return nil, errors.New("invalid pinned review identity")
	}
	return &OfflineReviewVerifier{
		agentPub:    append(ed25519.PublicKey(nil), agentPub...),
		reviewerPub: append(ed25519.PublicKey(nil), reviewerPub...),
		siteID:      siteID, actorID: actorID, reviewerID: reviewerID,
		seen:         make(map[string]time.Time),
		seenRequests: make(map[string]time.Time),
	}, nil
}
func validReview(e ReviewEvidence) error {
	if e.Protocol != OfflineReviewProtocol || e.Purpose != OfflineReviewPurpose ||
		!safeID.MatchString(e.ReviewID) || !safeID.MatchString(e.ReviewerID) ||
		!safeID.MatchString(e.SiteID) || !safeID.MatchString(e.RequestID) {
		return errors.New("invalid or misleading offline review identity")
	}
	if e.Operation != "PREVIEW_POLICY" && e.Operation != "APPLY_POLICY" {
		return errors.New("unknown reviewed operation")
	}
	d, err := hex.DecodeString(e.AgentIntentSHA256)
	if err != nil || len(d) != sha256.Size || hex.EncodeToString(d) != e.AgentIntentSHA256 {
		return errors.New("invalid canonical agent intent digest")
	}
	if e.IssuedUTC.IsZero() || e.ExpiresUTC.IsZero() ||
		!e.ExpiresUTC.After(e.IssuedUTC) || e.ExpiresUTC.Sub(e.IssuedUTC) > MaxOfflineReviewLifetime {
		return errors.New("invalid review TTL")
	}
	return nil
}
func reviewCanonical(e ReviewEvidence) ([]byte, error) {
	if err := validReview(e); err != nil {
		return nil, err
	}
	return json.Marshal(e)
}
func IntentReviewDigest(i Intent) (string, error) {
	payload, err := canonical(i)
	if err != nil {
		return "", err
	}
	data := append([]byte(agentIntentDigestDomain), payload...)
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}
func deniedReview(reason string) ReviewAssessment {
	return ReviewAssessment{Schema: "lightluna.review-assessment/v1alpha1",
		Status: "REJECTED", Reason: reason}
}
func (v *OfflineReviewVerifier) Check(agent SignedIntent, review SignedReviewEvidence, now time.Time) ReviewAssessment {
	if v == nil {
		return deniedReview("review verifier unavailable")
	}
	i := agent.Intent
	if err := i.Validate(); err != nil {
		return deniedReview("invalid agent intent")
	}
	if i.SiteID != v.siteID || i.ActorID != v.actorID || i.Origin != "LOCAL_AGENT" {
		return deniedReview("agent site or actor mismatch")
	}
	// The Agent signature is verified here against a separately pinned key;
	// knowing the review key is not sufficient to spoof an Agent.
	if agent.MACSHA256 != "" || len(agent.SignatureEd25519) != 128 {
		return deniedReview("only Agent Ed25519 envelope is accepted")
	}
	agentSig, err := hex.DecodeString(agent.SignatureEd25519)
	if err != nil || len(agentSig) != ed25519.SignatureSize ||
		hex.EncodeToString(agentSig) != agent.SignatureEd25519 {
		return deniedReview("noncanonical Agent signature")
	}
	canonicalIntent, err := canonical(i)
	if err != nil {
		return deniedReview("invalid Agent canonical payload")
	}
	if !ed25519.Verify(v.agentPub, append([]byte("lightluna.local-control.ed25519/v1alpha1\x00"), canonicalIntent...), agentSig) {
		return deniedReview("Agent signature mismatch")
	}
	if now.Before(i.IssuedUTC.Add(-30*time.Second)) || !now.Before(i.ExpiresUTC) {
		return deniedReview("Agent request outside clock window")
	}
	e := review.Evidence
	b, err := reviewCanonical(e)
	if err != nil {
		return deniedReview("invalid review evidence")
	}
	if e.SiteID != v.siteID || e.ReviewerID != v.reviewerID || e.RequestID != i.RequestID ||
		e.Operation != i.Operation {
		return deniedReview("review not scoped to pinned Agent request")
	}
	digest, err := IntentReviewDigest(i)
	if err != nil || e.AgentIntentSHA256 != digest {
		return deniedReview("review bound to different Agent intent")
	}
	if e.IssuedUTC.Before(i.IssuedUTC) || e.ExpiresUTC.After(i.ExpiresUTC) ||
		now.Before(e.IssuedUTC) || !now.Before(e.ExpiresUTC) {
		return deniedReview("review time scope invalid")
	}
	if len(review.SignatureEd25519) != ed25519.SignatureSize*2 {
		return deniedReview("invalid reviewer signature encoding")
	}
	sig, err := hex.DecodeString(review.SignatureEd25519)
	if err != nil || len(sig) != ed25519.SignatureSize || hex.EncodeToString(sig) != review.SignatureEd25519 {
		return deniedReview("noncanonical reviewer signature")
	}
	if !ed25519.Verify(v.reviewerPub, append([]byte(offlineReviewDomain), b...), sig) {
		return deniedReview("untrusted reviewer signature")
	}
	v.mu.Lock()
	defer v.mu.Unlock()
	for id, expires := range v.seen {
		if !now.Before(expires) {
			delete(v.seen, id)
		}
	}
	for id, expires := range v.seenRequests {
		if !now.Before(expires) {
			delete(v.seenRequests, id)
		}
	}
	if _, exists := v.seen[e.ReviewID]; exists {
		return deniedReview("review ID replayed")
	}
	if _, exists := v.seenRequests[digest]; exists {
		return deniedReview("Agent intent already reviewed; replay denied")
	}
	if len(v.seen) >= MaxOfflineReviewSeen || len(v.seenRequests) >= MaxOfflineReviewSeen {
		return deniedReview("review memory replay budget exhausted")
	}
	// For the optional offline journal path, persist BOTH replay dimensions
	// under the verifier mutex before reporting a cryptographic test match.
	// A failed/uncertain append never downgrades to memory-only acceptance.
	if v.journal != nil {
		if err := v.journal.record(e.ReviewID, digest, e.ExpiresUTC); err != nil {
			return deniedReview("offline review replay journal unavailable")
		}
	}
	v.seen[e.ReviewID] = e.ExpiresUTC
	v.seenRequests[digest] = e.ExpiresUTC
	return ReviewAssessment{
		Schema: "lightluna.review-assessment/v1alpha1",
		Status: "CRYPTOGRAPHIC_TEST_MATCH_ONLY", Reason: "two independent test signatures matched; not human consent or executor permission",
		CryptographicMatch: true,
	}
}
