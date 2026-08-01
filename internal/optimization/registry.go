package optimization

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/settingmeta"
)

type RegisteredField struct {
	Category           Category
	Descriptor         settingmeta.FieldDescriptor
	ConfigurationError string
}

type Registry struct {
	fields map[string]RegisteredField
}

// FieldRequirement is the frozen composition manifest for writable
// settingmeta providers present at a release HEAD. Owner-specific read-only
// projections that do not implement settingmeta are intentionally not listed.
type FieldRequirement struct {
	Key      string
	Owner    string
	Category Category
}

func BuildRegistry(ctx context.Context, bindings ...ProviderBinding) (Registry, error) {
	fields := make(map[string]RegisteredField)
	owners := make(map[string]struct{})
	for index, binding := range bindings {
		if binding.Provider == nil {
			return Registry{}, fmt.Errorf("%w: binding %d has no provider", ErrInvalidRegistry, index)
		}
		if _, ok := ParseCategory(string(binding.Category)); !ok || binding.Category == CategoryOverview ||
			binding.Category == CategoryPerformanceHistory {
			return Registry{}, fmt.Errorf("%w: binding %d has non-setting category %q", ErrInvalidRegistry, index, binding.Category)
		}
		owner := strings.TrimSpace(binding.Provider.OwnerChange())
		if owner == "" {
			return Registry{}, fmt.Errorf("%w: binding %d has no owner", ErrInvalidRegistry, index)
		}
		if _, duplicate := owners[owner]; duplicate {
			return Registry{}, fmt.Errorf("%w: owner %q has more than one provider", ErrInvalidRegistry, owner)
		}
		owners[owner] = struct{}{}
		descriptors, err := binding.Provider.Descriptors(ctx)
		if err != nil {
			return Registry{}, fmt.Errorf("%w: owner %s: %v", ErrInvalidRegistry, owner, err)
		}
		if len(descriptors) == 0 {
			return Registry{}, fmt.Errorf("%w: owner %s has no descriptors", ErrInvalidRegistry, owner)
		}
		for _, descriptor := range descriptors {
			key := strings.TrimSpace(descriptor.Key)
			if key == "" {
				return Registry{}, fmt.Errorf("%w: owner %s field has no stable key", ErrInvalidRegistry, owner)
			}
			if descriptor.Provenance.OwnerChange != owner {
				return Registry{}, fmt.Errorf("%w: field %q claims owner %q, provider is %q",
					ErrInvalidRegistry, descriptor.Key, descriptor.Provenance.OwnerChange, owner)
			}
			if _, duplicate := fields[key]; duplicate {
				return Registry{}, fmt.Errorf("%w: field %q has more than one owner", ErrInvalidRegistry, key)
			}
			descriptor.Key = key
			registered := RegisteredField{Category: binding.Category, Descriptor: cloneDescriptor(descriptor)}
			if validationErr := descriptor.Validate(); validationErr != nil {
				registered.ConfigurationError = validationErr.Error()
				registered.Descriptor.Control = settingmeta.ControlReadOnly
				registered.Descriptor.Options = nil
			}
			fields[key] = registered
		}
	}
	return Registry{fields: fields}, nil
}

// BuildRequiredRegistry combines the ordinary exact-one-owner validation with
// an exact key/owner/category coverage manifest. It rejects both omissions and
// surprise writable fields, so a provider integration cannot silently widen or
// narrow the control surface.
func BuildRequiredRegistry(ctx context.Context, required []FieldRequirement, bindings ...ProviderBinding) (Registry, error) {
	if len(required) == 0 {
		return Registry{}, fmt.Errorf("%w: required field manifest is empty", ErrInvalidRegistry)
	}
	want := make(map[string]FieldRequirement, len(required))
	for _, requirement := range required {
		requirement.Key = strings.TrimSpace(requirement.Key)
		requirement.Owner = strings.TrimSpace(requirement.Owner)
		if requirement.Key == "" || requirement.Owner == "" {
			return Registry{}, fmt.Errorf("%w: required field key and owner are required", ErrInvalidRegistry)
		}
		if _, duplicate := want[requirement.Key]; duplicate {
			return Registry{}, fmt.Errorf("%w: field %q is required more than once", ErrInvalidRegistry, requirement.Key)
		}
		want[requirement.Key] = requirement
	}
	registry, err := BuildRegistry(ctx, bindings...)
	if err != nil {
		return Registry{}, err
	}
	if len(registry.fields) != len(want) {
		return Registry{}, fmt.Errorf("%w: registry has %d fields, manifest requires %d", ErrInvalidRegistry, len(registry.fields), len(want))
	}
	for key, requirement := range want {
		field, ok := registry.fields[key]
		if !ok || field.Category != requirement.Category ||
			strings.TrimSpace(field.Descriptor.Provenance.OwnerChange) != requirement.Owner {
			return Registry{}, fmt.Errorf("%w: required field %q owner/category coverage mismatch", ErrInvalidRegistry, key)
		}
	}
	return registry, nil
}

func (r Registry) Field(key string) (RegisteredField, bool) {
	field, ok := r.fields[strings.TrimSpace(key)]
	if !ok {
		return RegisteredField{}, false
	}
	field.Descriptor = cloneDescriptor(field.Descriptor)
	return field, true
}

func (r Registry) Fields(category Category) []RegisteredField {
	keys := make([]string, 0, len(r.fields))
	for key, field := range r.fields {
		if field.Category == category {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	out := make([]RegisteredField, 0, len(keys))
	for _, key := range keys {
		field := r.fields[key]
		field.Descriptor = cloneDescriptor(field.Descriptor)
		out = append(out, field)
	}
	return out
}

func (r Registry) All() []RegisteredField {
	keys := make([]string, 0, len(r.fields))
	for key := range r.fields {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	out := make([]RegisteredField, 0, len(keys))
	for _, key := range keys {
		field := r.fields[key]
		field.Descriptor = cloneDescriptor(field.Descriptor)
		out = append(out, field)
	}
	return out
}

func cloneDescriptor(in settingmeta.FieldDescriptor) settingmeta.FieldDescriptor {
	out := in
	out.Options = append([]settingmeta.Option(nil), in.Options...)
	return out
}

type StaticProvider struct {
	Owner  string
	Fields []settingmeta.FieldDescriptor
}

func (p StaticProvider) OwnerChange() string { return p.Owner }

func (p StaticProvider) Descriptors(context.Context) ([]settingmeta.FieldDescriptor, error) {
	out := make([]settingmeta.FieldDescriptor, len(p.Fields))
	for i := range p.Fields {
		out[i] = cloneDescriptor(p.Fields[i])
	}
	return out, nil
}
