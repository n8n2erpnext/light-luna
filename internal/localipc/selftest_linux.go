//go:build linux

package localipc

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/n8n2erpnext/light-luna/internal/controlgate"
	"github.com/n8n2erpnext/light-luna/internal/policy"
)

// SelfTest demonstrates a real kernel Unix socket roundtrip, but both endpoints
// are goroutines of ONE unprivileged process. This is not a multi-service login.
type SelfTestReport struct {
	Schema          string              `json:"schema"`
	Mode            string              `json:"mode"`
	PeerCredentials string              `json:"peerCredentials"`
	Preview         controlgate.Receipt `json:"preview"`
	Replay          controlgate.Receipt `json:"replay"`
	Tampered        controlgate.Receipt `json:"tampered"`
	ApplyAttempt    controlgate.Receipt `json:"applyAttempt"`
	KernelWrites    bool                `json:"kernelWrites"`
}

func SelfTest(p policy.Bundle) (SelfTestReport, error) {
	if os.Geteuid() == 0 {
		return SelfTestReport{}, errors.New("P0 socket selftest must be nonroot")
	}
	if err := p.Validate(); err != nil {
		return SelfTestReport{}, err
	}
	pv, err := policy.MakePreview(p)
	if err != nil {
		return SelfTestReport{}, err
	}
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return SelfTestReport{}, err
	}
	defer func() {
		for i := range key {
			key[i] = 0
		}
	}()
	boundary, err := controlgate.NewBoundary(key, "synthetic-site", "synthetic-agent")
	if err != nil {
		return SelfTestReport{}, err
	}
	dir, err := os.MkdirTemp("", "luna-ipc-")
	if err != nil {
		return SelfTestReport{}, err
	}
	defer os.RemoveAll(dir)
	ln, path, err := ListenPrivate(dir)
	if err != nil {
		return SelfTestReport{}, err
	}
	defer ln.Close()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- Serve(ctx, ln, uint32(os.Geteuid()), boundary, p, time.Now) }()
	defer func() { cancel(); ln.Close(); <-done }()
	makeIntent := func(id, nonce, op string) controlgate.Intent {
		now := time.Now().UTC()
		return controlgate.Intent{Protocol: controlgate.Protocol, RequestID: id, Nonce: nonce,
			SiteID: "synthetic-site", ActorID: "synthetic-agent", ActorRole: "OWNER", Origin: "LOCAL_AGENT",
			PolicyName: p.Metadata.Name, ExpectedRevision: p.Metadata.Revision, PolicyHashSHA256: pv.HashSHA256,
			IssuedUTC: now.Add(-time.Second), ExpiresUTC: now.Add(time.Minute), Operation: op}
	}
	signed, err := controlgate.Sign(key, makeIntent("synthetic-ipc-preview-001", "synthetic-ipc-nonce-001", "PREVIEW_POLICY"))
	if err != nil {
		return SelfTestReport{}, err
	}
	preview, err := Call(ctx, path, signed)
	if err != nil {
		return SelfTestReport{}, err
	}
	replay, err := Call(ctx, path, signed)
	if err != nil {
		return SelfTestReport{}, err
	}
	tampered := signed
	tampered.Intent.ExpectedRevision++
	// Tampered content cannot be signed by caller without a trusted key; raw
	// network frame is used here because Call also validates envelope structure.
	invalid, err := Call(ctx, path, tampered)
	if err != nil {
		return SelfTestReport{}, err
	}
	apply, err := controlgate.Sign(key, makeIntent("synthetic-ipc-apply-001", "synthetic-ipc-apply-nonce-001", "APPLY_POLICY"))
	if err != nil {
		return SelfTestReport{}, err
	}
	denied, err := Call(ctx, path, apply)
	if err != nil {
		return SelfTestReport{}, err
	}
	if preview.Decision != "ADVISORY_ONLY" || replay.Decision != "REJECTED" ||
		invalid.Decision != "REJECTED" || denied.Decision != "REJECTED" ||
		preview.KernelWrites || replay.KernelWrites || invalid.KernelWrites || denied.KernelWrites {
		return SelfTestReport{}, fmt.Errorf("P0 IPC invariant failed")
	}
	return SelfTestReport{Schema: "lightluna.ipc-selftest/v1alpha1",
		Mode:            "SAME_PROCESS_GOROUTINES_REAL_LINUX_UNIX_SOCKET_NO_EXECUTOR",
		PeerCredentials: "SO_PEERCRED_UID_MATCH_ONLY",
		Preview:         preview, Replay: replay, Tampered: invalid, ApplyAttempt: denied, KernelWrites: false}, nil
}
