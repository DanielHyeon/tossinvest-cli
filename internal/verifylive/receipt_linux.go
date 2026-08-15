//go:build linux

package verifylive

import (
	"fmt"
	"path/filepath"
	"strings"

	"golang.org/x/sys/unix"
)

// openCausalReceipt binds every check to the file descriptor it will use. In
// particular, a pathname replacement after the directory is opened cannot make
// the leaf creation or directory fsync operate on a different directory.
func openCausalReceipt(dir string, hooks receiptHooks) (*CausalReceipt, error) {
	hooks = hooks.withDefaults()
	dirFD, err := openReceiptDirectoryNoSymlinks(dir)
	if err != nil {
		return nil, fmt.Errorf("verifylive: opening causal receipt directory: %w", err)
	}
	df := receiptFileFromFD(uintptr(dirFD), dir)
	if df == nil {
		_ = unix.Close(dirFD)
		return nil, fmt.Errorf("verifylive: opening causal receipt directory: invalid file descriptor")
	}
	closeDir := true
	defer func() {
		if closeDir {
			_ = df.Close()
		}
	}()
	if err := validateReceiptDirectoryFD(dirFD); err != nil {
		return nil, err
	}
	if hooks.afterDirOpen != nil {
		if err := hooks.afterDirOpen(dir); err != nil {
			return nil, fmt.Errorf("verifylive: preparing causal receipt directory validation: %w", err)
		}
	}

	var random [16]byte
	if _, err := hooks.random(random[:]); err != nil {
		return nil, fmt.Errorf("verifylive: generating causal receipt run id: %w", err)
	}
	runID := "m0-" + hexEncode(random[:])
	name := runID + ".jsonl"
	fileFD, err := unix.Openat(dirFD, name, unix.O_WRONLY|unix.O_CLOEXEC|unix.O_CREAT|unix.O_EXCL|unix.O_NOFOLLOW, 0o600)
	if err != nil {
		return nil, fmt.Errorf("verifylive: creating causal receipt: %w", err)
	}
	f := receiptFileFromFD(uintptr(fileFD), name)
	if f == nil {
		_ = unix.Close(fileFD)
		return nil, fmt.Errorf("verifylive: creating causal receipt: invalid file descriptor")
	}
	closeFile := true
	defer func() {
		if closeFile {
			_ = f.Close()
		}
	}()
	if hooks.afterCreate != nil {
		if err := hooks.afterCreate(filepath.Join(dir, name)); err != nil {
			return nil, fmt.Errorf("verifylive: preparing causal receipt validation: %w", err)
		}
	}
	if err := validateReceiptFileFD(fileFD); err != nil {
		return nil, err
	}

	r := newCausalReceipt(f, df, runID, hooks.syncFile)
	closeFile, closeDir = false, false
	if err := r.finishOpen(hooks.syncDir); err != nil {
		return nil, err
	}
	return r, nil
}

// openReceiptDirectoryNoSymlinks anchors traversal at the filesystem root and
// opens every component relative to the preceding descriptor. No component can
// be exchanged for a symlink between validation and later use.
func openReceiptDirectoryNoSymlinks(path string) (int, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return -1, fmt.Errorf("verifylive: resolving causal receipt directory: %w", err)
	}
	flags := unix.O_RDONLY | unix.O_CLOEXEC | unix.O_DIRECTORY | unix.O_NOFOLLOW
	fd, err := unix.Open(string(filepath.Separator), flags, 0)
	if err != nil {
		return -1, fmt.Errorf("verifylive: opening causal receipt root: %w", err)
	}
	if err := validateReceiptPathComponentFD(fd); err != nil {
		_ = unix.Close(fd)
		return -1, err
	}
	clean := strings.TrimPrefix(filepath.Clean(abs), string(filepath.Separator))
	if clean == "" {
		return fd, nil
	}
	for _, component := range strings.Split(clean, string(filepath.Separator)) {
		if component == "" || component == "." {
			continue
		}
		next, openErr := unix.Openat(fd, component, flags, 0)
		_ = unix.Close(fd)
		if openErr != nil {
			return -1, fmt.Errorf("verifylive: opening causal receipt directory component %q: %w", component, openErr)
		}
		fd = next
		if err := validateReceiptPathComponentFD(fd); err != nil {
			_ = unix.Close(fd)
			return -1, err
		}
	}
	return fd, nil
}

func validateReceiptPathComponentFD(fd int) error {
	var st unix.Stat_t
	if err := unix.Fstat(fd, &st); err != nil {
		return fmt.Errorf("verifylive: stat causal receipt directory component: %w", err)
	}
	if st.Mode&unix.S_IFMT != unix.S_IFDIR {
		return fmt.Errorf("verifylive: causal receipt path component must be a non-symlink directory")
	}
	return nil
}

func validateReceiptDirectoryFD(fd int) error {
	var st unix.Stat_t
	if err := unix.Fstat(fd, &st); err != nil {
		return fmt.Errorf("verifylive: stat causal receipt directory: %w", err)
	}
	if st.Mode&unix.S_IFMT != unix.S_IFDIR || st.Uid != currentReceiptUID() || !exactPrivateMode(st.Mode, 0o700) {
		return fmt.Errorf("verifylive: causal receipt directory must be current-uid non-symlink mode 0700 directory")
	}
	return nil
}

func validateReceiptFileFD(fd int) error {
	var st unix.Stat_t
	if err := unix.Fstat(fd, &st); err != nil {
		return fmt.Errorf("verifylive: stat causal receipt: %w", err)
	}
	if st.Mode&unix.S_IFMT != unix.S_IFREG || st.Uid != currentReceiptUID() || !exactPrivateMode(st.Mode, 0o600) {
		return fmt.Errorf("verifylive: causal receipt must be current-uid regular non-symlink mode 0600")
	}
	return nil
}

func exactPrivateMode(mode uint32, want uint32) bool {
	const permissionAndSpecial = unix.S_IRWXU | unix.S_IRWXG | unix.S_IRWXO | unix.S_ISUID | unix.S_ISGID | unix.S_ISVTX
	return mode&permissionAndSpecial == want
}

// Keep hexadecimal rendering here so this Linux-only construction file owns all
// path creation details; the generic receipt lifecycle never assembles a path.
func hexEncode(b []byte) string {
	const hex = "0123456789abcdef"
	out := make([]byte, len(b)*2)
	for i, v := range b {
		out[i*2], out[i*2+1] = hex[v>>4], hex[v&0x0f]
	}
	return string(out)
}
