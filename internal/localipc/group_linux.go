//go:build linux

package localipc

// Group socket for future separately configured Agent UID + shared IPC GID.
// Group directory 0710 allows traversal without listing; socket 0660 allows
// group connection. No chown, sudo, user/group mutation or daemon installation.
import (
	"errors"
	"net"
	"os"
	"path/filepath"
	"syscall"
)

func ListenGroup(dir string, expectedGID uint32) (*net.UnixListener, string, error) {
	if os.Geteuid() == 0 {
		return nil, "", errors.New("P0 group listener refuses root")
	}
	info, err := os.Lstat(dir)
	if err != nil {
		return nil, "", err
	}
	st, ok := info.Sys().(*syscall.Stat_t)
	if !ok || !info.IsDir() || info.Mode().Perm() != 0710 || st.Uid != uint32(os.Geteuid()) || st.Gid != expectedGID {
		return nil, "", errors.New("group IPC directory must be real 0710 dir owned by Gate UID and expected GID")
	}
	path := filepath.Join(dir, "gate.sock")
	if len(path) > 103 {
		return nil, "", errors.New("group socket path too long")
	}
	if _, err := os.Lstat(path); err == nil {
		return nil, "", errors.New("existing group socket cannot be replaced")
	} else if !os.IsNotExist(err) {
		return nil, "", err
	}
	ln, err := net.ListenUnix("unix", &net.UnixAddr{Name: path, Net: "unix"})
	if err != nil {
		return nil, "", err
	}
	ln.SetUnlinkOnClose(true)
	// Temporarily created socket group should inherit configured group from a
	// setgid directory in a real service deployment. Verify actual GID rather
	// than attempting any chown inside the agent.
	if err := os.Chmod(path, 0660); err != nil {
		ln.Close()
		return nil, "", err
	}
	stInfo, err := os.Lstat(path)
	if err != nil {
		ln.Close()
		return nil, "", err
	}
	own, ok := stInfo.Sys().(*syscall.Stat_t)
	if !ok || stInfo.Mode().Perm() != 0660 || own.Uid != uint32(os.Geteuid()) || own.Gid != expectedGID {
		ln.Close()
		return nil, "", errors.New("group socket UID/GID/permissions do not match expected Gate ownership")
	}
	return ln, path, nil
}
