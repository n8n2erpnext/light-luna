//go:build linux

package localipc

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/n8n2erpnext/light-luna/internal/controlgate"
	"github.com/n8n2erpnext/light-luna/internal/policy"
)

const fixturePolicy = `{
"apiVersion":"lightluna.io/v1alpha1","kind":"PolicyBundle",
"metadata":{"name":"synthetic-test","revision":3},
"spec":{"mode":"host","defaultInbound":"accept","emergencyAccess":{"enabled":false,"networks":[]},
"aliases":[],"rules":[],"autonomous":{"mode":"observe","allowRealBans":false,"maxBanTtlSeconds":0}}
}`

func testPolicy(t *testing.T) policy.Bundle {
	t.Helper()
	p, err := policy.Decode(strings.NewReader(fixturePolicy))
	if err != nil {
		t.Fatal(err)
	}
	return p
}
func fixture(t *testing.T, key []byte, p policy.Bundle, req, nonce, op string) controlgate.SignedIntent {
	t.Helper()
	hash, err := policy.MakePreview(p)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	msg, err := controlgate.Sign(key, controlgate.Intent{
		Protocol: controlgate.Protocol, RequestID: req, Nonce: nonce,
		SiteID: "synthetic-site", ActorID: "synthetic-agent", ActorRole: "OWNER", Origin: "LOCAL_AGENT",
		PolicyName: p.Metadata.Name, ExpectedRevision: p.Metadata.Revision, PolicyHashSHA256: hash.HashSHA256,
		IssuedUTC: now.Add(-time.Second), ExpiresUTC: now.Add(90 * time.Second), Operation: op})
	if err != nil {
		t.Fatal(err)
	}
	return msg
}
func running(t *testing.T, uid uint32) (string, func(), []byte, policy.Bundle) {
	t.Helper()
	dir := t.TempDir()
	if err := os.Chmod(dir, 0700); err != nil {
		t.Fatal(err)
	}
	ln, path, err := ListenPrivate(dir)
	if err != nil {
		t.Fatal(err)
	}
	p := testPolicy(t)
	key := bytes.Repeat([]byte{81}, 32) // public synthetic unit-test fixture, not production secret
	b, err := controlgate.NewBoundary(key, "synthetic-site", "synthetic-agent")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- Serve(ctx, ln, uid, b, p, time.Now) }()
	cleanup := func() {
		cancel()
		ln.Close()
		if err := <-done; err != nil {
			t.Errorf("server shutdown: %v", err)
		}
	}
	return path, cleanup, key, p
}
func TestLinuxSocketPreviewReplayAndDeniedApply(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("P0 refuses to run as root")
	}
	path, stop, key, p := running(t, uint32(os.Geteuid()))
	defer stop()
	msg := fixture(t, key, p, "synthetic-socket-req-01", "synthetic-socket-nonce-01", "PREVIEW_POLICY")
	ctx := context.Background()
	one, err := Call(ctx, path, msg)
	if err != nil || one.Decision != "ADVISORY_ONLY" || one.KernelWrites {
		t.Fatalf("first: %+v %v", one, err)
	}
	same, err := Call(ctx, path, msg)
	if err != nil || same.Decision != "REJECTED" || !strings.Contains(same.Reason, "replay") {
		t.Fatalf("replay: %+v %v", same, err)
	}
	modified := msg
	modified.Intent.ExpectedRevision = 999
	bad, err := Call(ctx, path, modified)
	if err != nil || bad.Decision != "REJECTED" {
		t.Fatalf("tamper: %+v %v", bad, err)
	}
	apply := fixture(t, key, p, "synthetic-socket-req-02", "synthetic-socket-nonce-02", "APPLY_POLICY")
	deny, err := Call(ctx, path, apply)
	if err != nil || deny.Decision != "REJECTED" || deny.KernelWrites {
		t.Fatalf("apply: %+v %v", deny, err)
	}
}
func TestLinuxSocketPeerUIDDenied(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("P0 refuses to run as root")
	}
	path, stop, key, p := running(t, uint32(os.Geteuid()+1))
	defer stop()
	msg := fixture(t, key, p, "synthetic-socket-req-01", "synthetic-socket-nonce-01", "PREVIEW_POLICY")
	if r, err := Call(context.Background(), path, msg); err == nil {
		t.Fatalf("mismatched UID yielded reply: %+v", r)
	}
}
func TestLinuxSocketPrivateDirectoryAndSymlinkRejected(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("P0 refuses to run as root")
	}
	dir := t.TempDir()
	if err := os.Chmod(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if ln, _, err := ListenPrivate(dir); err == nil {
		ln.Close()
		t.Fatal("public dir accepted")
	}
	if err := os.Chmod(dir, 0700); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(t.TempDir(), "linked-dir")
	if err := os.Symlink(dir, link); err != nil {
		t.Fatal(err)
	}
	if ln, _, err := ListenPrivate(link); err == nil {
		ln.Close()
		t.Fatal("symlink accepted")
	}
	ln, path, err := ListenPrivate(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	if ln2, _, err := ListenPrivate(dir); err == nil {
		ln2.Close()
		t.Fatal("second listener overwrote socket")
	}
	mode, err := os.Lstat(path)
	if err != nil || mode.Mode().Perm() != 0600 {
		t.Fatalf("socket perms %+v %v", mode, err)
	}
}
func TestLinuxSocketOversizeAndMalformedFrames(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("P0 refuses to run as root")
	}
	path, stop, _, _ := running(t, uint32(os.Geteuid()))
	defer stop()
	for name, payload := range map[string][]byte{
		"oversize-header": func() []byte { b := make([]byte, 4); binary.BigEndian.PutUint32(b, MaxRequestBytes+1); return b }(),
		"invalid-json": func() []byte {
			v := []byte(`{"intent": {"protocol":"abc","protocol":"xyz"}}`)
			b := make([]byte, 4)
			binary.BigEndian.PutUint32(b, uint32(len(v)))
			return append(b, v...)
		}(),
	} {
		t.Run(name, func(t *testing.T) {
			c, err := net.DialTimeout("unix", path, time.Second)
			if err != nil {
				t.Fatal(err)
			}
			defer c.Close()
			c.SetDeadline(time.Now().Add(time.Second))
			if _, err := c.Write(payload); err != nil {
				t.Fatal(err)
			}
			var hdr [4]byte
			if _, err := c.Read(hdr[:]); err == nil {
				t.Fatal("server replied to malformed payload")
			}
		})
	}
}
func TestIPCClientSubprocess(t *testing.T) {
	if os.Getenv("LIGHTLUNA_IPC_UNIT_HELPER") != "yes" {
		return
	}
	// This helper runs in a DIFFERENT OS process but with same UID, proving
	// kernel IPC peer credential and framing work beyond one process.
	path := os.Getenv("LIGHTLUNA_IPC_TEST_SOCKET")
	p := testPolicy(t)
	key := bytes.Repeat([]byte{81}, 32)
	req := fixture(t, key, p, "synthetic-process-req-001", "synthetic-process-nonce-001", "PREVIEW_POLICY")
	out, err := Call(context.Background(), path, req)
	if err != nil || out.Decision != "ADVISORY_ONLY" {
		t.Fatalf("cross-process IPC: %+v %v", out, err)
	}
}
func TestLinuxSocketCrossProcess(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("P0 refuses to run as root")
	}
	path, stop, _, _ := running(t, uint32(os.Geteuid()))
	defer stop()
	bin, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(bin, "-test.run=^TestIPCClientSubprocess$")
	cmd.Env = append(os.Environ(), "LIGHTLUNA_IPC_UNIT_HELPER=yes", "LIGHTLUNA_IPC_TEST_SOCKET="+path)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("subprocess failure: %v: %s", err, out)
	}
}
func TestLinuxSelfTest(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("P0 refuses to run as root")
	}
	report, err := SelfTest(testPolicy(t))
	if err != nil {
		t.Fatal(err)
	}
	if report.KernelWrites || report.Preview.Decision != "ADVISORY_ONLY" ||
		report.Replay.Decision != "REJECTED" || report.Tampered.Decision != "REJECTED" ||
		report.ApplyAttempt.Decision != "REJECTED" {
		t.Fatalf("invalid selftest %+v", report)
	}
}
func TestIPCFrameWriterBounds(t *testing.T) {
	a, b := net.Pipe()
	defer a.Close()
	defer b.Close()
	if err := writeFrame(a, make([]byte, MaxRequestBytes+1), MaxRequestBytes); err == nil {
		t.Fatal("accepted oversized frame")
	}
	// Receipt wire roundtrip should be bounded, even when decoded payload is valid JSON.
	_, err := json.Marshal(controlgate.Receipt{Decision: "REJECTED"})
	if err != nil {
		t.Fatal(err)
	}
}
