//go:build linux

package identitykeys

import (
	"bytes"
	"crypto/ed25519"
	"os"
	"path/filepath"
	"testing"
)

func TestEd25519MaterialSeparateAndNoWeakPaths(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("test only nonroot")
	}
	parent := t.TempDir()
	agent := filepath.Join(parent, "agent")
	gate := filepath.Join(parent, "gate")
	if err := Init(agent, gate); err != nil {
		t.Fatal(err)
	}
	if err := Init(agent, gate); err == nil {
		t.Fatal("overwrote existing material")
	}
	priv, err := LoadPrivate(agent)
	if err != nil {
		t.Fatal(err)
	}
	pub, err := LoadPublic(gate)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(priv.Public().(ed25519.PublicKey), pub) {
		t.Fatal("public/private mismatch")
	}
	if _, err := os.Stat(filepath.Join(gate, PrivateName)); !os.IsNotExist(err) {
		t.Fatal("private key present in Gate dir")
	}
	if _, err := os.Stat(filepath.Join(agent, PublicName)); !os.IsNotExist(err) {
		t.Fatal("Gate public key unexpectedly in Agent dir")
	}
	p := filepath.Join(agent, PrivateName)
	if err := os.Chmod(p, 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadPrivate(agent); err == nil {
		t.Fatal("world-readable private key accepted")
	}
	os.Chmod(p, 0600)
	if err := os.Chmod(agent, 0750); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadPrivate(agent); err == nil {
		t.Fatal("non-private Agent dir accepted")
	}
	os.Chmod(agent, 0700)
	if err := os.Remove(p); err != nil {
		t.Fatal(err)
	}
	ext := filepath.Join(parent, "fake")
	if err := os.WriteFile(ext, priv, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(ext, p); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadPrivate(agent); err == nil {
		t.Fatal("symlinked key accepted")
	}
}
