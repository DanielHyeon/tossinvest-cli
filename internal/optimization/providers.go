package optimization

import (
	"context"

	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/settingmeta"
)

// CoreRegistry composes only descriptors that are actually present at the
// caller's HEAD. Later changes plug their exact provider into BuildRegistry;
// this function never invents a label, default, range, or option for them.
func CoreRegistry(ctx context.Context) (Registry, error) {
	descriptor := exitpolicy.CommonPolicyFieldDescriptor()
	required := []FieldRequirement{{
		Key: "exit.common-policy", Owner: "a041-complete-exit-line-contract", Category: CategoryExitProtection,
	}}
	return BuildRequiredRegistry(ctx, required, ProviderBinding{
		Category: CategoryExitProtection,
		Provider: coreProvider{owner: descriptor.Provenance.OwnerChange, fields: []settingmeta.FieldDescriptor{descriptor}},
	})
}

type coreProvider struct {
	owner  string
	fields []settingmeta.FieldDescriptor
}

func (p coreProvider) OwnerChange() string { return p.owner }

func (p coreProvider) Descriptors(context.Context) ([]settingmeta.FieldDescriptor, error) {
	out := make([]settingmeta.FieldDescriptor, len(p.fields))
	for i := range p.fields {
		out[i] = cloneDescriptor(p.fields[i])
	}
	return out, nil
}
