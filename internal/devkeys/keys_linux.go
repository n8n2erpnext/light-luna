//go:build linux

// Package devkeys handles throwaway same-UID secrets for P0 isolated tests.
// NOT a production credential broker, owner login, TPM or service key vault.
package devkeys

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"
)

const KeyName = "p0-shared.key"

func VerifyPrivateDir(dir string) error {
	if os.Geteuid() == 0 {
		return errors.New("dev session refuses root")
	}
	info, err := os.Lstat(dir)
	if err != nil {
		return err
	}
	st, ok := info.Sys().(*syscall.Stat_t)
	if !ok || !info.IsDir() || info.Mode().Perm() != 0700 || st.Uid != uint32(os.Geteuid()) {
		return errors.New("directory must be non-symlink private 0700 owned by current UID")
	}
	return nil
}

// Init refuses any existing directory; callers must use isolated scratch space.
func Init(dir string) error {
	if os.Geteuid() == 0 {
		return errors.New("cannot initialize P0 key as root")
	}
	if err := os.Mkdir(dir, 0700); err != nil {
		return err
	}
	success := false
	defer func() {
		if !success {
			os.RemoveAll(dir)
		}
	}()
	if err := VerifyPrivateDir(dir); err != nil {
		return err
	}
	fd, err := syscall.Open(filepath.Join(dir, KeyName), syscall.O_WRONLY|syscall.O_CREAT|syscall.O_EXCL|syscall.O_NOFOLLOW|syscall.O_CLOEXEC, 0600)
	if err != nil {
		return err
	}
	file := os.NewFile(uintptr(fd), KeyName)
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		file.Close()
		return err
	}
	defer func() {
		for i := range key {
			key[i] = 0
		}
	}()
	if n, err := file.Write(key); err != nil || n != len(key) {
		file.Close()
		return errors.New("cannot store random demo secret")
	}
	if err := file.Sync(); err != nil {
		file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	syncErr := d.Sync()
	d.Close()
	if syncErr != nil {
		return fmt.Errorf("sync private demo directory: %w", syncErr)
	}
	success = true
	return nil
}
func Load(dir string) ([]byte, error) {
	if err := VerifyPrivateDir(dir); err != nil {
		return nil, err
	}
	p := filepath.Join(dir, KeyName)
	fd, err := syscall.Open(p, syscall.O_RDONLY|syscall.O_NOFOLLOW|syscall.O_CLOEXEC, 0)
	if err != nil {
		return nil, err
	}
	f := os.NewFile(uintptr(fd), p)
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	st, ok := info.Sys().(*syscall.Stat_t)
	if !ok || !info.Mode().IsRegular() || info.Mode().Perm() != 0600 || st.Uid != uint32(os.Geteuid()) || st.Nlink != 1 || info.Size() != 32 {
		return nil, errors.New("demo key must be single-link 0600 regular file owned by current UID with 32 bytes")
	}
	key := make([]byte, 32)
	if _, err := io.ReadFull(f, key); err != nil {
		return nil, err
	}
	return key, nil
}
