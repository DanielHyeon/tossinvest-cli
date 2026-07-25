package config

// engine_test.go covers the automation-gate config block (task 4.2).
//
// The tests that matter here are the boring ones: that the gate is off unless
// somebody wrote it on, and that every config file written by an earlier schema
// version still loads to exactly the behaviour it had before (§0.2, §0.6).

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func writeConfig(t *testing.T, body string) *Service {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return NewService(path)
}

// TestAutomationGateDefaultsOff is the §0.2 test: a config with no engine block
// — which is every config anyone has today — must produce a closed gate with no
// limits.
func TestAutomationGateDefaultsOff(t *testing.T) {
	svc := writeConfig(t, `{"schema_version":4,"trading":{}}`)

	cfg, err := svc.Load(t.Context())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	gate := cfg.Engine.AutomationGate
	if gate.Enabled {
		t.Error("the automation gate must default to off")
	}
	if gate.LimitsSet() {
		t.Errorf("limits must default to unset, got %+v", gate)
	}
	if gate.AttestationFile != "" {
		t.Errorf("attestation_file must default to empty, got %q", gate.AttestationFile)
	}
}

// TestDefaultFileHasGateOff pins the same thing for a freshly initialised config:
// `tossctl config init` must not write an open gate.
func TestDefaultFileHasGateOff(t *testing.T) {
	if DefaultFile().Engine.AutomationGate.Enabled {
		t.Fatal("DefaultFile must not enable the automation gate")
	}
	if DefaultFile().Engine.AutomationGate.LimitsSet() {
		t.Fatal("DefaultFile must not carry limits")
	}
}

// TestAutomationGateParsed covers a fully written block.
func TestAutomationGateParsed(t *testing.T) {
	svc := writeConfig(t, `{
  "schema_version": 5,
  "trading": {},
  "engine": {
    "automation_gate": {
      "enabled": true,
      "attestation_file": "/tmp/att.json",
      "max_order_quantity": 10,
      "max_order_notional": 1000000,
      "limit_currency": "KRW"
    }
  }
}`)

	cfg, err := svc.Load(t.Context())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	gate := cfg.Engine.AutomationGate
	if !gate.Enabled {
		t.Error("enabled = false, want true")
	}
	if gate.AttestationFile != "/tmp/att.json" {
		t.Errorf("attestation_file = %q", gate.AttestationFile)
	}
	if gate.MaxOrderQuantity != 10 {
		t.Errorf("max_order_quantity = %v", gate.MaxOrderQuantity)
	}
	if gate.MaxOrderNotional != 1000000 {
		t.Errorf("max_order_notional = %v", gate.MaxOrderNotional)
	}
	if gate.LimitCurrency != "KRW" {
		t.Errorf("limit_currency = %q", gate.LimitCurrency)
	}
	if !gate.LimitsSet() {
		t.Error("LimitsSet() = false with two positive ceilings")
	}
}

// TestAutomationGateEnabledWithoutLimits is the shape the interlock has to
// refuse. The config layer parses it faithfully — deciding it is unsafe is the
// engine's job, not the parser's — but LimitsSet must report the truth about it.
func TestAutomationGateEnabledWithoutLimits(t *testing.T) {
	svc := writeConfig(t, `{"schema_version":5,"trading":{},"engine":{"automation_gate":{"enabled":true}}}`)

	cfg, err := svc.Load(t.Context())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !cfg.Engine.AutomationGate.Enabled {
		t.Error("enabled = false, want true")
	}
	if cfg.Engine.AutomationGate.LimitsSet() {
		t.Error("LimitsSet() = true with no ceilings — an unlimited gate is not an authorised one")
	}
}

