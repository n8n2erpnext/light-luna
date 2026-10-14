//go:build linux

package controlgate

// Development-only replay journal. This never authorizes a kernel action.
// It is fail-closed: corrupt or full journals cannot silently degrade to RAM.
import (
	"bufio"
	"bytes"
	"crypto/ed25519"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/n8n2erpnext/light-luna/internal/strictjson"
)

const MaxReplayJournalBytes int64 = 2 << 20

type replayRecord struct {
	Nonce      string    `json:"nonce"`
	RequestID  string    `json:"requestId"`
	ExpiresUTC time.Time `json:"expiresUtc"`
}

// ReplayJournal must be opened once per Gate process. A nonblocking flock
// prevents two Gate processes writing the same journal at the same time.
type ReplayJournal struct {
	file   *os.File
	failed bool
}

func OpenReplayJournal(path string, now time.Time) (*ReplayJournal, map[string]time.Time, error) {
	if os.Geteuid() == 0 {
		return nil, nil, errors.New("replay journal must run nonroot")
	}
	dir := filepath.Dir(path)
	info, err := os.Lstat(dir)
	if err != nil {
		return nil, nil, err
	}
	owner, ok := info.Sys().(*syscall.Stat_t)
	if !ok || !info.IsDir() || info.Mode().Perm() != 0700 || owner.Uid != uint32(os.Geteuid()) {
		return nil, nil, errors.New("replay journal parent must be private 0700 directory owned by Gate UID")
	}
	// O_NOFOLLOW defeats symlink tricks on the journal path; existing content
	// must be trusted only as a candidate and revalidated on each restart.
	fd, err := syscall.Open(path, syscall.O_RDWR|syscall.O_CREAT|syscall.O_APPEND|syscall.O_NOFOLLOW|syscall.O_CLOEXEC, 0600)
	if err != nil {
		return nil, nil, err
	}
	f := os.NewFile(uintptr(fd), path)
	success := false
	defer func() {
		if !success {
			f.Close()
		}
	}()
	meta, err := f.Stat()
	if err != nil {
		return nil, nil, err
	}
	own, ok := meta.Sys().(*syscall.Stat_t)
	if !ok || !meta.Mode().IsRegular() || meta.Mode().Perm() != 0600 || own.Uid != uint32(os.Geteuid()) || own.Nlink != 1 {
		return nil, nil, errors.New("replay journal must be a single-link private 0600 regular file")
	}
	if meta.Size() > MaxReplayJournalBytes {
		return nil, nil, errors.New("journal capacity exhausted; fail closed")
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		return nil, nil, fmt.Errorf("journal already owned by another Gate: %w", err)
	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return nil, nil, err
	}
	seen := map[string]time.Time{}
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 4096), 4096)
	var consumed int64
	for scanner.Scan() {
		line := append([]byte(nil), scanner.Bytes()...)
		consumed += int64(len(line) + 1)
		if consumed > MaxReplayJournalBytes {
			return nil, nil, errors.New("journal exceeds quota")
		}
		var rec replayRecord
		if err := strictjson.Decode(bytes.NewReader(line), 512, &rec); err != nil {
			return nil, nil, fmt.Errorf("corrupt replay journal: %w", err)
		}
		if !safeID.MatchString(rec.Nonce) || !safeID.MatchString(rec.RequestID) || rec.ExpiresUTC.IsZero() {
			return nil, nil, errors.New("invalid replay journal record")
		}
		key1, key2 := "nonce:"+rec.Nonce, "request:"+rec.RequestID
		if _, present := seen[key1]; present {
			return nil, nil, errors.New("duplicate live nonce in replay journal")
		}
		if _, present := seen[key2]; present {
			return nil, nil, errors.New("duplicate live request in replay journal")
		}
		// Keep only live tickets in memory; older records still count toward disk cap.
		if now.Before(rec.ExpiresUTC) {
			seen[key1] = rec.ExpiresUTC
			seen[key2] = rec.ExpiresUTC
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, nil, err
	}
	if meta.Size() != consumed && meta.Size() > 0 {
		return nil, nil, errors.New("partial/corrupt replay journal tail")
	}
	success = true
	return &ReplayJournal{file: f}, seen, nil
}
func (j *ReplayJournal) record(id, nonce string, expires time.Time) error {
	if j == nil || j.file == nil || j.failed {
		return errors.New("journal unavailable; fail closed")
	}
	rec := replayRecord{Nonce: nonce, RequestID: id, ExpiresUTC: expires}
	data, err := json.Marshal(rec)
	if err != nil {
		j.failed = true
		return err
	}
	data = append(data, '\n')
	info, err := j.file.Stat()
	if err != nil {
		j.failed = true
		return err
	}
	if info.Size()+int64(len(data)) > MaxReplayJournalBytes {
		j.failed = true
		return errors.New("replay journal capacity exhausted; fail closed")
	}
	n, err := j.file.Write(data)
	if err != nil || n != len(data) {
		j.failed = true
		return errors.New("replay journal write failed; fail closed")
	}
	if err := j.file.Sync(); err != nil {
		j.failed = true
		return fmt.Errorf("replay journal sync failed: %w", err)
	}
	return nil
}

// NewBoundaryWithJournal restores previously accepted non-expired request IDs
// and nonces. This uses one nonroot Gate process and a single locked journal.
func NewBoundaryWithJournal(key []byte, siteID, actorID, path string, now time.Time) (*Boundary, *ReplayJournal, error) {
	j, seen, err := OpenReplayJournal(path, now)
	if err != nil {
		return nil, nil, err
	}
	b, err := NewBoundary(key, siteID, actorID)
	if err != nil {
		j.Close()
		return nil, nil, err
	}
	b.seen = seen
	b.journal = j
	return b, j, nil
}

// NewEd25519BoundaryWithJournal keeps only the Agent public key on Gate.
// This is still a development-only journal without power-loss/rollback proof.
func NewEd25519BoundaryWithJournal(publicKey ed25519.PublicKey, siteID, actorID, path string, now time.Time) (*Boundary, *ReplayJournal, error) {
	j, seen, err := OpenReplayJournal(path, now)
	if err != nil {
		return nil, nil, err
	}
	b, err := NewEd25519Boundary(publicKey, siteID, actorID)
	if err != nil {
		j.Close()
		return nil, nil, err
	}
	b.seen = seen
	b.journal = j
	return b, j, nil
}
func (j *ReplayJournal) Close() error {
	if j == nil || j.file == nil {
		return nil
	}
	err := j.file.Close()
	j.file = nil
	j.failed = true
	return err
}
