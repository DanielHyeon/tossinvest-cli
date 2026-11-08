// Package shadowsession stores a daily shadow-collection session without binding it to a transport.
package shadowsession

import "errors"

var (
	ErrClosed  = errors.New("shadowsession: closed")
	ErrBusy    = errors.New("shadowsession: busy")
	ErrInvalid = errors.New("shadowsession: invalid or unavailable state")
)

type Root struct {
	Path   string
	Device uint64
	Inode  uint64
}
