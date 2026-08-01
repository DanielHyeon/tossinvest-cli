package optimization_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/optimization"
	"github.com/JungHoonGhae/tossinvest-cli/internal/settingmeta"
)

type failingProvider struct{ owner string }

func (p failingProvider) OwnerChange() string { return p.owner }
func (p failingProvider) Descriptors(context.Context) ([]settingmeta.FieldDescriptor, error) {
	return nil, fmt.Errorf("descriptor source unavailable")
}

func descriptor(owner, key string) settingmeta.FieldDescriptor {
	return settingmeta.FieldDescriptor{
		Key: key, Label: "공통 정책", Description: "신규 포지션에 적용할 보호 정책",
		Type: settingmeta.TypeEnum, Control: settingmeta.ControlRadioTile,
		Options:     []settingmeta.Option{{ID: "SAFE", Label: "보수형"}, {ID: "BALANCED", Label: "균형형"}},
		Default:     settingmeta.State{Kind: settingmeta.StateUnapproved, Display: "미승인"},
		Effective:   settingmeta.State{Kind: settingmeta.StateUnapproved, Display: "기존 보호 유지"},
		ApplyTiming: settingmeta.ApplyNewPositionOnly, SafetyDirection: settingmeta.SafetyContextual,
		Provenance: settingmeta.Provenance{OwnerChange: owner, PolicyID: "policy", PolicyVersion: "v1"},
	}
}

func TestRegistryRequiresExactlyOneMatchingOwnerForEveryField(t *testing.T) {
	ctx := context.Background()
	one := optimization.StaticProvider{Owner: "a041", Fields: []settingmeta.FieldDescriptor{descriptor("a041", "exit.common-policy")}}
	registry, err := optimization.BuildRegistry(ctx, optimization.ProviderBinding{
		Category: optimization.CategoryExitProtection, Provider: one,
	})
	if err != nil {
		t.Fatalf("BuildRegistry: %v", err)
	}
	field, ok := registry.Field("exit.common-policy")
	if !ok || field.Category != optimization.CategoryExitProtection {
		t.Fatalf("registered field = %+v, %v", field, ok)
	}

	_, err = optimization.BuildRegistry(ctx,
		optimization.ProviderBinding{Category: optimization.CategoryExitProtection, Provider: one},
		optimization.ProviderBinding{Category: optimization.CategoryPositionManagement, Provider: one})
	if !errors.Is(err, optimization.ErrInvalidRegistry) {
		t.Fatalf("duplicate owner error = %v", err)
	}

	liar := optimization.StaticProvider{Owner: "a044", Fields: []settingmeta.FieldDescriptor{descriptor("a041", "adoption.enabled")}}
	if _, err := optimization.BuildRegistry(ctx, optimization.ProviderBinding{
		Category: optimization.CategoryPositionManagement, Provider: liar,
	}); !errors.Is(err, optimization.ErrInvalidRegistry) {
		t.Fatalf("mismatched owner error = %v", err)
	}
}

func TestRegistryReturnsCopiesInsteadOfMutableOwnerMetadata(t *testing.T) {
	provider := optimization.StaticProvider{Owner: "a041", Fields: []settingmeta.FieldDescriptor{descriptor("a041", "exit.common-policy")}}
	registry, err := optimization.BuildRegistry(context.Background(), optimization.ProviderBinding{
		Category: optimization.CategoryExitProtection, Provider: provider,
	})
	if err != nil {
		t.Fatal(err)
	}
	first, _ := registry.Field("exit.common-policy")
	first.Descriptor.Options[0].Label = "변조"
	second, _ := registry.Field("exit.common-policy")
	if second.Descriptor.Options[0].Label != "보수형" {
		t.Fatalf("registry metadata was mutable: %+v", second.Descriptor.Options)
	}
}

func TestCategoriesAreTheFixedSharedOrder(t *testing.T) {
	want := []optimization.Category{
		optimization.CategoryOverview, optimization.CategoryExitProtection,
		optimization.CategoryPositionManagement, optimization.CategoryCandidateFilters,
		optimization.CategoryStrategyRuntime, optimization.CategoryPerformanceHistory,
	}
	got := optimization.Categories()
	if len(got) != len(want) {
		t.Fatalf("categories = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i].ID != want[i] {
			t.Errorf("category %d = %q, want %q", i, got[i].ID, want[i])
		}
	}
}

func TestRegistryRejectsNilEmptyAndFailingProviders(t *testing.T) {
	ctx := context.Background()
	for _, tc := range []struct {
		name    string
		binding optimization.ProviderBinding
	}{
		{name: "nil", binding: optimization.ProviderBinding{Category: optimization.CategoryExitProtection}},
		{name: "invalid category", binding: optimization.ProviderBinding{Category: optimization.CategoryOverview,
			Provider: optimization.StaticProvider{Owner: "a041", Fields: []settingmeta.FieldDescriptor{descriptor("a041", "exit.common-policy")}}}},
		{name: "blank owner", binding: optimization.ProviderBinding{Category: optimization.CategoryExitProtection,
			Provider: optimization.StaticProvider{Fields: []settingmeta.FieldDescriptor{descriptor("", "exit.common-policy")}}}},
		{name: "empty descriptors", binding: optimization.ProviderBinding{Category: optimization.CategoryExitProtection,
			Provider: optimization.StaticProvider{Owner: "a041"}}},
		{name: "provider error", binding: optimization.ProviderBinding{Category: optimization.CategoryExitProtection,
			Provider: failingProvider{owner: "a041"}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := optimization.BuildRegistry(ctx, tc.binding); !errors.Is(err, optimization.ErrInvalidRegistry) {
				t.Fatalf("BuildRegistry error = %v", err)
			}
		})
	}
}

