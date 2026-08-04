package positioncampaign

import (
	"errors"
	"fmt"
)

var ErrInvalidCommand = errors.New("position campaign: invalid command identity")

type CommandKind string

const (
	CommandCreate            CommandKind = "CREATE"
	CommandPlanLeg           CommandKind = "PLAN_LEG"
	CommandLinkOrder         CommandKind = "LINK_ORDER"
	CommandCancelProspective CommandKind = "CANCEL_PROSPECTIVE"
	CommandApplyFill         CommandKind = "APPLY_FILL"
	CommandUpdateStop        CommandKind = "UPDATE_STOP"
	CommandRecordEvidence    CommandKind = "RECORD_EVIDENCE"
)

func ValidateCommand(kind, key string) error {
	switch CommandKind(kind) {
	case CommandCreate, CommandPlanLeg, CommandLinkOrder, CommandCancelProspective,
		CommandApplyFill, CommandUpdateStop, CommandRecordEvidence:
	default:
		return fmt.Errorf("%w: unknown kind %q", ErrInvalidCommand, kind)
	}
	if len(key) == 0 || len(key) > 128 {
		return fmt.Errorf("%w: key length %d is outside 1..128", ErrInvalidCommand, len(key))
	}
	for i := 0; i < len(key); i++ {
		c := key[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') || c == '.' || c == '_' || c == ':' || c == '-' {
			continue
		}
		return fmt.Errorf("%w: key %q contains a non-canonical byte at %d", ErrInvalidCommand, key, i)
	}
	return nil
}

// TypedCommandIdentity uses length-prefixed components, so even future key
// syntax changes cannot make (kind,key) pairs concatenate ambiguously.
func TypedCommandIdentity(kind, key string) (string, error) {
	if err := ValidateCommand(kind, key); err != nil {
		return "", err
	}
	return fmt.Sprintf("%d:%s%d:%s", len(kind), kind, len(key), key), nil
}