// TestEveryEarlierSchemaVersionStillLoads is the migration rule (§0.6) as a test.
//
// Each of these is a config file a user could have on disk right now. Every one
// of them must keep loading, keep its existing toggles, and produce a closed
// gate — an upgrade that silently opened the gate would be the worst possible
// outcome of adding it.
func TestEveryEarlierSchemaVersionStillLoads(t *testing.T) {
	cases := []struct {
		name string
		body string
		// wantPlace is the trading toggle that must survive the migration.
		wantPlace  bool
		wantLegacy string
	}{
		{
			name:      "v1 with the legacy execute toggle",
			body:      `{"schema_version":1,"trading":{"place":true,"allow_dangerous_execute":true}}`,
			wantPlace: true,
		},
		{
			name:       "v2 with the removed kr scope",
			body:       `{"schema_version":2,"trading":{"place":true,"kr":true}}`,
			wantPlace:  true,
			wantLegacy: "trading.kr",
		},
		{
			name:      "v3 with an openapi block",
			body:      `{"schema_version":3,"trading":{"place":true},"openapi":{"enabled":false,"prefer":"wts"}}`,
			wantPlace: true,
		},
		{
			name:      "v4, the version immediately before the engine block",
			body:      `{"schema_version":4,"trading":{"place":true,"cancel":true,"amend":true,"allow_live_order_actions":true}}`,
			wantPlace: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := writeConfig(t, tc.body)

			status, err := svc.Status(t.Context())
			if err != nil {
				t.Fatalf("Status: %v", err)
			}
			if status.Trading.Place != tc.wantPlace {
				t.Errorf("trading.place = %v, want %v", status.Trading.Place, tc.wantPlace)
			}
			if status.Engine.AutomationGate.Enabled {
				t.Error("a config upgrade must never turn the automation gate on")
			}
			if status.SchemaVersion != SchemaVersion {
				t.Errorf("effective schema version = %d, want %d", status.SchemaVersion, SchemaVersion)
			}
			if tc.wantLegacy != "" {
				found := false
				for _, f := range status.LegacyFields {
					if f == tc.wantLegacy {
						found = true
					}
				}
				if !found {
					t.Errorf("legacy field %q not reported; got %v", tc.wantLegacy, status.LegacyFields)
				}
			}
		})
	}
}

// TestPublishedSchemaMatchesSchemaVersion keeps schemas/config.schema.json and
// the constant from drifting. They are two halves of one contract: the file is
// what an editor validates a user's config against, and a stale const there means
// every v5 config is reported as invalid by the user's tooling.
func TestPublishedSchemaMatchesSchemaVersion(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "schemas", "config.schema.json"))
	if err != nil {
		t.Fatalf("reading the published schema: %v", err)
	}
	var schema struct {
		Properties struct {
			SchemaVersion struct {
				Const int `json:"const"`
			} `json:"schema_version"`
		} `json:"properties"`
	}
	if err := json.Unmarshal(data, &schema); err != nil {
		t.Fatalf("parsing the published schema: %v", err)
	}
	if got := schema.Properties.SchemaVersion.Const; got != SchemaVersion {
		t.Errorf("schemas/config.schema.json pins schema_version %d, the code writes %d", got, SchemaVersion)
	}
}

// TestPublishedSchemaAcceptsTheEngineBlock guards the other half: the schema
// declares additionalProperties:false, so a block the code writes but the schema
// does not declare makes every generated config invalid.
func TestPublishedSchemaAcceptsTheEngineBlock(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "schemas", "config.schema.json"))
	if err != nil {
		t.Fatalf("reading the published schema: %v", err)
	}
	var schema struct {
		Properties map[string]struct {
			Properties map[string]json.RawMessage `json:"properties"`
		} `json:"properties"`
	}
	if err := json.Unmarshal(data, &schema); err != nil {
		t.Fatalf("parsing the published schema: %v", err)
	}
	engine, ok := schema.Properties["engine"]
	if !ok {
		t.Fatal("the published schema has no `engine` block, but the code writes one")
	}
	if _, ok := engine.Properties["automation_gate"]; !ok {
		t.Error("the published schema's engine block declares no automation_gate")
	}
}
