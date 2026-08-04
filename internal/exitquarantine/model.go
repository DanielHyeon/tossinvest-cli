// Package exitquarantine is the capability-neutral contract for lifting an exit
// snapshot quarantine. Like internal/positionpolicy it holds no journal, broker,
// config or HTTP capability: it is the shape the engine-owned command service,
// the loopback transport and the operator console all agree on.
//
// # Why this is not a positionpolicy action
//
// positionpolicy answers "which exit policy governs this position, and is it
// under automatic management". A quarantine answers a different question — "can
// the stored exit snapshot be trusted" — and lifting one changes neither the
// lifecycle status, nor the desired policy, nor the adoption generation, nor the
// lifecycle version. Folding it into that CAS pipeline would mean a
// compare-and-swap whose version never moves, handled by a special branch in the
// ledger transaction that commits policy changes. Change a079 design D1.
//
// # What a release actually is
//
// It fills in released_at on one row. The next observation re-runs the very same
// recovery selection that quarantined the position, so a release cannot suppress
// a defect — it can only ask the judgement to be attempted again. A position
// whose underlying ambiguity still holds is quarantined again immediately, under
// a new version.
package exitquarantine

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// ActorLocalOperator is the only actor this contract accepts. A release is a
// §0.7 human decision; there is deliberately no enum value for an automatic one.
const ActorLocalOperator = "LOCAL_OPERATOR"

var (
	ErrInvalidRequest    = errors.New("exit quarantine: invalid request")
	ErrNotQuarantined    = errors.New("exit quarantine: position generation has no active quarantine")
	ErrVersionMismatch   = errors.New("exit quarantine: quarantine version mismatch")
	ErrCapabilityInvalid = errors.New("exit quarantine: release capability is invalid or consumed")
	// ErrCapabilityTooEarly is the mandatory danger delay, not a rate limit. The
	// operator has to have had time to read the screen.
	ErrCapabilityTooEarly   = errors.New("exit quarantine: release capability is not active yet")
	ErrCapabilityExpired    = errors.New("exit quarantine: release capability expired")
	ErrConfirmationRequired = errors.New("exit quarantine: release requires the danger acknowledgement")
	// ErrUnwired means this build reached a control plane that does not offer the
	// quarantine capability at all — an engine older than a079, or a console with
	// no command seam. It is not a refusal of the request; nothing was attempted.
	ErrUnwired = errors.New("exit quarantine: the engine control plane does not offer quarantine release")
)

// Row is one active quarantine, joined with enough position identity for an
// operator to recognise it. It carries no account reference: the console already
// knows whose console it is, and the audit record lives in the ledger.
type Row struct {
	PositionID string `json:"position_id"`
	Market     string `json:"market"`
	Symbol     string `json:"symbol"`
	// Generation is the position instance_seq the quarantine was written against.
	// A quarantine from an earlier generation is not active for the current one.
	Generation    int64  `json:"generation"`
	Version       int64  `json:"version"`
	Reason        string `json:"reason"`
	Evidence      string `json:"evidence"`
	QuarantinedAt string `json:"quarantined_at"`
	// Protection is the protection line the stored snapshot keeps, as a decimal
	// string, or empty when the stored snapshot cannot be read. It is display
	// evidence for the operator's decision and grants nothing.
	Protection string `json:"protection,omitempty"`
	// ProtectionUnknown says why Protection is empty, so the screen can be honest
	// rather than showing a blank cell.
	ProtectionUnknown string `json:"protection_unknown,omitempty"`
}

// Request is the operator's intent to lift one specific quarantine.
type Request struct {
	PositionID string    `json:"position_id"`
	Generation int64     `json:"generation"`
	Version    int64     `json:"version"`
	Actor      string    `json:"actor"`
	At         time.Time `json:"at"`
}

// Validate rejects a request the engine must not even look up.
//
// Version must be positive: quarantine versions start at 1, so a zero here is a
// caller that never read a real row, and letting it through would turn a
// compare-and-swap into a blind write.
func (r Request) Validate() error {
	if strings.TrimSpace(r.PositionID) == "" {
		return fmt.Errorf("%w: position id is required", ErrInvalidRequest)
	}
	if r.Generation < 0 {
		return fmt.Errorf("%w: generation must not be negative", ErrInvalidRequest)
	}
	if r.Version <= 0 {
		return fmt.Errorf("%w: quarantine version must be positive", ErrInvalidRequest)
	}
	if r.Actor != ActorLocalOperator {
		return fmt.Errorf("%w: actor is not the local operator enum", ErrInvalidRequest)
	}
	return nil
}

// Preview is what the operator is shown before approving, plus the one-time
// grant that Release consumes.
type Preview struct {
	Row Row `json:"row"`
	// Capability is an engine-instance-local, one-time opaque grant. Its bytes
	// reveal no scope and Release accepts no replacement scope fields.
	Capability string `json:"capability,omitempty"`
	// WaitSeconds is the danger delay the console must honour before enabling
	// its apply button. The engine enforces the same delay regardless.
	WaitSeconds int `json:"wait_seconds"`
}

// ApplyRequest carries the grant and the danger acknowledgement, and nothing
// else. There is no scope here for a browser to substitute.
type ApplyRequest struct {
	Capability string `json:"capability"`
	Confirmed  bool   `json:"confirmed"`
}

// Result is the committed release.
type Result struct {
	Row        Row    `json:"row"`
	ReleasedAt string `json:"released_at"`
	// Evidence is the string the server composed and wrote to the ledger. It is
	// returned so the console can show exactly what was recorded rather than
	// re-deriving it.
	Evidence string `json:"evidence"`
}

// ComposeEvidence builds the ledger evidence for a human release.
//
// The operator is never asked to type this. A typed confirmation phrase is
// friction the user has explicitly ruled out of this console, and it would be
// less accurate than the fields the server already holds: who, which version,
// what the recorded reason was, and when. The danger acknowledgement and the
// mandatory delay carry the "are you sure", which is what they are for.
func ComposeEvidence(actor string, row Row, at time.Time) string {
	actor = strings.TrimSpace(actor)
	if actor == "" {
		actor = ActorLocalOperator
	}
	reason := strings.TrimSpace(row.Reason)
	if reason == "" {
		reason = "unknown"
	}
	return fmt.Sprintf("%s released quarantine v%d (reason=%s) from the console at %s",
		actor, row.Version, reason, at.UTC().Format(time.RFC3339))
}
