package config_test

// operating_io_test.go is the disjointness of the three write surfaces
// (change console-owns-the-operating-toggles).
//
// limits_io.go could state its guarantee as a type: Save took GuardianLimits,
// which has no `enabled` field, so the switch was unwritable by construction.
// That stops being available the moment the console is allowed to write the
// switch at all — so the guarantee becomes three closed member lists, and these
// tests are what makes "closed" mean something. Each save is pointed at a file
// carrying values it must not touch, and the assertion is on the bytes.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
)

// full is a config file with every key this change can reach set to a value the
// saves under test must be caught moving.
const full = `{
  "schema_version": 4,
  "trading": {
    "place": false,
    "sell": false,
    "fractional": true,
    "cancel": false,
    "amend": true,
    "conditional": true,
    "allow_live_order_actions": false,
    "dangerous_automation": { "accept_fx_consent": true }
  },
  "engine": {
    "automation_gate": {
      "enabled": false,
      "limit_currency": "USD",
      "max_order_quantity": 100,
      "max_order_notional": 300,
      "max_total_exposure": 1000,
      "max_daily_loss_amount": 50,
      "max_daily_loss_ratio": 0.01
    },
    "adoption": { "enabled": true, "exclude_symbols": ["TSLA"] }
  },
  "openapi": { "enabled": true }
}
`

func writeConfig(t *testing.T, body string) (*config.Service, string) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("seeding: %v", err)
	}
	return config.NewService(path), path
}

func readDoc(t *testing.T, path string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading back: %v", err)
	}
	var doc map[string]any
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("the save produced invalid JSON: %v\n%s", err, data)
	}
	return doc
}

func gateOf(t *testing.T, doc map[string]any) map[string]any {
	t.Helper()
	engine, _ := doc["engine"].(map[string]any)
	gate, ok := engine["automation_gate"].(map[string]any)
	if !ok {
		t.Fatalf("no automation_gate in %v", doc)
	}
	return gate
}

// TestTheSwitchSaveMovesOnlyTheSwitch. The ceilings beside it are the ones an
// operator spent a screen choosing; a save that re-emits them from a read taken
// outside the lock is the lost update limits_io.go exists to refuse.
func TestTheSwitchSaveMovesOnlyTheSwitch(t *testing.T) {
	svc, path := writeConfig(t, full)

	if err := svc.SaveEngineGateEnabled(true); err != nil {
		t.Fatalf("SaveEngineGateEnabled: %v", err)
	}

	gate := gateOf(t, readDoc(t, path))
	if gate["enabled"] != true {
		t.Errorf("enabled = %v, want true", gate["enabled"])
	}
	for key, want := range map[string]any{
		"limit_currency":        "USD",
		"max_order_quantity":    float64(100),
		"max_order_notional":    float64(300),
		"max_total_exposure":    float64(1000),
		"max_daily_loss_amount": float64(50),
		"max_daily_loss_ratio":  0.01,
	} {
		if gate[key] != want {
			t.Errorf("the switch save moved %s: %v, want %v", key, gate[key], want)
		}
	}
}

// TestTheLimitSaveStillCannotMoveTheSwitch is the guarantee limits_io.go had as
// a type, restated now that a switch save exists in the same package.
func TestTheLimitSaveStillCannotMoveTheSwitch(t *testing.T) {
	svc, path := writeConfig(t, strings.Replace(full, `"enabled": false,`, `"enabled": true,`, 1))

	if err := svc.SaveEngineGateLimits(config.GuardianLimits{
		MaxOrderQuantity: 100, MaxOrderNotional: 500, MaxTotalExposure: 1_500,
		MaxDailyLossAmount: 50, MaxDailyLossRatio: 0.01, Currency: "USD",
	}); err != nil {
		t.Fatalf("SaveEngineGateLimits: %v", err)
	}

	gate := gateOf(t, readDoc(t, path))
	if gate["enabled"] != true {
		t.Fatalf("a limit save turned the gate off: enabled = %v", gate["enabled"])
	}
	if gate["max_order_notional"] != float64(500) {
		t.Errorf("the limit save did not land: %v", gate["max_order_notional"])
	}
}

