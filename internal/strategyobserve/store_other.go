//go:build !linux

package strategyobserve

import "errors"

type Store struct{}

func Open(string) (*Store, error) {
	return nil, errors.New("offline observation storage unsupported on this platform")
}
func (*Store) Append([]byte) (bool, error) { return false, errors.New("unsupported") }
func (*Store) Close() error                { return nil }
