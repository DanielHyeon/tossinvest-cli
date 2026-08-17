//go:build !unix

package main

import "errors"

func snapshotExecutionBinary(string) (*executionSnapshot, error) {
	return nil, errors.New("private execution snapshots are unsupported on this platform")
}

func validateExecutionSnapshot(*executionSnapshot) error {
	return errors.New("private execution snapshots are unsupported on this platform")
}

func closeExecutionSnapshot(*executionSnapshot) error { return nil }
