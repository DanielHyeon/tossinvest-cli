package main

import (
	"errors"
	"sync"
)

const (
	maxGoDistributionBytes   int64 = 512 << 20
	maxGoDistributionEntries       = 50_000
)

// goDistributionSnapshot is the private GOROOT capability paired with the
// private selected-Go executable.  Its manifest is intentionally deterministic
// so the receipt binds the exact source/tool distribution that Go may load.
type goDistributionSnapshot struct {
	sourceRoot  string
	privateRoot string
	digest      string
	rootFD      int
	entries     map[string]goDistributionEntry
	mu          sync.Mutex
	closed      bool
}

type goDistributionEntry struct {
	path   string
	mode   uint32
	size   int64
	digest string
	dir    bool
}

func (s *goDistributionSnapshot) rootPath() (string, error) {
	if s == nil {
		return "", errors.New("private Go distribution snapshot is unavailable")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return "", errors.New("private Go distribution snapshot is closed")
	}
	if err := validateGoDistributionSnapshot(s); err != nil {
		return "", err
	}
	return s.privateRoot, nil
}

func (s *goDistributionSnapshot) close() error {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil
	}
	s.closed = true
	return closeGoDistributionSnapshot(s)
}
