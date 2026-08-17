package main

import (
	"errors"
	"sync"
)

// executionSnapshot is a short-lived execution capability.  sourcePath is
// retained only as receipt evidence; command execution is allowed exclusively
// through executionPath, which names a private exact-byte copy.
type executionSnapshot struct {
	sourcePath    string
	executionPath string
	directory     string
	name          string
	digest        string
	directoryFD   int
	fileFD        int
	mu            sync.Mutex
	closed        bool
}

func (s *executionSnapshot) commandPath() (string, error) {
	if s == nil {
		return "", errors.New("execution snapshot is unavailable")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return "", errors.New("execution snapshot is closed")
	}
	if err := validateExecutionSnapshot(s); err != nil {
		return "", err
	}
	return s.executionPath, nil
}

func (s *executionSnapshot) close() error {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil
	}
	s.closed = true
	return closeExecutionSnapshot(s)
}
