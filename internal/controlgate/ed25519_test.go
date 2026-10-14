package controlgate

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"strings"
	"testing"
	"time"
)

func TestEd25519SignatureIsolationAndDowngrade(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	p := baseline(t)
	now := time.Date(2026, 10, 14, 13, 0, 0, 0, time.UTC)
	legacyKey := bytes.Repeat([]byte{49}, 32)
	base := signedFixture(t, legacyKey, p, now).Intent
	signed, err := SignEd25519(priv, base)
	if err != nil {
		t.Fatal(err)
	}
	b, err := NewEd25519Boundary(pub, "synthetic-site", "synthetic-agent")
	if err != nil {
		t.Fatal(err)
	}
	got := b.Evaluate(signed, p, now)
	if got.Decision != "ADVISORY_ONLY" || got.KernelWrites {
		t.Fatal(got)
	}
	if r := b.Evaluate(signed, p, now); r.Decision != "REJECTED" {
		t.Fatal(r)
	}
	wrongMAC, err := Sign(legacyKey, base)
	if err != nil {
		t.Fatal(err)
	}
	if r := gateFromPub(t, pub).Evaluate(wrongMAC, p, now); r.Decision != "REJECTED" {
		t.Fatal("accepted HMAC downgrade", r)
	}
	mixed := signed
	mixed.MACSHA256 = wrongMAC.MACSHA256
	if r := gateFromPub(t, pub).Evaluate(mixed, p, now); r.Decision != "REJECTED" {
		t.Fatal("accepted mixed signature", r)
	}
	tampered := signed
	tampered.Intent.ExpectedRevision++
	if r := gateFromPub(t, pub).Evaluate(tampered, p, now); r.Decision != "REJECTED" {
		t.Fatal("accepted signature tamper", r)
	}
	otherPub, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	if r := gateFromPub(t, otherPub).Evaluate(signed, p, now); r.Decision != "REJECTED" {
		t.Fatal("accepted wrong public key", r)
	}
	signedApply := signed
	signedApply.Intent.RequestID = "synthetic-ed-apply-100"
	signedApply.Intent.Nonce = "synthetic-ed-nonce-100"
	signedApply.Intent.Operation = "APPLY_POLICY"
	signedApply, err = SignEd25519(priv, signedApply.Intent)
	if err != nil {
		t.Fatal(err)
	}
	if r := gateFromPub(t, pub).Evaluate(signedApply, p, now); r.Decision != "REJECTED" || !strings.Contains(r.Reason, "no privileged executor") {
		t.Fatal(r)
	}
	legacy, err := NewBoundary(legacyKey, "synthetic-site", "synthetic-agent")
	if err != nil {
		t.Fatal(err)
	}
	if r := legacy.Evaluate(signed, p, now); r.Decision != "REJECTED" {
		t.Fatal("legacy HMAC Gate accepted Ed signature", r)
	}
}
func gateFromPub(t *testing.T, pub ed25519.PublicKey) *Boundary {
	t.Helper()
	b, e := NewEd25519Boundary(pub, "synthetic-site", "synthetic-agent")
	if e != nil {
		t.Fatal(e)
	}
	return b
}
