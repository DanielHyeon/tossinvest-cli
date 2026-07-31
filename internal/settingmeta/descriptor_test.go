package settingmeta_test

import (
	"errors"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/settingmeta"
)

func TestFiniteFieldDescriptorAcceptsOnlyRegisteredStableOptions(t *testing.T) {
	d := settingmeta.FieldDescriptor{
		Key: "exit.common-policy", Label: "공통 정책", Description: "새 포지션의 exit 정책",
		Type: settingmeta.TypeEnum, Control: settingmeta.ControlRadioTile,
		Options: []settingmeta.Option{
			{ID: "BALANCED", Label: "균형형"},
			{ID: "RUNNER", Label: "러너형"},
		},
		Default:         settingmeta.State{Kind: settingmeta.StateUnapproved, Display: "기존 RATCHET 유지"},
		Effective:       settingmeta.State{Kind: settingmeta.StateUnapproved, Display: "기존 RATCHET 유지"},
		ApplyTiming:     settingmeta.ApplyNewPositionOnly,
		SafetyDirection: settingmeta.SafetyContextual,
		Provenance:      settingmeta.Provenance{OwnerChange: "a041-complete-exit-line-contract"},
	}
	if err := d.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if err := d.ValidateOption("RUNNER"); err != nil {
		t.Fatalf("registered option: %v", err)
	}
	if err := d.ValidateOption("invented-number"); !errors.Is(err, settingmeta.ErrInvalidDescriptorValue) {
		t.Fatalf("unregistered option error = %v", err)
	}
}

func TestWritableDescriptorFailsClosedWhenMetadataOrFiniteOptionsAreMissing(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*settingmeta.FieldDescriptor)
	}{
		{"description", func(d *settingmeta.FieldDescriptor) { d.Description = "" }},
		{"owner", func(d *settingmeta.FieldDescriptor) { d.Provenance.OwnerChange = "" }},
		{"options", func(d *settingmeta.FieldDescriptor) { d.Options = nil }},
		{"duplicate option", func(d *settingmeta.FieldDescriptor) { d.Options[1].ID = d.Options[0].ID }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			d := settingmeta.FieldDescriptor{
				Key: "x", Label: "X", Description: "X setting", Type: settingmeta.TypeEnum,
				Control:         settingmeta.ControlSelect,
				Options:         []settingmeta.Option{{ID: "A", Label: "A"}, {ID: "B", Label: "B"}},
				Default:         settingmeta.State{Kind: settingmeta.StateValue, OptionID: "A"},
				Effective:       settingmeta.State{Kind: settingmeta.StateValue, OptionID: "A"},
				ApplyTiming:     settingmeta.ApplyNextEvaluation,
				SafetyDirection: settingmeta.SafetyNeutral,
				Provenance:      settingmeta.Provenance{OwnerChange: "a041"},
			}
			tc.mutate(&d)
			if err := d.Validate(); !errors.Is(err, settingmeta.ErrInvalidDescriptor) {
				t.Fatalf("Validate error = %v", err)
			}
		})
	}
}
