package config

// adoption_io_test.go covers the console's config seam (change
// console-adoption-controls task 1.2): reading the engine.adoption block as
// written — before merge zeroing can eat a refused block's lists — and writing
// it back surgically, so a hand-edited file loses nothing but the block that was
// asked to change.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// TestRawLoadReturnsARefusedBlockAsWritten is review round-1 P1-1: the operator
// fixes only the fraction on the settings screen; the exclude that was
// protecting a long-term holding must still be there to save back.
func TestRawLoadReturnsARefusedBlockAsWritten(t *testing.T) {
	svc := writeConfig(t, `{"schema_version":5,"trading":{},
	  "engine":{"adoption":{"enabled":true,"default_stop_pct":0.005,
	    "exclude_symbols":["005930"],"include_symbols":["000660"]}}}`)

	raw, verdict, err := svc.LoadRawEngineAdoption()
	if err != nil {
		t.Fatalf("LoadRawEngineAdoption: %v", err)
	}
	if verdict == "" {
		t.Error("a 0.005 fraction is out of band; the verdict must say so")
	}
	if !raw.Enabled || raw.DefaultStopPct != 0.005 {
		t.Errorf("raw block = %+v, want the file's own values, not the zeroed merge", raw)
	}
	if !raw.Excludes("005930") || !raw.Included("000660") {
		t.Error("the refused block's lists were lost on the read side; saving this back would " +
			"silently drop an exclusion that was protecting a holding")
	}
}

func TestRawLoadOfAMissingFileIsTheZeroBlock(t *testing.T) {
	svc := NewService(filepath.Join(t.TempDir(), "config.json"))
	raw, verdict, err := svc.LoadRawEngineAdoption()
	if err != nil {
		t.Fatalf("LoadRawEngineAdoption on a missing file: %v", err)
	}
	if raw.Enabled || verdict != "" {
		t.Errorf("missing file: raw=%+v verdict=%q, want the zero block and no verdict", raw, verdict)
	}
}

// TestSaveReplacesOnlyTheAdoptionBlock is the surgical-write SHALL: every byte
// outside the engine.adoption value — unknown keys, odd spacing, key order —
// survives identically.
func TestSaveReplacesOnlyTheAdoptionBlock(t *testing.T) {
	body := `{
  "schema_version": 5,
  "$future_key": {"the": ["console", "does", "not", "know", "this"]},
  "trading": {"note":   "odd   spacing kept"},
  "engine": {
    "automation_gate": {"enabled": false},
    "adoption": {"enabled": false},
    "another_future_key": 7
  }
}`
	svc := writeConfig(t, body)

	if err := svc.SaveEngineAdoption(Adoption{
		Enabled: true, DefaultStopPct: 0.05, IncludeSymbols: []string{"005930"},
	}); err != nil {
		t.Fatalf("SaveEngineAdoption: %v", err)
	}

	after, err := os.ReadFile(svc.Path())
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	// The block itself changed…
	raw, verdict, err := svc.LoadRawEngineAdoption()
	if err != nil || verdict != "" {
		t.Fatalf("re-load: %v verdict=%q", err, verdict)
	}
	if !raw.Enabled || raw.DefaultStopPct != 0.05 || !raw.Included("005930") {
		t.Errorf("saved block did not round-trip: %+v", raw)
	}
	// …and nothing else did: cut the adoption value out of both texts and the
	// remainders must be byte-identical.
	gutBefore := cutAdoptionValue(t, []byte(body))
	gutAfter := cutAdoptionValue(t, after)
	if gutBefore != gutAfter {
		t.Errorf("bytes outside engine.adoption changed:\n--- before ---\n%s\n--- after ---\n%s",
			gutBefore, gutAfter)
	}
	// The merge path still accepts the file.
	cfg, err := svc.Load(t.Context())
	if err != nil {
		t.Fatalf("Load after save: %v", err)
	}
	if !cfg.Engine.Adoption.Enabled || cfg.Engine.Adoption.Rejected != "" {
		t.Errorf("merged block after save = %+v", cfg.Engine.Adoption)
	}
}

// cutAdoptionValue removes the engine.adoption value span so the surrounding
// bytes can be compared exactly.
func cutAdoptionValue(t *testing.T, data []byte) string {
	t.Helper()
	start, end, found, err := adoptionValueSpan(data)
	if err != nil {
		t.Fatalf("adoptionValueSpan: %v", err)
	}
	if !found {
		return string(data)
	}
	return string(data[:start]) + string(data[end:])
}

