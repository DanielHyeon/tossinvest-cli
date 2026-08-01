package positionpolicy

import "testing"

func TestDescriptorPinsSafeDefaultsAndStopRegistry(t *testing.T) {
	d := Descriptor()
	if d.Category != "position-management" || d.AutoEnabledDefault || d.AutoEnabledDesired || d.AutoEnabledEffective {
		t.Fatalf("descriptor defaults = %+v", d)
	}
	if d.StopDefault != "5%" || len(d.StopOptions) != 37 {
		t.Fatalf("stop registry default=%q options=%d", d.StopDefault, len(d.StopOptions))
	}
	if d.StopOptions[0].Label != "2%" || d.StopOptions[len(d.StopOptions)-1].Label != "20%" {
		t.Fatalf("stop range = %q..%q", d.StopOptions[0].Label, d.StopOptions[len(d.StopOptions)-1].Label)
	}
	for index := 1; index < len(d.StopOptions); index++ {
		if d.StopOptions[index].ID == d.StopOptions[index-1].ID {
			t.Fatalf("duplicate stop option at %d", index)
		}
	}
	if d.IncludeDefault == nil || d.ExcludeDefault == nil || len(d.IncludeDefault) != 0 || len(d.ExcludeDefault) != 0 {
		t.Fatalf("include/exclude defaults are not explicit empty lists: %+v", d)
	}
	if d.ExcludePrecedence != "exclude 우선" || d.OneShareBehavior == "" {
		t.Fatalf("behavioral help missing: %+v", d)
	}
}

func TestUnsetDesiredAndEffectiveLabelsBothExplainInheritance(t *testing.T) {
	state := State{}
	if state.DesiredLabel() != "공통 정책 상속" || state.EffectiveLabel() != "공통 정책 상속" {
		t.Fatalf("labels = desired %q effective %q", state.DesiredLabel(), state.EffectiveLabel())
	}
}
