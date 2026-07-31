// Package settingmeta defines transport-neutral metadata for finite,
// server-authoritative settings. It intentionally knows nothing about HTML,
// JSON, HTTP, persistence, or the component that will render a descriptor.
package settingmeta

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var (
	ErrInvalidDescriptor      = errors.New("settingmeta: invalid field descriptor")
	ErrInvalidDescriptorValue = errors.New("settingmeta: value is not a registered option")
)

type ValueType string

const (
	TypeBoolean    ValueType = "boolean"
	TypeDecimal    ValueType = "decimal"
	TypeInteger    ValueType = "integer"
	TypeEnum       ValueType = "enum"
	TypeSymbolList ValueType = "symbol-list"
)

func (v ValueType) valid() bool {
	switch v {
	case TypeBoolean, TypeDecimal, TypeInteger, TypeEnum, TypeSymbolList:
		return true
	default:
		return false
	}
}

type ControlKind string

const (
	ControlRadioTile    ControlKind = "radio-tile"
	ControlSelect       ControlKind = "select"
	ControlChip         ControlKind = "chip"
	ControlToggle       ControlKind = "toggle"
	ControlDiscreteStep ControlKind = "discrete-step"
	ControlRowAction    ControlKind = "row-action"
	ControlReadOnly     ControlKind = "read-only"
)

func (c ControlKind) valid() bool {
	switch c {
	case ControlRadioTile, ControlSelect, ControlChip, ControlToggle,
		ControlDiscreteStep, ControlRowAction, ControlReadOnly:
		return true
	default:
		return false
	}
}

type StateKind string

const (
	StateValue         StateKind = "value"
	StateUnapproved    StateKind = "unapproved"
	StateNotApplicable StateKind = "not-applicable"
)

type State struct {
	Kind     StateKind
	OptionID string
	Display  string
}

type ApplyTiming string

const (
	ApplyImmediate       ApplyTiming = "immediate"
	ApplyNextEvaluation  ApplyTiming = "next-evaluation"
	ApplyNextEngineStart ApplyTiming = "next-engine-start"
	ApplyNewPositionOnly ApplyTiming = "new-position-only"
)

func (a ApplyTiming) valid() bool {
	switch a {
	case ApplyImmediate, ApplyNextEvaluation, ApplyNextEngineStart, ApplyNewPositionOnly:
		return true
	default:
		return false
	}
}

type SafetyDirection string

const (
	SafetySaferWhenHigher SafetyDirection = "safer-when-higher"
	SafetySaferWhenLower  SafetyDirection = "safer-when-lower"
	SafetyNeutral         SafetyDirection = "neutral"
	SafetyContextual      SafetyDirection = "contextual"
)

func (s SafetyDirection) valid() bool {
	switch s {
	case SafetySaferWhenHigher, SafetySaferWhenLower, SafetyNeutral, SafetyContextual:
		return true
	default:
		return false
	}
}

// Option is one stable finite choice. Value is a canonical domain value; ID is
// the only value a transport is allowed to submit.
type Option struct {
	ID          string
	Label       string
	Description string
	Value       string
	Recommended bool
}

type Provenance struct {
	OwnerChange    string
	PolicyID       string
	PolicyVersion  string
	PolicyDigest   string
	EvidenceDigest string
}

// FieldDescriptor is deliberately finite. Even decimal and integer settings
// are represented by approved options rather than arbitrary transport input.
type FieldDescriptor struct {
	Key             string
	Label           string
	Description     string
	Type            ValueType
	Unit            string
	Control         ControlKind
	Options         []Option
	Default         State
	Effective       State
	ApplyTiming     ApplyTiming
	SafetyDirection SafetyDirection
	Provenance      Provenance
}

var stableID = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:-]*$`)

func (d FieldDescriptor) Validate() error {
	if strings.TrimSpace(d.Key) == "" || strings.TrimSpace(d.Label) == "" ||
		strings.TrimSpace(d.Description) == "" {
		return fmt.Errorf("%w: key, label, and description are required", ErrInvalidDescriptor)
	}
	if !d.Type.valid() || !d.Control.valid() || !d.ApplyTiming.valid() || !d.SafetyDirection.valid() {
		return fmt.Errorf("%w: type, control, apply timing, and safety direction must be known", ErrInvalidDescriptor)
	}
	if strings.TrimSpace(d.Provenance.OwnerChange) == "" {
		return fmt.Errorf("%w: owner change is required", ErrInvalidDescriptor)
	}
	if d.Control != ControlReadOnly && len(d.Options) == 0 {
		return fmt.Errorf("%w: writable fields require finite options", ErrInvalidDescriptor)
	}
	options := make(map[string]struct{}, len(d.Options))
	for _, option := range d.Options {
		id := strings.TrimSpace(option.ID)
		if !stableID.MatchString(id) || strings.TrimSpace(option.Label) == "" {
			return fmt.Errorf("%w: every option needs a stable id and label", ErrInvalidDescriptor)
		}
		if _, exists := options[id]; exists {
			return fmt.Errorf("%w: duplicate option id %q", ErrInvalidDescriptor, id)
		}
		options[id] = struct{}{}
	}
	if err := validateState("default", d.Default, options); err != nil {
		return err
	}
	if err := validateState("effective", d.Effective, options); err != nil {
		return err
	}
	return nil
}

func validateState(name string, state State, options map[string]struct{}) error {
	switch state.Kind {
	case StateValue:
		if _, ok := options[strings.TrimSpace(state.OptionID)]; !ok {
			return fmt.Errorf("%w: %s state option %q is not registered", ErrInvalidDescriptor, name, state.OptionID)
		}
	case StateUnapproved, StateNotApplicable:
		if strings.TrimSpace(state.OptionID) != "" || strings.TrimSpace(state.Display) == "" {
			return fmt.Errorf("%w: %s non-value state needs display text and no option", ErrInvalidDescriptor, name)
		}
	default:
		return fmt.Errorf("%w: %s state kind %q is unknown", ErrInvalidDescriptor, name, state.Kind)
	}
	return nil
}

func (d FieldDescriptor) ValidateOption(id string) error {
	if err := d.Validate(); err != nil {
		return err
	}
	id = strings.TrimSpace(id)
	for _, option := range d.Options {
		if option.ID == id {
			return nil
		}
	}
	return fmt.Errorf("%w: field %s does not register %q", ErrInvalidDescriptorValue, d.Key, id)
}
