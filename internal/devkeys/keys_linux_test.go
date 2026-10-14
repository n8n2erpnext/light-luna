//go:build linux

package devkeys

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDemoKeyAndPrivatePath(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("demo refuses root")
	}
	root := t.TempDir()
	path := filepath.Join(root, "dev-session")
	if err := Init(path); err != nil {
		t.Fatal(err)
	}
	if err := Init(path); err == nil {
		t.Fatal("reinitialized existing key store")
	}
	first, err := Load(path)
	if err != nil || len(first) != 32 {
		t.Fatalf("cannot load demo key: %v", err)
	}
	st, err := os.Stat(filepath.Join(path, KeyName))
	if err != nil || st.Mode().Perm() != 0600 {
		t.Fatal("key perms not 0600")
	}
	os.Chmod(filepath.Join(path, KeyName), 0644)
	if _, err := Load(path); err == nil {
		t.Fatal("world-readable key accepted")
	}
	os.Chmod(filepath.Join(path, KeyName), 0600)
	os.Chmod(path, 0755)
	if _, err := Load(path); err == nil {
		t.Fatal("nonprivate directory accepted")
	}
	os.Chmod(path, 0700)
	external := filepath.Join(root, "external")
	os.WriteFile(external, first, 0600)
	os.Remove(filepath.Join(path, KeyName))
	os.Symlink(external, filepath.Join(path, KeyName))
	if _, err := Load(path); err == nil {
		t.Fatal("symlink accepted")
	}
}