func TestRequiredRegistryCoverageRejectsMissingUnexpectedAndDriftedFields(t *testing.T) {
	ctx := context.Background()
	required := []optimization.FieldRequirement{{
		Key: "exit.common-policy", Owner: "a041", Category: optimization.CategoryExitProtection,
	}}
	valid := optimization.ProviderBinding{Category: optimization.CategoryExitProtection,
		Provider: optimization.StaticProvider{Owner: "a041", Fields: []settingmeta.FieldDescriptor{
			descriptor("a041", "exit.common-policy"),
		}}}
	if _, err := optimization.BuildRequiredRegistry(ctx, required, valid); err != nil {
		t.Fatalf("valid required registry: %v", err)
	}
	for _, tc := range []struct {
		name     string
		required []optimization.FieldRequirement
		bindings []optimization.ProviderBinding
	}{
		{name: "empty manifest"},
		{name: "blank key", required: []optimization.FieldRequirement{{
			Owner: "a041", Category: optimization.CategoryExitProtection,
		}}, bindings: []optimization.ProviderBinding{valid}},
		{name: "blank owner", required: []optimization.FieldRequirement{{
			Key: "exit.common-policy", Category: optimization.CategoryExitProtection,
		}}, bindings: []optimization.ProviderBinding{valid}},
		{name: "missing", required: required},
		{name: "unexpected", required: required, bindings: []optimization.ProviderBinding{{
			Category: optimization.CategoryExitProtection,
			Provider: optimization.StaticProvider{Owner: "a041", Fields: []settingmeta.FieldDescriptor{
				descriptor("a041", "exit.common-policy"), descriptor("a041", "exit.unapproved-extra"),
			}},
		}}},
		{name: "wrong owner", required: []optimization.FieldRequirement{{
			Key: "exit.common-policy", Owner: "a044", Category: optimization.CategoryExitProtection,
		}}, bindings: []optimization.ProviderBinding{valid}},
		{name: "wrong category", required: []optimization.FieldRequirement{{
			Key: "exit.common-policy", Owner: "a041", Category: optimization.CategoryPositionManagement,
		}}, bindings: []optimization.ProviderBinding{valid}},
		{name: "duplicate requirement", required: append(required, required[0]), bindings: []optimization.ProviderBinding{valid}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := optimization.BuildRequiredRegistry(ctx, tc.required, tc.bindings...); !errors.Is(err, optimization.ErrInvalidRegistry) {
				t.Fatalf("BuildRequiredRegistry error = %v", err)
			}
		})
	}
}

func TestCoreRegistryCoversExactApprovedSettingmetaManifest(t *testing.T) {
	registry, err := optimization.CoreRegistry(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	fields := registry.All()
	if len(fields) != 1 || fields[0].Descriptor.Key != "exit.common-policy" ||
		fields[0].Descriptor.Provenance.OwnerChange != "a041-complete-exit-line-contract" ||
		fields[0].Category != optimization.CategoryExitProtection {
		t.Fatalf("core registry = %+v", fields)
	}
}

func TestIncompleteKnownDescriptorIsExposedReadOnlyWithConfigurationError(t *testing.T) {
	broken := descriptor("a041", "exit.common-policy")
	broken.Description = ""
	broken.ApplyTiming = ""
	registry, err := optimization.BuildRegistry(context.Background(), optimization.ProviderBinding{
		Category: optimization.CategoryExitProtection,
		Provider: optimization.StaticProvider{Owner: "a041", Fields: []settingmeta.FieldDescriptor{broken}},
	})
	if err != nil {
		t.Fatalf("safe known field should remain visible: %v", err)
	}
	field, ok := registry.Field("exit.common-policy")
	if !ok || field.Descriptor.Control != settingmeta.ControlReadOnly || field.ConfigurationError == "" ||
		len(field.Descriptor.Options) != 0 {
		t.Fatalf("configuration error field = %+v, ok=%v", field, ok)
	}

	broken.Key = ""
	if _, err := optimization.BuildRegistry(context.Background(), optimization.ProviderBinding{
		Category: optimization.CategoryExitProtection,
		Provider: optimization.StaticProvider{Owner: "a041", Fields: []settingmeta.FieldDescriptor{broken}},
	}); !errors.Is(err, optimization.ErrInvalidRegistry) {
		t.Fatalf("descriptor without stable identity error = %v", err)
	}
}