// TestTheTradingSaveMovesOnlyItsFour. amend, conditional and fractional are not
// on the screen, so a save must leave whatever the file spells for them —
// including values the operator set by hand for a reason this build cannot see.
func TestTheTradingSaveMovesOnlyItsFour(t *testing.T) {
	svc, path := writeConfig(t, full)

	if err := svc.SaveTradingPolicy(config.TradingPolicy{
		Place: true, Sell: true, Cancel: true, AllowLiveOrderActions: true,
	}); err != nil {
		t.Fatalf("SaveTradingPolicy: %v", err)
	}

	doc := readDoc(t, path)
	trading, _ := doc["trading"].(map[string]any)
	for _, key := range []string{"place", "sell", "cancel", "allow_live_order_actions"} {
		if trading[key] != true {
			t.Errorf("%s = %v, want true", key, trading[key])
		}
	}
	for key, want := range map[string]any{
		"fractional":  true,
		"amend":       true,
		"conditional": true,
	} {
		if trading[key] != want {
			t.Errorf("the policy save moved %s: %v, want %v — it is not on the screen", key, trading[key], want)
		}
	}
	if da, ok := trading["dangerous_automation"].(map[string]any); !ok || da["accept_fx_consent"] != true {
		t.Errorf("the policy save disturbed dangerous_automation: %v", trading["dangerous_automation"])
	}
	// And it did not reach out of its own block.
	if gate := gateOf(t, doc); gate["enabled"] != false {
		t.Errorf("a trading save moved the gate: %v", gate["enabled"])
	}
}

// TestTheSwitchSaveDoesNotJudgeTheInterlock. Saving ON with an unsatisfiable
// interlock is a configuration the engine refuses loudly, not a corrupt file.
// Refusing it here would mean reproducing the interlock's judgement in a second
// place, which is how two implementations of one rule drift apart.
func TestTheSwitchSaveDoesNotJudgeTheInterlock(t *testing.T) {
	svc, path := writeConfig(t, `{"schema_version": 4, "engine": {"automation_gate": {}}}`)

	if err := svc.SaveEngineGateEnabled(true); err != nil {
		t.Fatalf("a switch save must not validate the rest of the gate: %v", err)
	}
	if gateOf(t, readDoc(t, path))["enabled"] != true {
		t.Error("the switch did not land")
	}
}

// TestAMissingBlockIsCreatedRatherThanRefused: an operator who has never saved
// anything still gets a working screen.
func TestAMissingBlockIsCreatedRatherThanRefused(t *testing.T) {
	svc, path := writeConfig(t, `{"schema_version": 4}`)

	if err := svc.SaveTradingPolicy(config.TradingPolicy{Place: true, Sell: true}); err != nil {
		t.Fatalf("SaveTradingPolicy into a bare document: %v", err)
	}
	if err := svc.SaveEngineGateEnabled(true); err != nil {
		t.Fatalf("SaveEngineGateEnabled into a bare document: %v", err)
	}

	doc := readDoc(t, path)
	trading, _ := doc["trading"].(map[string]any)
	if trading["place"] != true || trading["sell"] != true {
		t.Errorf("trading = %v", trading)
	}
	if trading["cancel"] != false {
		t.Errorf("cancel = %v, want the false that was saved", trading["cancel"])
	}
	if gateOf(t, doc)["enabled"] != true {
		t.Errorf("gate = %v", gateOf(t, doc))
	}
}

// TestAHandEditInProgressIsNotOverwritten. The file may be open in an editor
// with a comma missing; a save that reformats it would destroy the edit.
func TestAHandEditInProgressIsNotOverwritten(t *testing.T) {
	svc, path := writeConfig(t, `{"schema_version": 4, "trading": {`)

	err := svc.SaveTradingPolicy(config.TradingPolicy{Place: true})
	if err == nil {
		t.Fatal("a save into invalid JSON must be refused")
	}
	if !strings.Contains(err.Error(), "valid JSON") {
		t.Errorf("err = %v, want it to name the cause", err)
	}
	data, _ := os.ReadFile(path)
	if string(data) != `{"schema_version": 4, "trading": {` {
		t.Errorf("the refused save still wrote: %s", data)
	}
}

// TestMissingNamesTheTogglesTheInterlockWillName. The screen's list and the
// engine's refusal have to be the same words, or the operator fixes one thing
// and is told about another.
func TestMissingNamesTheTogglesTheInterlockWillName(t *testing.T) {
	got := config.TradingPolicy{Sell: true}.Missing()
	want := []string{"trading.place", "trading.cancel", "trading.allow_live_order_actions"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("Missing() = %v, want %v", got, want)
	}
	if !(config.TradingPolicy{Place: true, Sell: true, Cancel: true, AllowLiveOrderActions: true}).Complete() {
		t.Error("all four on must be Complete")
	}
	if (config.TradingPolicy{Place: true, Sell: true, Cancel: true}).Complete() {
		t.Error("a policy missing the live switch is not Complete")
	}
}
