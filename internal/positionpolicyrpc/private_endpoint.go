package positionpolicyrpc

// private_endpoint.go generalises the checks this package already performs on its
// own control endpoint so a second engine-owned endpoint does not have to
// re-implement them.
//
// Every function here is the name-independent form of one that already exists:
// ValidateControlDirectory and ValidateRuntimeControlDirectory both pin a
// filename and then call the same private directory check, and PrepareRuntimeSocket
// and ValidateRuntimeSocket do the same for a socket. The naming is what differs
// between endpoints; the checks are what must not.
//
// # Why they live here rather than next to the new endpoint
//
// The primitives underneath them — symlink-traversal rejection, owner and
// hard-link verification, the no-follow open — are unexported in this package and
// portability-split across private_fs_unix.go / private_fs_other.go. Copying them
// to a second package would be copying the security boundary, and a copied check
// is one that stops matching (a098, design D7.1).
//
// Nothing existing was changed to add these: the name-pinned functions still call
// the same private helpers directly.

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// ValidatePrivateControlDirectory verifies any engine-owned control directory:
// its parent must be a plausible engine directory and the leaf itself must be an
// exact 0700 real directory owned by this process.
//
// It does not check the directory's name. The caller owns that, because the name
// is the only thing that distinguishes one endpoint from another.
func ValidatePrivateControlDirectory(path string) error {
	clean := strings.TrimSpace(path)
	if clean == "" {
		return errors.New("private endpoint: control directory path is empty")
	}
	if err := ValidateEngineDirectory(filepath.Dir(clean)); err != nil {
		return err
	}
	return validatePrivateDirectory(clean, true)
}

// ValidatePrivateSocket verifies that path is an exact 0600 Unix socket owned by
// this process, inside a valid control directory.
func ValidatePrivateSocket(path string) error {
	info, err := statPrivateSocket(path)
	if err != nil {
		return err
	}
	return validateOwnerAndLinks(info, false)
}

// PreparePrivateSocket removes a stale endpoint so a listener can bind.
//
// It refuses to remove anything that is not already exactly what this process
// would have created. A previous run's socket is a socket; something else at that
// path is somebody else's file, and unlinking it would be this process deleting
// what it did not make.
func PreparePrivateSocket(path string) error {
	info, err := statPrivateSocket(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if err := validateOwnerAndLinks(info, false); err != nil {
		return err
	}
	return os.Remove(strings.TrimSpace(path))
}

func statPrivateSocket(path string) (os.FileInfo, error) {
	clean := strings.TrimSpace(path)
	if clean == "" {
		return nil, errors.New("private endpoint: socket path is empty")
	}
	if err := ValidatePrivateControlDirectory(filepath.Dir(clean)); err != nil {
		return nil, err
	}
	info, err := os.Lstat(clean)
	if err != nil {
		return nil, err
	}
	if info.Mode()&os.ModeSocket == 0 || info.Mode()&os.ModeSymlink != 0 ||
		info.Mode().Perm() != 0o600 {
		return nil, errors.New("private endpoint: path is not an exact 0600 Unix socket")
	}
	return info, nil
}

// OpenPrivateDescriptorFile opens a published descriptor without following
// symlinks and verifies that the file it opened is the file it looked at.
//
// The three-way identity check (before / opened / current) is the precedent's and
// is kept exactly: a descriptor swapped between the stat and the open would
// otherwise hand the caller somebody else's token.
func OpenPrivateDescriptorFile(path string) (*os.File, error) {
	clean := strings.TrimSpace(path)
	if clean == "" {
		return nil, errors.New("private endpoint: descriptor path is empty")
	}
	if err := ValidatePrivateControlDirectory(filepath.Dir(clean)); err != nil {
		return nil, err
	}
	before, err := os.Lstat(clean)
	if err != nil {
		return nil, err
	}
	if err := validateDescriptorInfo(before); err != nil {
		return nil, err
	}
	f, err := openDescriptorNoFollow(clean)
	if err != nil {
		return nil, err
	}
	opened, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, err
	}
	current, err := os.Lstat(clean)
	if err != nil || !os.SameFile(before, opened) || !os.SameFile(opened, current) {
		f.Close()
		return nil, errors.New("private endpoint: descriptor changed while it was opened")
	}
	if err := validateDescriptorInfo(opened); err != nil {
		f.Close()
		return nil, err
	}
	return f, nil
}
