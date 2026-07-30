package config

import (
	"os"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
)

func TestMissingCommonExitPolicyPreservesLegacyRatchetSelection(t *testing.T) {
	svc := writeConfig(t, `{"schema_version":5,"trading":{},"engine":{"adoption":{"enabled":false}}}`)
	cfg, err := svc.Load(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Engine.ExitPolicy.CommonPolicy != "" || cfg.Engine.ExitPolicy.Rejected != "" {
		t.Fatalf("exit policy = %+v, want empty legacy selection", cfg.Engine.ExitPolicy)
	}
}

func TestUnknownCommonExitPolicyIsRetainedAsRejected(t *testing.T) {
	svc := writeConfig(t, `{"schema_version":5,"trading":{},
	  "engine":{"exit_policy":{"common_policy":"SOMETHING_ELSE"}}}`)
	cfg, err := svc.Load(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Engine.ExitPolicy.CommonPolicy != "SOMETHING_ELSE" || cfg.Engine.ExitPolicy.Rejected == "" {
		t.Fatalf("exit policy = %+v, want the unknown value plus refusal", cfg.Engine.ExitPolicy)
	}
}

func TestSaveCommonExitPolicyChangesOnlyItsValueBlock(t *testing.T) {
	body := `{
  "schema_version":5,
  "$future":{"keep":"byte for byte"},
  "trading":{"enabled":false},
  "engine":{
    "automation_gate":{"enabled":false},
    "adoption":{"enabled":false},
    "exit_policy":{"common_policy":"COMMON_LADDER_BALANCED"},
    "future_engine_key":7
  }
}`
	svc := writeConfig(t, body)
	if err := svc.SaveEngineExitPolicy(ExitPolicy{CommonPolicy: exitpolicy.CommonLadderHybrid50}); err != nil {
		t.Fatalf("SaveEngineExitPolicy: %v", err)
	}
	raw, err := svc.LoadRawEngineExitPolicy()
	if err != nil {
		t.Fatal(err)
	}
	if raw.CommonPolicy != exitpolicy.CommonLadderHybrid50 || raw.Rejected != "" {
		t.Fatalf("raw policy = %+v", raw)
	}
	after, _ := os.ReadFile(svc.Path())
	if !strings.Contains(string(after), `"$future":{"keep":"byte for byte"}`) ||
		!strings.Contains(string(after), `"automation_gate":{"enabled":false}`) ||
		!strings.Contains(string(after), `"adoption":{"enabled":false}`) ||
		!strings.Contains(string(after), `"future_engine_key":7`) {
		t.Fatalf("save disturbed an unrelated block:\n%s", after)
	}
}

func TestSaveUnknownCommonExitPolicyChangesNothing(t *testing.T) {
	svc := writeConfig(t, `{"schema_version":5,"trading":{},"engine":{"future":1}}`)
	before, _ := os.ReadFile(svc.Path())
	if err := svc.SaveEngineExitPolicy(ExitPolicy{CommonPolicy: "UNKNOWN"}); err == nil {
		t.Fatal("unknown policy was saved")
	}
	after, _ := os.ReadFile(svc.Path())
	if string(after) != string(before) {
		t.Fatal("refused policy save changed config bytes")
	}
}
