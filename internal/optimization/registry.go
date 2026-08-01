package optimization

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/settingmeta"
)

type RegisteredField struct {
	Category   Category
	Descriptor settingmeta.FieldDescriptor
}

type Registry struct {
	fields map[string]RegisteredField
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
		for _, descriptor := range descriptors {
			if err := descriptor.Validate(); err != nil {
				return Registry{}, fmt.Errorf("%w: owner %s field %q: %v", ErrInvalidRegistry, owner, descriptor.Key, err)
			}
			if descriptor.Provenance.OwnerChange != owner {
				return Registry{}, fmt.Errorf("%w: field %q claims owner %q, provider is %q",
					ErrInvalidRegistry, descriptor.Key, descriptor.Provenance.OwnerChange, owner)
			}
			if _, duplicate := fields[descriptor.Key]; duplicate {
				return Registry{}, fmt.Errorf("%w: field %q has more than one owner", ErrInvalidRegistry, descriptor.Key)
			}
			fields[descriptor.Key] = RegisteredField{Category: binding.Category, Descriptor: cloneDescriptor(descriptor)}
		}
	}
	return Registry{fields: fields}, nil
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
