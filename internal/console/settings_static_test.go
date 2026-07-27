package console

// settings_static_test.go pins the narrowness of the console's one config write
// surface (console-adoption-controls, review round 1 P2-8).

import (
	"reflect"
	"strings"
	"testing"
)

// TestTheSettingsSeamDeclaresExactlyLoadAndSave: every method added to the seam
// is a capability the console gains over the config file; there are two.
func TestTheSettingsSeamDeclaresExactlyLoadAndSave(t *testing.T) {
	iface := reflect.TypeOf((*AdoptionSettings)(nil)).Elem()
	if iface.NumMethod() != 2 {
		t.Fatalf("AdoptionSettings declares %d methods, want exactly Load and Save", iface.NumMethod())
	}
	for _, name := range []string{"Load", "Save"} {
		if _, ok := iface.MethodByName(name); !ok {
			t.Errorf("AdoptionSettings lost %s; the seam's shape is part of the spec", name)
		}
	}
}

// TestTheConsoleNeverNamesTheConfigService: the seam is injected. A console that
// constructed config.Service itself would hold the whole file — Init writes
// defaults, and a future exported writer would arrive silently.
func TestTheConsoleNeverNamesTheConfigService(t *testing.T) {
	for name, src := range packageFiles(t) {
		code := strings.Join(nonCommentLines(src), "\n")
		for _, banned := range []string{"config.NewService", "config.Service"} {
			if strings.Contains(code, banned) {
				t.Errorf("%s names %q; the console receives an AdoptionSettings seam and must not "+
					"hold the config service", name, banned)
			}
		}
	}
}
