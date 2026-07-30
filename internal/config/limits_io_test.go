package config

// limits_io_test.go covers the Guardian-limit writer (change
// console-sets-guardian-limits, task 4.x).
//
// The load-bearing test in this file is TestSavingLimitsNeverRewritesEnabled.
// The console may write the numbers and may not write the switch, and design D6
// makes that a property of the writer rather than of the caller's discipline:
// the six limit keys are spliced individually and `enabled` is never emitted.

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func limitsFor(t *testing.T, id string) GuardianLimits {
	t.Helper()
	tier, ok := GuardianTierByID(id)
	if !ok {
		t.Fatalf("tier %q is not registered", id)
	}
	return tier.Limits
}

func readBack(t *testing.T, svc *Service) map[string]any {
	t.Helper()
	data, err := os.ReadFile(svc.Path())
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	var doc map[string]any
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("read back is not JSON: %v\n%s", err, data)
	}
	return doc
}

func gateOf(t *testing.T, doc map[string]any) map[string]any {
	t.Helper()
	engine, ok := doc["engine"].(map[string]any)
	if !ok {
		t.Fatalf("no engine block in %v", doc)
	}
	gate, ok := engine["automation_gate"].(map[string]any)
	if !ok {
		t.Fatalf("no automation_gate block in %v", engine)
	}
	return gate
}

// TestSavingLimitsNeverRewritesEnabled — both directions, because the failure
// this prevents is symmetric: a stale read could as easily turn a live gate off
// as turn a dead one on, and only one of those is merely an outage.
func TestSavingLimitsNeverRewritesEnabled(t *testing.T) {
	for _, enabled := range []string{"true", "false"} {
		svc := writeConfig(t, `{
  "schema_version": 5,
  "trading": {},
  "engine": {"automation_gate": {"enabled": `+enabled+`, "attestation_file": "/tmp/att.json"}}
}`)
		if err := svc.SaveEngineGateLimits(limitsFor(t, "kr-small-live")); err != nil {
			t.Fatalf("SaveEngineGateLimits: %v", err)
		}
		gate := gateOf(t, readBack(t, svc))
		if got := gate["enabled"]; got != (enabled == "true") {
			t.Errorf("enabled = %v after a limit save, want %s unchanged", got, enabled)
		}
		if got := gate["attestation_file"]; got != "/tmp/att.json" {
			t.Errorf("attestation_file = %v, want it preserved", got)
		}
		if got := gate["max_order_notional"]; got != float64(1_000_000) {
			t.Errorf("max_order_notional = %v, want 1000000", got)
		}
		if got := gate["limit_currency"]; got != "KRW" {
			t.Errorf("limit_currency = %v, want KRW", got)
		}
	}
}

// TestTheSaveCarriesNoEnabledKeyOfItsOwn is the stronger form of the same
// claim, read off the bytes: a file whose gate never spelled `enabled` must not
// acquire one from a limit save. Emitting `"enabled": false` here would be the
// console writing the switch, even to its safe position.
func TestTheSaveCarriesNoEnabledKeyOfItsOwn(t *testing.T) {
	svc := writeConfig(t, `{"schema_version":5,"trading":{},"engine":{"automation_gate":{}}}`)
	if err := svc.SaveEngineGateLimits(limitsFor(t, "us-smoke")); err != nil {
		t.Fatalf("SaveEngineGateLimits: %v", err)
	}
	if _, ok := gateOf(t, readBack(t, svc))["enabled"]; ok {
		t.Error("the limit save introduced an enabled key; the console does not write the switch")
	}
}

// TestUnknownKeysSurviveToTheByte — config.json is hand-edited, and a key this
// build has never heard of is somebody's setting.
func TestUnknownKeysSurviveToTheByte(t *testing.T) {
	svc := writeConfig(t, `{
  "schema_version": 5,
  "trading": {},
  "$unknown": "keep me",
  "engine": {"automation_gate": {"enabled": false, "$future": {"nested": [1, 2]}}}
}`)
	if err := svc.SaveEngineGateLimits(limitsFor(t, "kr-smoke")); err != nil {
		t.Fatalf("SaveEngineGateLimits: %v", err)
	}
	data, err := os.ReadFile(svc.Path())
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	for _, want := range []string{`"$unknown"`, "keep me", `"$future"`, `"nested"`} {
		if !strings.Contains(string(data), want) {
			t.Errorf("the save dropped %s:\n%s", want, data)
		}
	}
}