func TestSaveIntoAFileWithNoEngineBlock(t *testing.T) {
	svc := writeConfig(t, `{"schema_version":5,"trading":{}}`)
	if err := svc.SaveEngineAdoption(Adoption{Enabled: true, DefaultStopPct: 0.05}); err != nil {
		t.Fatalf("SaveEngineAdoption: %v", err)
	}
	cfg, err := svc.Load(t.Context())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !cfg.Engine.Adoption.Enabled || cfg.Engine.Adoption.Rejected != "" {
		t.Errorf("adoption = %+v, want the saved block", cfg.Engine.Adoption)
	}
	if data, _ := os.ReadFile(svc.Path()); !strings.Contains(string(data), `"trading"`) {
		t.Error("the existing keys were lost when the engine block was created")
	}
}

func TestSaveIntoAnEngineBlockWithoutAdoption(t *testing.T) {
	svc := writeConfig(t, `{"schema_version":5,"trading":{},
	  "engine":{"automation_gate":{"enabled":false}}}`)
	if err := svc.SaveEngineAdoption(Adoption{Enabled: true, DefaultStopPct: 0.05}); err != nil {
		t.Fatalf("SaveEngineAdoption: %v", err)
	}
	cfg, err := svc.Load(t.Context())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !cfg.Engine.Adoption.Enabled {
		t.Errorf("adoption = %+v, want the saved block", cfg.Engine.Adoption)
	}
	if cfg.Engine.AutomationGate.Enabled {
		t.Error("the automation gate block was disturbed")
	}
}

// TestSaveRefusesAnInvalidBlock: writing a block the engine would zero is the
// worst outcome — the file says on, the engine runs off, and nobody is told.
func TestSaveRefusesAnInvalidBlock(t *testing.T) {
	svc := writeConfig(t, `{"schema_version":5,"trading":{}}`)
	before, _ := os.ReadFile(svc.Path())
	if err := svc.SaveEngineAdoption(Adoption{Enabled: true, DefaultStopPct: 0.001}); err == nil {
		t.Fatal("an out-of-band fraction was saved; the engine would zero it at load")
	}
	after, _ := os.ReadFile(svc.Path())
	if string(before) != string(after) {
		t.Error("a refused save still changed the file")
	}
}

// TestSaveRefusesACorruptFile: a file that cannot be parsed is somebody's
// hand-edit in progress or damage; overwriting it with a skeleton would destroy
// it. Skeleton creation is for a file that does not exist, only.
func TestSaveRefusesACorruptFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(`{"schema_version": 5,`), 0o600); err != nil {
		t.Fatal(err)
	}
	svc := NewService(path)
	if err := svc.SaveEngineAdoption(Adoption{Enabled: true, DefaultStopPct: 0.05}); err == nil {
		t.Fatal("a corrupt config file was overwritten")
	}
	if data, _ := os.ReadFile(path); string(data) != `{"schema_version": 5,` {
		t.Error("the corrupt file's bytes were not left alone")
	}
}

func TestSaveCreatesASkeletonOnlyWhenTheFileIsAbsent(t *testing.T) {
	svc := NewService(filepath.Join(t.TempDir(), "config.json"))
	if err := svc.SaveEngineAdoption(Adoption{DefaultStopPct: 0.05,
		IncludeSymbols: []string{"005930"}}); err != nil {
		t.Fatalf("SaveEngineAdoption on a missing file: %v", err)
	}
	cfg, err := svc.Load(t.Context())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !cfg.Engine.Adoption.Included("005930") || cfg.Engine.Adoption.Rejected != "" {
		t.Errorf("adoption = %+v, want the saved include", cfg.Engine.Adoption)
	}
	var doc map[string]json.RawMessage
	data, _ := os.ReadFile(svc.Path())
	if json.Unmarshal(data, &doc) != nil || doc["schema_version"] == nil {
		t.Error("the skeleton is not a loadable config file")
	}
}

// TestConcurrentSavesLoseNothing is review round-1 P2-4: two writers, one lock —
// the last block wins but no writer installs a half-written or stale file over
// the other's rename.
func TestConcurrentSavesLoseNothing(t *testing.T) {
	svc := writeConfig(t, `{"schema_version":5,"trading":{"note":"keep me"}}`)

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			block := Adoption{DefaultStopPct: 0.05, IncludeSymbols: []string{"005930"}}
			if n%2 == 0 {
				block.Enabled = true
			}
			if err := svc.SaveEngineAdoption(block); err != nil {
				t.Errorf("concurrent save %d: %v", n, err)
			}
		}(i)
	}
	wg.Wait()

	cfg, err := svc.Load(t.Context())
	if err != nil {
		t.Fatalf("Load after concurrent saves: %v", err)
	}
	if cfg.Engine.Adoption.Rejected != "" || !cfg.Engine.Adoption.Included("005930") {
		t.Errorf("adoption = %+v, want one of the written blocks intact", cfg.Engine.Adoption)
	}
	if data, _ := os.ReadFile(svc.Path()); !strings.Contains(string(data), "keep me") {
		t.Error("a concurrent save dropped the unrelated keys")
	}
}
