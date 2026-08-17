//go:build unix

package official

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/sys/unix"
)

// a112MBUSCachedToken reads only the existing cache through pinned directory
// descriptors. It deliberately never calls token(), refresh(), exchange(), or
// saveCache().
func a112MBUSCachedToken(path string, now time.Time) (string, error) {
	fd, err := a112MBUSOpenTokenNoFollow(path)
	if err != nil {
		return "", err
	}
	file := os.NewFile(uintptr(fd), filepath.Base(path))
	if file == nil {
		_ = unix.Close(fd)
		return "", fmt.Errorf("invalid opened token descriptor")
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, 64<<10))
	if err != nil {
		return "", err
	}
	var token cachedToken
	if err := json.Unmarshal(data, &token); err != nil || !isStillValid(&token, now) || token.AccessToken == "" {
		return "", fmt.Errorf("cache is missing, malformed, or expired")
	}
	return token.AccessToken, nil
}

func a112MBUSOpenTokenNoFollow(path string) (int, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return -1, err
	}
	components := strings.Split(strings.TrimPrefix(filepath.Clean(abs), string(filepath.Separator)), string(filepath.Separator))
	if len(components) == 0 || components[len(components)-1] == "" {
		return -1, fmt.Errorf("token path is empty")
	}
	flags := unix.O_RDONLY | unix.O_CLOEXEC | unix.O_DIRECTORY | unix.O_NOFOLLOW
	dirFD, err := unix.Open(string(filepath.Separator), flags, 0)
	if err != nil {
		return -1, err
	}
	for _, component := range components[:len(components)-1] {
		if component == "" || component == "." || component == ".." {
			_ = unix.Close(dirFD)
			return -1, fmt.Errorf("unsafe token path component")
		}
		nextFD, openErr := unix.Openat(dirFD, component, flags, 0)
		_ = unix.Close(dirFD)
		if openErr != nil {
			return -1, openErr
		}
		dirFD = nextFD
		var directory unix.Stat_t
		if err := unix.Fstat(dirFD, &directory); err != nil || directory.Mode&unix.S_IFMT != unix.S_IFDIR {
			_ = unix.Close(dirFD)
			return -1, fmt.Errorf("token ancestor is not directory")
		}
	}
	defer unix.Close(dirFD)
	fd, err := unix.Openat(dirFD, components[len(components)-1], unix.O_RDONLY|unix.O_CLOEXEC|unix.O_NOFOLLOW|unix.O_NONBLOCK, 0)
	if err != nil {
		return -1, err
	}
	var stat unix.Stat_t
	if err := unix.Fstat(fd, &stat); err != nil || stat.Mode&unix.S_IFMT != unix.S_IFREG || stat.Uid != uint32(os.Geteuid()) || stat.Mode&0o777 != 0o600 {
		_ = unix.Close(fd)
		return -1, fmt.Errorf("token cache is not current-uid regular 0600")
	}
	return fd, nil
}
