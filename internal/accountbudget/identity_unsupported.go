//go:build !linux

package accountbudget

type FileID struct {
	Device uint64 `json:"device"`
	Inode  uint64 `json:"inode"`
}
type Identity struct {
	Root  FileID `json:"root"`
	Queue FileID `json:"queue"`
	Wire  FileID `json:"wire"`
}

func (*Budget) VerifyIdentity(Identity) error { return ErrUnavailable }