// TestTheAdoptionBlockIsUntouched: the two writers share a file and must not
// share a blast radius.
func TestTheAdoptionBlockIsUntouched(t *testing.T) {
	svc := writeConfig(t, `{
  "schema_version": 5,
  "trading": {},
  "engine": {
    "automation_gate": {"enabled": false},
    "adoption": {"enabled": true, "default_stop_pct": 0.05, "exclude_symbols": ["TSLA"]}
  }
}`)
	if err := svc.SaveEngineGateLimits(limitsFor(t, "kr-small-live")); err != nil {
		t.Fatalf("SaveEngineGateLimits: %v", err)
	}
	block, verdict, err := svc.LoadRawEngineAdoption()
	if err != nil {
		t.Fatalf("LoadRawEngineAdoption: %v", err)
	}
	if verdict != "" {
		t.Errorf("the adoption block became invalid: %s", verdict)
	}
	if !block.Enabled || block.DefaultStopPct != 0.05 || len(block.ExcludeSymbols) != 1 ||
		block.ExcludeSymbols[0] != "TSLA" {
		t.Errorf("the adoption block changed: %+v", block)
	}
}

// TestSavingRefusesWhatTheInterlockWouldRefuse.
func TestSavingRefusesWhatTheInterlockWouldRefuse(t *testing.T) {
	ok := limitsFor(t, "kr-small-live")
	for _, tc := range []struct {
		name string
		bend func(*GuardianLimits)
	}{
		{"zero quantity", func(l *GuardianLimits) { l.MaxOrderQuantity = 0 }},
		{"negative notional", func(l *GuardianLimits) { l.MaxOrderNotional = -1 }},
		{"zero exposure", func(l *GuardianLimits) { l.MaxTotalExposure = 0 }},
		{"NaN daily loss", func(l *GuardianLimits) { l.MaxDailyLossAmount = math.NaN() }},
		{"infinite ratio", func(l *GuardianLimits) { l.MaxDailyLossRatio = math.Inf(1) }},
		{"ratio above one", func(l *GuardianLimits) { l.MaxDailyLossRatio = 1.5 }},
		{"no currency", func(l *GuardianLimits) { l.Currency = "  " }},
	} {
		body := `{"schema_version":5,"trading":{},"engine":{"automation_gate":{"enabled":false}}}`
		svc := writeConfig(t, body)
		bad := ok
		tc.bend(&bad)
		if err := svc.SaveEngineGateLimits(bad); err == nil {
			t.Errorf("%s: the save was accepted; the engine would refuse to start on it", tc.name)
		}
		data, _ := os.ReadFile(svc.Path())
		if string(data) != body {
			t.Errorf("%s: a refused save still rewrote the file:\n%s", tc.name, data)
		}
	}
}

// TestSavingRefusesAboveTheCeiling — the fat-finger backstop at the write path,
// so a caller that skipped the check cannot get past it (design D5).
func TestSavingRefusesAboveTheCeiling(t *testing.T) {
	svc := writeConfig(t, `{"schema_version":5,"trading":{},"engine":{"automation_gate":{"enabled":false}}}`)
	over := limitsFor(t, "kr-small-live")
	over.MaxTotalExposure = 100_000_000
	err := svc.SaveEngineGateLimits(over)
	if err == nil {
		t.Fatal("a value above the registered ceiling was saved")
	}
	if !strings.Contains(err.Error(), "max_total_exposure") {
		t.Errorf("the refusal %q does not name the field", err)
	}
}

// TestSavingRefusesAnUnregisteredCurrency — fail-closed, matching StockOS's
// _market_ceiling.
func TestSavingRefusesAnUnregisteredCurrency(t *testing.T) {
	svc := writeConfig(t, `{"schema_version":5,"trading":{},"engine":{"automation_gate":{"enabled":false}}}`)
	odd := limitsFor(t, "kr-small-live")
	odd.Currency = "JPY"
	if err := svc.SaveEngineGateLimits(odd); err == nil {
		t.Fatal("an unregistered currency was saved; it has no ceiling and so no backstop")
	}
}

