package config_test

// soak_io_test.go is soak.autostart's half of the disjointness operating_io_test
// establishes for the engine's keys (openspec change a101).
//
// The tests are pointed at the same question those are: a save writes its one
// key and moves nothing else, and a load that finds nothing says false rather
// than guessing. What is new here is the pair of independence checks — the soak
// key and the engine key must not be able to move each other, because the whole
// reason there are two is that approving one is not approving the other.

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestSoakAutostartDefaultsOffAndLoadsExplicitValues(t *testing.T) {
	for _, tc := range []struct {
		name string
		body string
		want bool
	}{
		{"missing soak block", `{"schema_version": 4, "engine": {"autostart": true}}`, false},
		{"missing key", `{"soak": {}}`, false},
		{"explicit false", `{"soak": {"autostart": false}}`, false},
		{"explicit true", `{"soak": {"autostart": true}}`, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			svc, _ := writeConfig(t, tc.body)
			got, err := svc.LoadSoakAutostart()
			if err != nil {
				t.Fatalf("LoadSoakAutostart: %v", err)
			}
			if got != tc.want {
				t.Errorf("LoadSoakAutostart = %v, want %v", got, tc.want)
			}
		})
	}
}

// TestSoakAutostartLoadIsAbsentForAMissingFile. A machine that has never been
// configured has not approved anything, and that is not an error.
func TestSoakAutostartLoadIsAbsentForAMissingFile(t *testing.T) {
	svc, path := writeConfig(t, `{}`)
	if err := os.Remove(path); err != nil {
		t.Fatalf("removing the config: %v", err)
	}
	got, err := svc.LoadSoakAutostart()
	if err != nil {
		t.Fatalf("LoadSoakAutostart on a missing file: %v", err)
	}
	if got {
		t.Error("a missing config reported an approval")
	}
}

// TestSoakAutostartLoadRefusesUnreadableJSON. "There is no approval here" and
// "this file cannot be read" are different facts. Returning false for the second
// would start no survey and say nothing about why.
func TestSoakAutostartLoadRefusesUnreadableJSON(t *testing.T) {
	svc, _ := writeConfig(t, `{"soak": {"autostart": tru`)
	if _, err := svc.LoadSoakAutostart(); err == nil {
		t.Fatal("unparseable config was read as an absent approval")
	}
}

func TestSoakAutostartSaveMovesOnlyItsOwnKey(t *testing.T) {
	svc, path := writeConfig(t, full)

	if err := svc.SaveSoakAutostart(true); err != nil {
		t.Fatalf("SaveSoakAutostart: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading back: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("the saved config is not JSON: %v\n%s", err, data)
	}

	soak, ok := got["soak"].(map[string]any)
	if !ok {
		t.Fatalf("no soak block after the save:\n%s", data)
	}
	if soak["autostart"] != true {
		t.Errorf("soak.autostart = %v, want true", soak["autostart"])
	}

	var before map[string]any
	if err := json.Unmarshal([]byte(full), &before); err != nil {
		t.Fatalf("the fixture is not JSON: %v", err)
	}
	for key, want := range before {
		if key == "soak" {
			continue
		}
		gotJSON, _ := json.Marshal(got[key])
		wantJSON, _ := json.Marshal(want)
		if string(gotJSON) != string(wantJSON) {
			t.Errorf("%s changed:\n got %s\nwant %s", key, gotJSON, wantJSON)
		}
	}
}

// TestSoakAutostartSaveCreatesAMissingBlock. The key is new, so every existing
// configuration on disk is missing its block.
func TestSoakAutostartSaveCreatesAMissingBlock(t *testing.T) {
	svc, path := writeConfig(t, `{"schema_version": 4}`)

	if err := svc.SaveSoakAutostart(true); err != nil {
		t.Fatalf("SaveSoakAutostart: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading back: %v", err)
	}
	got, err := svc.LoadSoakAutostart()
	if err != nil || !got {
		t.Errorf("after creating the block, load = (%v, %v); want (true, nil)\n%s", got, err, data)
	}
	if !strings.Contains(string(data), `"schema_version"`) {
		t.Errorf("the save dropped an existing key:\n%s", data)
	}
}

// TestSoakAndEngineAutostartAreIndependent is the reason there are two keys.
// Approving the survey is not approving the engine, in either direction.
func TestSoakAndEngineAutostartAreIndependent(t *testing.T) {
	svc, _ := writeConfig(t, `{"engine": {"autostart": true}, "soak": {"autostart": false}}`)

	if err := svc.SaveSoakAutostart(true); err != nil {
		t.Fatalf("SaveSoakAutostart: %v", err)
	}
	engine, err := svc.LoadEngineAutostart()
	if err != nil {
		t.Fatalf("LoadEngineAutostart: %v", err)
	}
	if !engine {
		t.Error("saving the soak approval cleared the engine's")
	}

	if err := svc.SaveEngineAutostart(false); err != nil {
		t.Fatalf("SaveEngineAutostart: %v", err)
	}
	soak, err := svc.LoadSoakAutostart()
	if err != nil {
		t.Fatalf("LoadSoakAutostart: %v", err)
	}
	if !soak {
		t.Error("clearing the engine's approval cleared the survey's")
	}
}

// TestSoakAutostartSaveRefusesInvalidJSONWithoutOverwritingIt. The splice reads
// the file it is editing; a file it cannot parse must come back untouched rather
// than be replaced with a valid file that lost everything.
func TestSoakAutostartSaveRefusesInvalidJSONWithoutOverwritingIt(t *testing.T) {
	const broken = `{"soak": {"autostart": tru`
	svc, path := writeConfig(t, broken)

	if err := svc.SaveSoakAutostart(true); err == nil {
		t.Fatal("the save accepted an unparseable config")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading back: %v", err)
	}
	if string(data) != broken {
		t.Errorf("the refused save still rewrote the file:\n got %s\nwant %s", data, broken)
	}
}
