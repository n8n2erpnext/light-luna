// Package controlgate models the future Agent → Policy Gate authentication
// boundary. It is an in-process, non-privileged P0 contract, NOT a socket server,
// persistent replay store, credential manager or firewall executor.
package controlgate

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sync"
	"time"

	"github.com/n8n2erpnext/light-luna/internal/policy"
)

const Protocol = "lightluna.local-control/v1alpha1"
const MaxLifetime = 5 * time.Minute
const MaxReplayRecords = 2048

var safeID = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._:-]{7,127}$`)

type Intent struct {
	Protocol         string    `json:"protocol"`
	RequestID        string    `json:"requestId"`
	Nonce            string    `json:"nonce"`
	SiteID           string    `json:"siteId"`
	ActorID          string    `json:"actorId"`
	ActorRole        string    `json:"actorRole"`
	Origin           string    `json:"origin"`
	PolicyName       string    `json:"policyName"`
	ExpectedRevision int64     `json:"expectedRevision"`
	PolicyHashSHA256 string    `json:"policyHashSha256"`
	IssuedUTC        time.Time `json:"issuedUtc"`
	ExpiresUTC       time.Time `json:"expiresUtc"`
	Operation        string    `json:"operation"`
}
type SignedIntent struct {
	Intent    Intent `json:"intent"`
	MACSHA256 string `json:"macSha256"`
}
type Receipt struct {
	Schema           string `json:"schema"`
	RequestID        string `json:"requestId"`
	Decision         string `json:"decision"`
	Reason           string `json:"reason"`
	ExecutorStatus   string `json:"executorStatus"`
	KernelWrites     bool   `json:"kernelWrites"`
	ReplayProtection string `json:"replayProtection"`
}
type Boundary struct {
	key     []byte
	siteID  string
	actorID string
	mu      sync.Mutex
	seen    map[string]time.Time
	journal replayWriter
}

type replayWriter interface {
	record(id, nonce string, expires time.Time) error
}

func NewBoundary(ephemeralKey []byte, siteID, actorID string) (*Boundary, error) {
	if len(ephemeralKey) < 32 {
		return nil, errors.New("ephemeral shared key must contain at least 32 bytes")
	}
	if !safeID.MatchString(siteID) || !safeID.MatchString(actorID) {
		return nil, errors.New("invalid pinned local identity")
	}
	return &Boundary{key: append([]byte(nil), ephemeralKey...), siteID: siteID, actorID: actorID, seen: map[string]time.Time{}}, nil
}
func (i Intent) Validate() error {
	if i.Protocol != Protocol {
		return errors.New("unsupported control protocol")
	}
	if !safeID.MatchString(i.RequestID) || !safeID.MatchString(i.Nonce) ||
		!safeID.MatchString(i.SiteID) || !safeID.MatchString(i.ActorID) {
		return errors.New("invalid control identifier")
	}
	if i.ActorRole != "OWNER" || i.Origin != "LOCAL_AGENT" {
		return errors.New("unsupported identity or remote origin in P0")
	}
	if i.PolicyName == "" || len(i.PolicyName) > 64 || i.ExpectedRevision < 1 {
		return errors.New("policy scope/revision invalid")
	}
	if len(i.PolicyHashSHA256) != 64 {
		return errors.New("policy content digest required")
	}
	raw, err := hex.DecodeString(i.PolicyHashSHA256)
	if err != nil || len(raw) != sha256.Size || hex.EncodeToString(raw) != i.PolicyHashSHA256 {
		return errors.New("noncanonical policy SHA256")
	}
	if i.Operation != "PREVIEW_POLICY" && i.Operation != "APPLY_POLICY" {
		return errors.New("unsupported operation")
	}
	if i.IssuedUTC.IsZero() || i.ExpiresUTC.IsZero() || !i.ExpiresUTC.After(i.IssuedUTC) ||
		i.ExpiresUTC.Sub(i.IssuedUTC) > MaxLifetime {
		return errors.New("expired or oversized control lifetime")
	}
	return nil
}
func canonical(i Intent) ([]byte, error) {
	if err := i.Validate(); err != nil {
		return nil, err
	}
	return json.Marshal(i)
}

// Sign creates an HMAC envelope with an externally supplied key.
// The P0 demo uses a random per-invocation key; there is NO persistent key
// provisioning, remote transport or user access implied.
func Sign(key []byte, i Intent) (SignedIntent, error) {
	if len(key) < 32 {
		return SignedIntent{}, errors.New("key too short")
	}
	b, err := canonical(i)
	if err != nil {
		return SignedIntent{}, err
	}
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(Protocol + "\x00"))
	mac.Write(b)
	return SignedIntent{Intent: i, MACSHA256: hex.EncodeToString(mac.Sum(nil))}, nil
}
func rejected(id, reason string) Receipt {
	return Receipt{Schema: "lightluna.control-receipt/v1alpha1", RequestID: id, Decision: "REJECTED", Reason: reason,
		ExecutorStatus: "NOT_AVAILABLE", KernelWrites: false, ReplayProtection: "IN_MEMORY_ONLY"}
}

func (b *Boundary) deny(id, reason string) Receipt {
	r := rejected(id, reason)
	if b.journal != nil {
		r.ReplayProtection = "FSYNC_JOURNAL_DEMO_ONLY"
	}
	return r
}

// Evaluate is intentionally unable to invoke anything privileged. All requests
// for APPLY_POLICY are rejected even if signed, fresh and owner-scoped.
func (b *Boundary) Evaluate(message SignedIntent, p policy.Bundle, now time.Time) Receipt {
	i := message.Intent
	if err := i.Validate(); err != nil {
		return b.deny(i.RequestID, err.Error())
	}
	rawMAC, err := hex.DecodeString(message.MACSHA256)
	if err != nil || len(rawMAC) != sha256.Size {
		return b.deny(i.RequestID, "invalid authenticator")
	}
	canonicalBytes, err := canonical(i)
	if err != nil {
		return b.deny(i.RequestID, "invalid canonical intent")
	}
	mac := hmac.New(sha256.New, b.key)
	mac.Write([]byte(Protocol + "\x00"))
	mac.Write(canonicalBytes)
	if !hmac.Equal(rawMAC, mac.Sum(nil)) {
		return b.deny(i.RequestID, "invalid authenticator")
	}
	if i.SiteID != b.siteID || i.ActorID != b.actorID {
		return b.deny(i.RequestID, "actor/site boundary mismatch")
	}
	if now.Before(i.IssuedUTC.Add(-30*time.Second)) || !now.Before(i.ExpiresUTC) {
		return b.deny(i.RequestID, "stale or not-yet-valid control intent")
	}
	if err := p.Validate(); err != nil {
		return b.deny(i.RequestID, "invalid local policy")
	}
	if i.PolicyName != p.Metadata.Name || i.ExpectedRevision != p.Metadata.Revision {
		return b.deny(i.RequestID, "local policy revision mismatch")
	}
	preview, err := policy.MakePreview(p)
	if err != nil {
		return b.deny(i.RequestID, "policy content validation failed")
	}
	if i.PolicyHashSHA256 != preview.HashSHA256 {
		return b.deny(i.RequestID, "local policy content hash mismatch")
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	// Old replays are forgotten only after their expiry. A corrupt/full store
	// denies new requests instead of silently discarding replay protection.
	for nonce, expires := range b.seen {
		if !now.Before(expires) {
			delete(b.seen, nonce)
		}
	}
	if _, exists := b.seen["nonce:"+i.Nonce]; exists {
		return b.deny(i.RequestID, "replayed control nonce")
	}
	if _, exists := b.seen["request:"+i.RequestID]; exists {
		return b.deny(i.RequestID, "replayed control request ID")
	}
	if len(b.seen)+2 > MaxReplayRecords {
		return b.deny(i.RequestID, "replay budget exhausted; fail closed")
	}
	if b.journal != nil {
		if err := b.journal.record(i.RequestID, i.Nonce, i.ExpiresUTC); err != nil {
			return b.deny(i.RequestID, "durable replay journal unavailable; fail closed")
		}
	}
	b.seen["nonce:"+i.Nonce] = i.ExpiresUTC
	b.seen["request:"+i.RequestID] = i.ExpiresUTC
	if i.Operation == "APPLY_POLICY" {
		return b.deny(i.RequestID, "P0 has no privileged executor, lease or independently verified rollback")
	}
	return Receipt{Schema: "lightluna.control-receipt/v1alpha1", RequestID: i.RequestID,
		Decision: "ADVISORY_ONLY", Reason: fmt.Sprintf("Authorized synthetic preview for policy revision %d; no real execution", p.Metadata.Revision),
		ExecutorStatus: "NOT_AVAILABLE", KernelWrites: false, ReplayProtection: b.replayMode()}
}
func (b *Boundary) replayMode() string {
	if b.journal != nil {
		return "FSYNC_JOURNAL_DEMO_ONLY"
	}
	return "IN_MEMORY_ONLY"
}
