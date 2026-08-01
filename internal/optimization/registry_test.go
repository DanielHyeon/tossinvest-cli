package optimization_test

import (
	"context"
	"errors"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/optimization"
	"github.com/JungHoonGhae/tossinvest-cli/internal/settingmeta"
)

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
