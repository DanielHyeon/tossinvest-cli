//go:build unix

package httpapi

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/unix"
)

func validateSecurityStoreDirectory(path string) error {
	return validateSecurityStoreDirectoryForUID(path, uint32(os.Geteuid()))
}

// validateSecurityStoreDirectoryForUID walks every component without following
// symlinks. The final directory is an authority boundary and must be private to
// root or the service identity; root-owned sticky ancestors such as /tmp are
// permitted, but no replaceable intermediate component is.
func validateSecurityStoreDirectoryForUID(path string, serviceUID uint32) error {
	abs, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return fmt.Errorf("resolving directory: %w", err)
	}
	components := strings.Split(strings.TrimPrefix(abs, string(filepath.Separator)), string(filepath.Separator))
	if len(components) < 1 || components[len(components)-1] == "" {
		return errors.New("a dedicated security database directory is required")
	}
	fd, err := unix.Open(string(filepath.Separator), unix.O_RDONLY|unix.O_CLOEXEC|unix.O_DIRECTORY, 0)
	if err != nil {
		return fmt.Errorf("opening filesystem root: %w", err)
	}
	defer func() { _ = unix.Close(fd) }()
	var root unix.Stat_t
	if err := unix.Fstat(fd, &root); err != nil {
		return fmt.Errorf("inspecting filesystem root: %w", err)
	}
	for index, component := range components {
		if component == "" || component == "." || component == ".." {
			return errors.New("ambiguous directory component")
		}
		next, openErr := unix.Openat(fd, component, unix.O_RDONLY|unix.O_CLOEXEC|unix.O_DIRECTORY|unix.O_NOFOLLOW, 0)
		if openErr != nil {
			return fmt.Errorf("opening directory component %q: %w", component, openErr)
		}
		var stat unix.Stat_t
		if statErr := unix.Fstat(next, &stat); statErr != nil {
			_ = unix.Close(next)
			return fmt.Errorf("inspecting directory component %q: %w", component, statErr)
		}
		final := index == len(components)-1
		if validationErr := validateSecurityStoreDirectoryStat(&stat, serviceUID, root.Uid, final); validationErr != nil {
			_ = unix.Close(next)
			return fmt.Errorf("directory component %q: %w", component, validationErr)
		}
		_ = unix.Close(fd)
		fd = next
	}
	return nil
}

func validateSecurityStoreDirectoryStat(stat *unix.Stat_t, serviceUID, rootUID uint32, final bool) error {
	if stat == nil || stat.Mode&unix.S_IFMT != unix.S_IFDIR {
		return errors.New("not a real directory")
	}
	if stat.Uid != rootUID && stat.Uid != serviceUID {
		return errors.New("owned by another account")
	}
	if final {
		if stat.Mode&0o077 != 0 {
			return errors.New("dedicated directory must have mode 0700")
		}
		return nil
	}
	if stat.Mode&0o022 != 0 && (stat.Uid != rootUID || stat.Mode&unix.S_ISVTX == 0) {
		return errors.New("writable ancestor is not root-owned and sticky")
	}
	return nil
}

func validateSecurityStoreFile(info os.FileInfo) error {
	stat, ok := info.Sys().(*unix.Stat_t)
	if !ok || info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return errors.New("database must be a regular file, not a symlink")
	}
	rootUID := uint32(0)
	if root, err := os.Stat(string(filepath.Separator)); err == nil {
		if rootStat, rootOK := root.Sys().(*unix.Stat_t); rootOK {
			rootUID = rootStat.Uid
		}
	}
	if stat.Uid != rootUID && stat.Uid != uint32(os.Geteuid()) {
		return errors.New("database is owned by another account")
	}
	if info.Mode().Perm()&0o022 != 0 {
		return errors.New("database is group/other writable")
	}
	return nil
}
