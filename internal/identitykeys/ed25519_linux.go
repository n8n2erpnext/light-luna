//go:build linux

// Package identitykeys holds test-only Ed25519 material in separate directories.
// No private key ever needs to be read by the Gate. Real UID separation must
// be provisioned and tested in a dedicated VM; this API does not create users.
package identitykeys

import (
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"
)

const PrivateName = "agent-private.ed25519"
const PublicName = "agent-public.ed25519"

func ownerDir(dir string) error {
	if os.Geteuid() == 0 {
		return errors.New("refuses root")
	}
	st, err := os.Lstat(dir)
	if err != nil {
		return err
	}
	own, ok := st.Sys().(*syscall.Stat_t)
	if !ok || !st.IsDir() || st.Mode().Perm() != 0700 || own.Uid != uint32(os.Geteuid()) {
		return errors.New("identity directory must be owned non-symlink mode 0700")
	}
	return nil
}
func secureRead(path string, size int) ([]byte, error) {
	fd, err := syscall.Open(path, syscall.O_RDONLY|syscall.O_NOFOLLOW|syscall.O_CLOEXEC, 0)
	if err != nil {
		return nil, err
	}
	f := os.NewFile(uintptr(fd), path)
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		return nil, err
	}
	own, ok := st.Sys().(*syscall.Stat_t)
	if !ok || !st.Mode().IsRegular() || st.Mode().Perm() != 0600 || own.Uid != uint32(os.Geteuid()) || own.Nlink != 1 || st.Size() != int64(size) {
		return nil, errors.New("wrong identity key file owner/mode/link count or length")
	}
	buf := make([]byte, size)
	if _, err := io.ReadFull(f, buf); err != nil {
		return nil, err
	}
	return buf, nil
}
func secureWrite(path string, content []byte) error {
	fd, err := syscall.Open(path, syscall.O_WRONLY|syscall.O_CREAT|syscall.O_EXCL|syscall.O_NOFOLLOW|syscall.O_CLOEXEC, 0600)
	if err != nil {
		return err
	}
	f := os.NewFile(uintptr(fd), path)
	n, err := f.Write(content)
	if err != nil || n != len(content) {
		f.Close()
		return errors.New("write identity key failed")
	}
	if err := f.Sync(); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

// Init creates both directories and new keys; refuses existing paths to
// avoid accidental overwriting. In the same-UID test runner, this is NOT
// separate OS ownership. Production enrollment must provision directories
// under distinct users in an isolated installer.
func Init(agentDir, gateDir string) error {
	if os.Geteuid() == 0 || agentDir == gateDir {
		return errors.New("invalid identity separation")
	}
	if err := os.Mkdir(agentDir, 0700); err != nil {
		return err
	}
	agentSuccess := false
	defer func() {
		if !agentSuccess {
			os.RemoveAll(agentDir)
		}
	}()
	if err := os.Mkdir(gateDir, 0700); err != nil {
		return err
	}
	gateSuccess := false
	defer func() {
		if !gateSuccess {
			os.RemoveAll(gateDir)
		}
	}()
	if err := ownerDir(agentDir); err != nil {
		return err
	}
	if err := ownerDir(gateDir); err != nil {
		return err
	}
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}
	defer func() {
		for i := range priv {
			priv[i] = 0
		}
	}()
	if err := secureWrite(filepath.Join(agentDir, PrivateName), priv); err != nil {
		return err
	}
	if err := secureWrite(filepath.Join(gateDir, PublicName), pub); err != nil {
		return err
	}
	for _, dir := range []string{agentDir, gateDir} {
		f, err := os.Open(dir)
		if err != nil {
			return err
		}
		err = f.Sync()
		f.Close()
		if err != nil {
			return fmt.Errorf("sync identity directory: %w", err)
		}
	}
	agentSuccess, gateSuccess = true, true
	return nil
}
func LoadPrivate(dir string) (ed25519.PrivateKey, error) {
	if err := ownerDir(dir); err != nil {
		return nil, err
	}
	b, err := secureRead(filepath.Join(dir, PrivateName), ed25519.PrivateKeySize)
	if err != nil {
		return nil, err
	}
	return ed25519.PrivateKey(b), nil
}
func LoadPublic(dir string) (ed25519.PublicKey, error) {
	if err := ownerDir(dir); err != nil {
		return nil, err
	}
	b, err := secureRead(filepath.Join(dir, PublicName), ed25519.PublicKeySize)
	if err != nil {
		return nil, err
	}
	return ed25519.PublicKey(b), nil
}
