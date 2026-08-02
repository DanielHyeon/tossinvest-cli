//go:build !unix

package engine

import (
	"github.com/JungHoonGhae/tossinvest-cli/internal/positionpolicy"
)

type PositionPolicyRuntimeServer struct{}

func StartPositionPolicyRuntimeServer(string, positionpolicy.RuntimeReader) (*PositionPolicyRuntimeServer, error) {
	// The private API sidecar is deployed on Unix. Other release targets keep
	// engine operation available and simply expose runtime as unknown to readers.
	return &PositionPolicyRuntimeServer{}, nil
}

func (*PositionPolicyRuntimeServer) Close() error { return nil }
