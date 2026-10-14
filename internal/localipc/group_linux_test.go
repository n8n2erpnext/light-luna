//go:build linux

package localipc

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGroupSocketPermissionsAndPinnedGID(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("P0 nonroot only")
	}
	dir := filepath.Join(t.TempDir(), "shared")
	if err := os.Mkdir(dir, 0710); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(dir, 0710); err != nil {
		t.Fatal(err)
	}
	gid := uint32(os.Getegid())
	ln, path, err := ListenGroup(dir, gid)
	if err != nil {
		t.Fatal(err)
	}
	stat, err := os.Lstat(path)
	if err != nil || stat.Mode().Perm() != 0660 {
		t.Fatalf("bad group socket mode %+v %v", stat, err)
	}
	if other, _, err := ListenGroup(dir, gid); err == nil {
		other.Close()
		t.Fatal("overwrote active socket")
	}
	ln.Close()
	if wrong, _, err := ListenGroup(dir, gid+1); err == nil {
		wrong.Close()
		t.Fatal("wrong IPC GID accepted")
	}
	os.Chmod(dir, 0770)
	if weak, _, err := ListenGroup(dir, gid); err == nil {
		weak.Close()
		t.Fatal("writable group directory accepted")
	}
}
