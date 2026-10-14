//go:build linux

// Package localipc exposes a bounded, nonprivileged P0 Unix socket roundtrip.
// SO_PEERCRED establishes the Linux peer UID, not the human owner identity.
// All operations remain advisory/denied through controlgate.Boundary.
package localipc

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/n8n2erpnext/light-luna/internal/controlgate"
	"github.com/n8n2erpnext/light-luna/internal/policy"
	"github.com/n8n2erpnext/light-luna/internal/strictjson"
)

const MaxRequestBytes = 16 << 10
const MaxResponseBytes = 4 << 10
const SocketTimeout = 3 * time.Second
const MaxClients = 8

// ListenPrivate creates a socket *only* inside an existing, private 0700
// directory owned by this nonroot user. Never use a shared /tmp socket root.
func ListenPrivate(dir string) (*net.UnixListener, string, error) {
	if os.Geteuid() == 0 {
		return nil, "", errors.New("P0 IPC cannot run as root")
	}
	info, err := os.Lstat(dir)
	if err != nil {
		return nil, "", err
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok || !info.IsDir() || info.Mode().Perm() != 0700 || stat.Uid != uint32(os.Geteuid()) {
		return nil, "", errors.New("IPC directory must be owned by current UID and mode 0700; symlinks forbidden")
	}
	path := filepath.Join(dir, "gate.sock")
	if len(path) > 103 {
		return nil, "", errors.New("Unix socket path too long")
	}
	if _, err := os.Lstat(path); err == nil {
		return nil, "", errors.New("refuse to reuse pre-existing socket path")
	} else if !os.IsNotExist(err) {
		return nil, "", err
	}
	ln, err := net.ListenUnix("unix", &net.UnixAddr{Name: path, Net: "unix"})
	if err != nil {
		return nil, "", err
	}
	ln.SetUnlinkOnClose(true)
	if err := os.Chmod(path, 0600); err != nil {
		ln.Close()
		return nil, "", err
	}
	return ln, path, nil
}

func peerUID(conn *net.UnixConn) (uint32, error) {
	raw, err := conn.SyscallConn()
	if err != nil {
		return 0, err
	}
	var uid uint32
	var inner error
	err = raw.Control(func(fd uintptr) {
		var cred *syscall.Ucred
		cred, inner = syscall.GetsockoptUcred(int(fd), syscall.SOL_SOCKET, syscall.SO_PEERCRED)
		if inner == nil {
			uid = cred.Uid
		}
	})
	if err != nil {
		return 0, err
	}
	return uid, inner
}
func readFrame(c net.Conn, max uint32) ([]byte, error) {
	var length [4]byte
	if _, err := io.ReadFull(c, length[:]); err != nil {
		return nil, fmt.Errorf("frame header: %w", err)
	}
	size := binary.BigEndian.Uint32(length[:])
	if size == 0 || size > max {
		return nil, errors.New("frame length invalid or exceeds budget")
	}
	buf := make([]byte, size)
	if _, err := io.ReadFull(c, buf); err != nil {
		return nil, fmt.Errorf("truncated frame: %w", err)
	}
	return buf, nil
}
func writeFrame(c net.Conn, b []byte, max uint32) error {
	if len(b) == 0 || len(b) > int(max) {
		return errors.New("output frame exceeds budget")
	}
	var header [4]byte
	binary.BigEndian.PutUint32(header[:], uint32(len(b)))
	if _, err := c.Write(header[:]); err != nil {
		return err
	}
	for len(b) > 0 {
		n, err := c.Write(b)
		if err != nil {
			return err
		}
		if n == 0 {
			return io.ErrShortWrite
		}
		b = b[n:]
	}
	return nil
}

// Serve processes bounded, one-request-per-connection messages. Concurrency is
// capped and the listener is owned by the caller. Closing ctx stops acceptance.
// No endpoint has access to nftables, shell, Cloud or privileged capabilities.
func Serve(ctx context.Context, ln *net.UnixListener, allowedUID uint32, boundary *controlgate.Boundary, p policy.Bundle, clock func() time.Time) error {
	if ln == nil || boundary == nil || clock == nil {
		return errors.New("IPC missing listener/boundary/clock")
	}
	if err := p.Validate(); err != nil {
		return err
	}
	sem := make(chan struct{}, MaxClients)
	var wg sync.WaitGroup
	done := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			ln.Close()
		case <-done:
		}
	}()
	defer close(done)
	defer wg.Wait()
	for {
		c, err := ln.AcceptUnix()
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return err
		}
		select {
		case sem <- struct{}{}:
			wg.Add(1)
			go func() {
				defer wg.Done()
				defer func() { <-sem }()
				defer c.Close()
				serveOne(c, allowedUID, boundary, p, clock)
			}()
		default:
			c.Close() // full queue is fail-closed, never block accept indefinitely
		}
	}
}
func serveOne(c *net.UnixConn, allowedUID uint32, boundary *controlgate.Boundary, p policy.Bundle, clock func() time.Time) {
	c.SetDeadline(time.Now().Add(SocketTimeout))
	uid, err := peerUID(c)
	if err != nil || uid != allowedUID {
		return
	} // do not disclose trust details to unknown peers
	raw, err := readFrame(c, MaxRequestBytes)
	if err != nil {
		return
	}
	var message controlgate.SignedIntent
	if err := strictjson.Decode(bytes.NewReader(raw), MaxRequestBytes, &message); err != nil {
		return
	}
	response := boundary.Evaluate(message, p, clock())
	data, err := json.Marshal(response)
	if err != nil {
		return
	}
	_ = writeFrame(c, data, MaxResponseBytes)
}

