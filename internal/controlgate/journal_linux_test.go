//go:build linux

package controlgate

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestReplayJournalSurvivesGateRestart(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("P0 nonroot only")
	}
	dir := t.TempDir()
	os.Chmod(dir, 0700)
	file := filepath.Join(dir, "replay.jsonl")
	key := bytes.Repeat([]byte{85}, 32)
	now := time.Date(2026, 10, 14, 13, 0, 0, 0, time.UTC)
	p := baseline(t)
	msg := signedFixture(t, key, p, now)
	b, j, err := NewBoundaryWithJournal(key, "synthetic-site", "synthetic-agent", file, now)
	if err != nil {
		t.Fatal(err)
	}
	first := b.Evaluate(msg, p, now)
	if first.Decision != "ADVISORY_ONLY" || first.ReplayProtection != "FSYNC_JOURNAL_DEMO_ONLY" {
		t.Fatal(first)
	}
	if err := j.Close(); err != nil {
		t.Fatal(err)
	}
	b2, j2, err := NewBoundaryWithJournal(key, "synthetic-site", "synthetic-agent", file, now)
	if err != nil {
		t.Fatal(err)
	}
	defer j2.Close()
	replay := b2.Evaluate(msg, p, now)
	if replay.Decision != "REJECTED" || !strings.Contains(replay.Reason, "replay") {
		t.Fatal(replay)
	}
	// Changing nonce with identical request id also cannot bypass journal.
	alt := msg.Intent
	alt.Nonce = "synthetic-nonce-alternate-999"
	signed, err := Sign(key, alt)
	if err != nil {
		t.Fatal(err)
	}
	r := b2.Evaluate(signed, p, now)
	if r.Decision != "REJECTED" || !strings.Contains(r.Reason, "request ID") {
		t.Fatal(r)
	}
}
func TestReplayJournalSingleGateAndCorruptionFailClosed(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("nonroot only")
	}
	dir := t.TempDir()
	os.Chmod(dir, 0700)
	path := filepath.Join(dir, "replay.jsonl")
	now := time.Now().UTC()
	j, _, err := OpenReplayJournal(path, now)
	if err != nil {
		t.Fatal(err)
	}
	if other, _, err := OpenReplayJournal(path, now); err == nil {
		other.Close()
		t.Fatal("second writer acquired journal lock")
	}
	j.Close()
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString("{corrupt-content\n")
	f.Close()
	if j, _, err := OpenReplayJournal(path, now); err == nil {
		j.Close()
		t.Fatal("corrupt journal accepted")
	}
}
func TestJournalInsecurePermissionsAndSymlinkDenied(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("nonroot only")
	}
	dir := t.TempDir()
	os.Chmod(dir, 0700)
	path := filepath.Join(dir, "replay.jsonl")
	os.WriteFile(path, nil, 0600)
	os.Chmod(path, 0644) // override process umask: explicit insecure fixture
	if j, _, err := OpenReplayJournal(path, time.Now()); err == nil {
		j.Close()
		t.Fatal("insecure mode accepted")
	}
	os.Remove(path)
	outside := filepath.Join(dir, "actual")
	os.WriteFile(outside, nil, 0600)
	os.Symlink(outside, path)
	if j, _, err := OpenReplayJournal(path, time.Now()); err == nil {
		j.Close()
		t.Fatal("symlink accepted")
	}
}
func TestJournalWriteFailClosed(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("nonroot only")
	}
	dir := t.TempDir()
	os.Chmod(dir, 0700)
	file := filepath.Join(dir, "replay.jsonl")
	key := bytes.Repeat([]byte{86}, 32)
	now := time.Date(2026, 10, 14, 13, 0, 0, 0, time.UTC)
	p := baseline(t)
	b, j, err := NewBoundaryWithJournal(key, "synthetic-site", "synthetic-agent", file, now)
	if err != nil {
		t.Fatal(err)
	}
	defer j.Close()
	j.failed = true
	msg := signedFixture(t, key, p, now)
	out := b.Evaluate(msg, p, now)
	if out.Decision != "REJECTED" || out.KernelWrites || !strings.Contains(out.Reason, "journal") {
		t.Fatal(out)
	}
}
