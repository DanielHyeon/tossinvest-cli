//go:build linux

package accountbudget

import "errors"

// FileID identifies a pinned filesystem object, not an account.
type FileID struct {
	Device uint64 `json:"device"`
	Inode  uint64 `json:"inode"`
}
type Identity struct {
	Root  FileID `json:"root"`
	Queue FileID `json:"queue"`
	Wire  FileID `json:"wire"`
}

// VerifyIdentity compares externally trusted IDs with this Budget's pinned files.
func (b *Budget) VerifyIdentity(expected Identity) error {
	if b == nil {
		return ErrUnavailable
	}
	if err := b.verify(); err != nil {
		return errors.Join(ErrUnavailable, err)
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		return ErrUnavailable
	}
	match := func(dev, ino uint64, id FileID) bool {
		return id.Device != 0 && id.Inode != 0 && dev == id.Device && ino == id.Inode
	}
	q, w := b.locks["queue.lock"], b.locks["wire.lock"]
	if !match(uint64(b.dirStat.Dev), b.dirStat.Ino, expected.Root) || !match(uint64(q.Dev), q.Ino, expected.Queue) || !match(uint64(w.Dev), w.Ino, expected.Wire) {
		return ErrUnavailable
	}
	return nil
}