// Call executes one bounded request and verifies that the server socket peer
// belongs to the same local UID. A same-UID malicious process is NOT excluded;
// production IPC still needs dedicated UIDs/key custody/owner auth.
func Call(ctx context.Context, path string, message controlgate.SignedIntent) (controlgate.Receipt, error) {
	var out controlgate.Receipt
	if err := message.Intent.Validate(); err != nil {
		return out, err
	}
	data, err := json.Marshal(message)
	if err != nil {
		return out, err
	}
	if len(data) > MaxRequestBytes {
		return out, errors.New("request exceeds IPC size cap")
	}
	ctx, cancel := context.WithTimeout(ctx, SocketTimeout)
	defer cancel()
	d := net.Dialer{}
	raw, err := d.DialContext(ctx, "unix", path)
	if err != nil {
		return out, err
	}
	defer raw.Close()
	c, ok := raw.(*net.UnixConn)
	if !ok {
		return out, errors.New("unexpected non-Unix connection")
	}
	uid, err := peerUID(c)
	if err != nil || uid != uint32(os.Geteuid()) {
		return out, errors.New("IPC server UID mismatch")
	}
	deadline := time.Now().Add(SocketTimeout)
	if d, ok := ctx.Deadline(); ok && d.Before(deadline) {
		deadline = d
	}
	c.SetDeadline(deadline)
	if err := writeFrame(c, data, MaxRequestBytes); err != nil {
		return out, err
	}
	reply, err := readFrame(c, MaxResponseBytes)
	if err != nil {
		return out, err
	}
	if err := strictjson.Decode(bytes.NewReader(reply), MaxResponseBytes, &out); err != nil {
		return out, err
	}
	if out.Schema != "lightluna.control-receipt/v1alpha1" || out.RequestID != message.Intent.RequestID || out.KernelWrites ||
		out.ExecutorStatus != "NOT_AVAILABLE" || (out.Decision != "ADVISORY_ONLY" && out.Decision != "REJECTED") {
		return controlgate.Receipt{}, errors.New("unexpected or unsafe IPC receipt")
	}
	return out, nil
}
