//go:build linux

package accountbudgetruntime

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/accountbudget"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"golang.org/x/sys/unix"
)

type offlineFault func(string) error

func offlineDir(path string) (fd int, ancestors []accountbudget.FileID, err error) {
	if !filepath.IsAbs(path) || filepath.Clean(path) != path || path == "/" {
		return -1, nil, errors.New("invalid directory")
	}
	fd, err = unix.Open("/", unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC, 0)
	if err != nil {
		return -1, nil, err
	}
	fail := func(e error) (int, []accountbudget.FileID, error) { return -1, nil, errors.Join(e, unix.Close(fd)) }
	for _, part := range strings.Split(strings.TrimPrefix(path, "/"), "/") {
		var st unix.Stat_t
		if err = unix.Fstat(fd, &st); err != nil {
			return fail(err)
		}
		ancestors = append(ancestors, accountbudget.FileID{Device: uint64(st.Dev), Inode: st.Ino})
		next, e := unix.Openat(fd, part, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0)
		if e != nil {
			return fail(e)
		}
		if e = unix.Close(fd); e != nil {
			unix.Close(next)
			return -1, nil, e
		}
		fd = next
	}
	var st unix.Stat_t
	if err = unix.Fstat(fd, &st); err != nil {
		return fail(err)
	}
	if !safeOfflineDirectory(st, uint32(os.Geteuid())) {
		return fail(errors.New("unsafe directory"))
	}
	return fd, ancestors, nil
}
func safeOfflineDirectory(st unix.Stat_t, owner uint32) bool {
	return st.Uid == owner && st.Mode&unix.S_IFMT == unix.S_IFDIR && st.Mode&07777 == 0700
}
func supportedOfflineFS(kind int64) bool {
	return kind == unix.EXT4_SUPER_MAGIC || kind == unix.TMPFS_MAGIC
}
func offlineID(fd int) (accountbudget.FileID, error) {
	var st unix.Stat_t
	err := unix.Fstat(fd, &st)
	return accountbudget.FileID{Device: uint64(st.Dev), Inode: st.Ino}, err
}
func offlineFileID(fd int, name string) (accountbudget.FileID, error) {
	var st unix.Stat_t
	err := unix.Fstatat(fd, name, &st, unix.AT_SYMLINK_NOFOLLOW)
	if err != nil {
		return accountbudget.FileID{}, err
	}
	if st.Mode&unix.S_IFMT != unix.S_IFREG {
		return accountbudget.FileID{}, errors.New("not a regular file")
	}
	return accountbudget.FileID{Device: uint64(st.Dev), Inode: st.Ino}, nil
}
func offlineSameRoot(path string, expected accountbudget.FileID) error {
	fd, _, err := offlineDir(path)
	if err != nil {
		return err
	}
	id, e := offlineID(fd)
	e = errors.Join(e, unix.Close(fd))
	if e != nil {
		return e
	}
	if id != expected {
		return errors.New("root identity changed")
	}
	return nil
}
func ProvisionOffline(ctx context.Context, root, path, scope, instance string) error {
	return provisionOffline(ctx, root, path, scope, instance, nil)
}
func provisionOffline(ctx context.Context, root, path, scope, instance string, fault offlineFault) (result error) {
	if ctx == nil {
		return accountbudget.ErrUnavailable
	}
	inject := func(stage string) error {
		if fault == nil {
			return nil
		}
		return fault(stage)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if !filepath.IsAbs(path) || filepath.Clean(path) != path || path == "/" {
		return errors.New("invalid manifest path")
	}
	rel, err := filepath.Rel(root, path)
	if err != nil || rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))) {
		return errors.New("manifest inside root")
	}
	rootFD, _, err := offlineDir(root)
	if err != nil {
		return err
	}
	defer func() { result = errors.Join(result, unix.Close(rootFD)) }()
	rootID, err := offlineID(rootFD)
	if err != nil {
		return err
	}
	var fs unix.Statfs_t
	if err = unix.Fstatfs(rootFD, &fs); err != nil {
		return err
	}
	if !supportedOfflineFS(int64(fs.Type)) {
		return errors.New("unsupported filesystem")
	}
	dup, err := unix.FcntlInt(uintptr(rootFD), unix.F_DUPFD_CLOEXEC, 0)
	if err != nil {
		return err
	}
	dir := os.NewFile(uintptr(dup), root)
	names, e := dir.Readdirnames(1)
	closeErr := dir.Close()
	if e != nil && !errors.Is(e, io.EOF) {
		return errors.Join(e, closeErr)
	}
	if closeErr != nil {
		return closeErr
	}
	if len(names) != 0 {
		return errors.New("root is not empty")
	}
	parent, ancestors, err := offlineDir(filepath.Dir(path))
	if err != nil {
		return err
	}
	defer func() { result = errors.Join(result, unix.Close(parent)) }()
	parentID, err := offlineID(parent)
	if err != nil {
		return err
	}
	ancestors = append(ancestors, parentID)
	for _, id := range ancestors {
		if id == rootID {
			return errors.New("manifest inside root")
		}
	}
	name := filepath.Base(path)
	var st unix.Stat_t
	e = unix.Fstatat(parent, name, &st, unix.AT_SYMLINK_NOFOLLOW)
	if e == nil {
		return errors.New("manifest already exists")
	}
	if !errors.Is(e, unix.ENOENT) {
		return e
	}
	if err = ctx.Err(); err != nil {
		return err
	}
	if err = accountbudget.Provision(root, scope, instance); err != nil {
		return err
	}
	if err = inject("after-provision"); err != nil {
		return err
	}
	if err = ctx.Err(); err != nil {
		return err
	}
	if err = offlineSameRoot(root, rootID); err != nil {
		return err
	}
	b, err := accountbudget.OpenContext(ctx, accountbudget.Config{Root: root, Scope: scope, Instance: instance, Clock: clock.System()})
	if err != nil {
		return err
	}
	event, err := accountbudget.ProvisionObservation(b)
	var queue, wire accountbudget.FileID
	if err == nil {
		queue, err = offlineFileID(rootFD, "queue.lock")
	}
	if err == nil {
		wire, err = offlineFileID(rootFD, "wire.lock")
	}
	ident := accountbudget.Identity{Root: rootID, Queue: queue, Wire: wire}
	if err == nil {
		err = b.VerifyIdentity(ident)
	}
	err = errors.Join(err, b.Close())
	if err != nil {
		return err
	}
	m := manifest{Version: 2, Root: root, Scope: scope, Instance: instance, Identity: ident, EventLock: event}
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	var random [16]byte
	if _, err = rand.Read(random[:]); err != nil {
		return err
	}
	tmp := ".a166-" + hex.EncodeToString(random[:])
	fd, err := unix.Openat(parent, tmp, unix.O_WRONLY|unix.O_CREAT|unix.O_EXCL|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0600)
	if err != nil {
		return err
	}
	file := os.NewFile(uintptr(fd), tmp)
	pending := true
	defer func() {
		if pending {
			if e := unix.Unlinkat(parent, tmp, 0); e != nil && !errors.Is(e, unix.ENOENT) {
				result = errors.Join(result, e)
			}
		}
	}()
	var fileStat unix.Stat_t
	if e = unix.Fstat(fd, &fileStat); e != nil || fileStat.Mode&unix.S_IFMT != unix.S_IFREG || fileStat.Mode&07777 != 0600 || fileStat.Uid != uint32(os.Geteuid()) || fileStat.Nlink != 1 {
		return errors.Join(errors.New("unsafe temporary manifest"), e, file.Close())
	}
	e = inject("write")
	if e == nil {
		var n int
		n, e = file.Write(data)
		if e == nil && n != len(data) {
			e = io.ErrShortWrite
		}
	}
	if e == nil {
		e = inject("file-sync")
	}
	if e == nil {
		e = file.Sync()
	}
	fileCloseErr := errors.Join(inject("close"), file.Close())
	e = errors.Join(e, fileCloseErr)
	if e != nil {
		return e
	}
	if e = ctx.Err(); e != nil {
		return e
	}
	if e = inject("before-final-root"); e != nil {
		return e
	}
	if e = offlineSameRoot(root, rootID); e != nil {
		return e
	}
	if e = inject("rename"); e != nil {
		return e
	}
	if e = unix.Renameat2(parent, tmp, parent, name, unix.RENAME_NOREPLACE); e != nil {
		return e
	}
	pending = false
	if e = inject("dir-sync"); e != nil {
		return e
	}
	if e = unix.Fsync(parent); e != nil {
		return e
	}
	return nil
}
func CheckOffline(ctx context.Context, path string) error {
	return checkOffline(ctx, path, clock.SystemBootSource().Snapshot)
}
func checkOffline(ctx context.Context, path string, snapshot func() (clock.BootStamp, error)) error {
	f, err := Open(ctx, Config{Mode: ModeRequired, Manifest: path, CleanupError: func(error) {}})
	if err != nil {
		return err
	}
	if err = ctx.Err(); err == nil {
		_, err = snapshot()
	}
	if err == nil {
		err = ctx.Err()
	}
	return errors.Join(err, f.Close())
}
