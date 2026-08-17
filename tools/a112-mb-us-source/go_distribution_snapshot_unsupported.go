//go:build !unix

package main

import (
	"context"
	"errors"
)

func snapshotGoDistribution(string, string) (*goDistributionSnapshot, error) {
	return nil, errors.New("private Go distribution snapshots are unsupported on this platform")
}

func snapshotGoDistributionContext(context.Context, string, string) (*goDistributionSnapshot, error) {
	return nil, errors.New("private Go distribution snapshots are unsupported on this platform")
}

func validateGoDistributionSnapshot(*goDistributionSnapshot) error {
	return errors.New("private Go distribution snapshots are unsupported on this platform")
}

func closeGoDistributionSnapshot(*goDistributionSnapshot) error { return nil }
