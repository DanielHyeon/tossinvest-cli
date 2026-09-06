//go:build !linux && !darwin

package soak

import (
	"errors"
	"os"
)

// 진단은 advisory다. no-follow primitive가 없는 플랫폼에서는 소유권과 교체 계약을
// 약화하지 않고 읽기를 거부해 unknown으로 보인다.
func openRenewalStatus(string) (*os.File, error) {
	return nil, errors.New("renewal status secure open unsupported")
}

func renewalStatusOwnedByCurrentUser(os.FileInfo) bool { return false }

func renewalStatusReaderSupported() bool { return false }
