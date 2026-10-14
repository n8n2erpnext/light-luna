package controlgate

import (
	"crypto/rand"
	"fmt"
	"time"

	"github.com/n8n2erpnext/light-luna/internal/policy"
)

// SelfTest is a *local one-shot developer demonstration* using a random key.
// It tests a future trust-boundary contract; NO remote listener or trusted
// host process is created. Real actor identity, IPC and key custody are absent.
type SelfTestReport struct {
	Schema       string  `json:"schema"`
	Mode         string  `json:"mode"`
	Preview      Receipt `json:"preview"`
	Replay       Receipt `json:"replay"`
	Tampered     Receipt `json:"tampered"`
	ApplyAttempt Receipt `json:"applyAttempt"`
}

func SelfTest(p policy.Bundle, now time.Time) (SelfTestReport, error) {
	if err := p.Validate(); err != nil {
		return SelfTestReport{}, err
	}
	summary, err := policy.MakePreview(p)
	if err != nil {
		return SelfTestReport{}, err
	}
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return SelfTestReport{}, fmt.Errorf("random dev key: %w", err)
	}
	defer func() {
		for j := range key {
			key[j] = 0
		}
	}()
	gate, err := NewBoundary(key, "synthetic-site", "synthetic-agent")
	if err != nil {
		return SelfTestReport{}, err
	}
	makeIntent := func(id, nonce, op string) Intent {
		return Intent{Protocol: Protocol, RequestID: id, Nonce: nonce, SiteID: "synthetic-site", ActorID: "synthetic-agent",
			ActorRole: "OWNER", Origin: "LOCAL_AGENT", PolicyName: p.Metadata.Name, ExpectedRevision: p.Metadata.Revision,
			PolicyHashSHA256: summary.HashSHA256, IssuedUTC: now.Add(-time.Second), ExpiresUTC: now.Add(time.Minute),
			Operation: op}
	}
	signed, err := Sign(key, makeIntent("synthetic-preview-001", "synthetic-nonce-preview-001", "PREVIEW_POLICY"))
	if err != nil {
		return SelfTestReport{}, err
	}
	first := gate.Evaluate(signed, p, now)
	duplicate := gate.Evaluate(signed, p, now)
	forged := signed
	forged.Intent.ExpectedRevision++ // original MAC must fail verification
	tampered := gate.Evaluate(forged, p, now)
	signedApply, err := Sign(key, makeIntent("synthetic-apply-001", "synthetic-nonce-apply-001", "APPLY_POLICY"))
	if err != nil {
		return SelfTestReport{}, err
	}
	apply := gate.Evaluate(signedApply, p, now)
	if first.Decision != "ADVISORY_ONLY" || duplicate.Decision != "REJECTED" || tampered.Decision != "REJECTED" ||
		apply.Decision != "REJECTED" || first.KernelWrites || apply.KernelWrites {
		return SelfTestReport{}, fmt.Errorf("selftest security invariant failed")
	}
	return SelfTestReport{Schema: "lightluna.control-selftest/v1alpha1", Mode: "EPHEMERAL_IN_PROCESS_NO_IPC_NO_EXECUTOR",
		Preview: first, Replay: duplicate, Tampered: tampered, ApplyAttempt: apply}, nil
}