// TestTheGateBlockIsCreatedWhenAbsent covers the three insertion shapes.
func TestTheGateBlockIsCreatedWhenAbsent(t *testing.T) {
	for _, tc := range []struct{ name, body string }{
		{"no automation_gate", `{"schema_version":5,"trading":{},"engine":{"adoption":{}}}`},
		{"no engine", `{"schema_version":5,"trading":{}}`},
	} {
		svc := writeConfig(t, tc.body)
		if err := svc.SaveEngineGateLimits(limitsFor(t, "us-small-live")); err != nil {
			t.Fatalf("%s: SaveEngineGateLimits: %v", tc.name, err)
		}
		gate := gateOf(t, readBack(t, svc))
		if got := gate["max_order_notional"]; got != float64(300) {
			t.Errorf("%s: max_order_notional = %v, want 300", tc.name, got)
		}
		if _, ok := gate["enabled"]; ok {
			t.Errorf("%s: a created gate block carries an enabled key", tc.name)
		}
	}
}

// TestSavingIntoNoFileCreatesASkeleton — with the gate still off, because
// DefaultFile's gate is off and this writer adds no enabled key.
func TestSavingIntoNoFileCreatesASkeleton(t *testing.T) {
	svc := NewService(filepath.Join(t.TempDir(), "config.json"))
	if err := svc.SaveEngineGateLimits(limitsFor(t, "kr-small-live")); err != nil {
		t.Fatalf("SaveEngineGateLimits: %v", err)
	}
	cfg, err := svc.Load(t.Context())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Engine.AutomationGate.Enabled {
		t.Error("a limit save created a config with the gate on")
	}
	if cfg.Engine.AutomationGate.MaxOrderNotional != 1_000_000 {
		t.Errorf("max_order_notional = %v", cfg.Engine.AutomationGate.MaxOrderNotional)
	}
}

// TestSavingRefusesAFileThatIsNotJSON — a hand-edit in progress is not a file to
// overwrite.
func TestSavingRefusesAFileThatIsNotJSON(t *testing.T) {
	body := `{"schema_version": 5, "trading": {`
	svc := writeConfig(t, body)
	if err := svc.SaveEngineGateLimits(limitsFor(t, "kr-smoke")); err == nil {
		t.Fatal("an unparseable config was overwritten")
	}
	data, _ := os.ReadFile(svc.Path())
	if string(data) != body {
		t.Errorf("the refused save still wrote:\n%s", data)
	}
}

// TestLoadRawEngineGateReadsTheSwitchWithoutOfferingIt: the screen shows whether
// the gate is on; the seam's write side takes limits only, so there is no shape
// in which that value travels back.
func TestLoadRawEngineGateReadsTheSwitchWithoutOfferingIt(t *testing.T) {
	svc := writeConfig(t, `{
  "schema_version": 5,
  "trading": {},
  "engine": {"automation_gate": {"enabled": true, "max_order_quantity": 7, "limit_currency": "krw"}}
}`)
	gate, err := svc.LoadRawEngineGate()
	if err != nil {
		t.Fatalf("LoadRawEngineGate: %v", err)
	}
	if !gate.Enabled {
		t.Error("enabled = false, want the file's true")
	}
	if gate.MaxOrderQuantity != 7 {
		t.Errorf("max_order_quantity = %v, want 7", gate.MaxOrderQuantity)
	}
	if gate.LimitCurrency != "krw" {
		t.Errorf("limit_currency = %q, want the file's own spelling", gate.LimitCurrency)
	}
}

// TestLoadRawEngineGateOnAMissingFile is the zero gate and no error — the screen
// renders 미설정 rather than an error nobody can act on.
func TestLoadRawEngineGateOnAMissingFile(t *testing.T) {
	svc := NewService(filepath.Join(t.TempDir(), "config.json"))
	gate, err := svc.LoadRawEngineGate()
	if err != nil {
		t.Fatalf("LoadRawEngineGate: %v", err)
	}
	if gate.Enabled || gate.LimitsSet() {
		t.Errorf("a missing file produced %+v", gate)
	}
}
