package engine

// alert_control.go names the alert operator endpoint. The names live apart from
// the socket so the CLI can resolve the descriptor on a platform where the server
// itself does not build.

import (
	"path/filepath"
	"strings"
)

const (
	// alertControlDirectoryName is the endpoint's own directory. It is not the
	// position-policy one: two endpoints sharing a directory share the blast
	// radius of whoever can read it, and these two grant different powers.
	alertControlDirectoryName      = ".alert-control"
	alertControlDescriptorFileName = "endpoint.json"
	alertControlSocketFileName     = "alerts.sock"

	// AlertControlListPath returns the undelivered critical alerts.
	AlertControlListPath = "/v1/alerts"
	// AlertControlAcknowledgePath records that a person has seen them.
	AlertControlAcknowledgePath = "/v1/alerts/acknowledge"
)

// AlertControlDescriptor is how a local client finds the socket and proves it may
// use it.
//
// Socket is a bare filename rather than a path, following the runtime endpoint's
// reasoning: a descriptor that carried an absolute path would let whoever could
// write it point a client somewhere else.
type AlertControlDescriptor struct {
	Socket string `json:"socket"`
	Token  string `json:"token"`
	PID    int    `json:"pid"`
}

// AcknowledgeRequest is the acknowledge route's body.
//
// Operator is required and has no default anywhere in this package. The ledger
// refuses a blank name (outbox.go:482), and that refusal is the audit trail's only
// guard — a default here would satisfy it with a value no person chose.
type AcknowledgeRequest struct {
	IDs      []int64 `json:"ids,omitempty"`
	Operator string  `json:"operator"`
}

func AlertControlDirectory(engineDir string) string {
	return filepath.Join(strings.TrimSpace(engineDir), alertControlDirectoryName)
}

func AlertControlDescriptorPath(engineDir string) string {
	return filepath.Join(AlertControlDirectory(engineDir), alertControlDescriptorFileName)
}

func AlertControlSocketPath(engineDir string) string {
	return filepath.Join(AlertControlDirectory(engineDir), alertControlSocketFileName)
}

// AlertControlSocketFileName is what a descriptor's Socket field must equal. A
// client that accepts any name accepts a descriptor pointing at another endpoint.
func AlertControlSocketFileName() string { return alertControlSocketFileName }
