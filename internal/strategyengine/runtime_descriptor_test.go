package strategyengine

import "testing"

func TestDormantRuntimeDescriptorSeparatesSectionsAndNeverPretendsActivation(t *testing.T) {
	d := DormantRuntimeDescriptor()
	if d.Category != "strategy-runtime" {
		t.Fatal(d.Category)
	}
	want := [4]string{"parameters", "lane", "autostart", "live"}
	for i, s := range d.Sections {
		if s.ID != want[i] || s.ActionOwner != "a050" {
			t.Fatalf("section=%+v", s)
		}
	}
	for _, f := range d.Fields {
		if f.Label == "" || f.Help == "" || f.Default == "" || f.Desired == "" || f.Effective != "미구성" || f.Unit == "" || f.Range == "" || f.Provenance == "" || f.ApplyTiming == "" {
			t.Fatalf("field incomplete=%+v", f)
		}
	}
	if d.Blockers[0].Effective != "UNWIRED" || d.Blockers[3].Effective != "NOT_CONFIGURED" {
		t.Fatalf("blockers=%+v", d.Blockers)
	}
}
