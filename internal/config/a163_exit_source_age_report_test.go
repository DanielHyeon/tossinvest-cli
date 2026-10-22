package config

// a163 tasks 2.13 — 체결 나이 보고 토글은 기본 꺼짐이고, 키 없음 · 명시 false 는 꺼짐, automation_gate 키가 없는 파일에서도 읽힘
// (design D8 「저장」 · 「merge 자리」). 꺼져 있으면 `tossctl config show` 의 engine 블록 바이트가 편집 전과 같음(omitempty).
// 골든은 편집 전 HEAD 의 Status.Engine JSON 바이트임.

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

const a163EngineStatusGolden = "{\"autostart\":false,\"automation_gate\":{\"enabled\":false},\"adoption\":{\"enabled\":false},\"exit_policy\":{},\"notifications\":{\"enabled\":false}}"

func TestA163TheSourceAgeReportToggleDefaultsOff(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		want       bool
	}{
		{"no engine block", `{"schema_version": 5}`, false},
		{"no key", `{"schema_version": 5, "engine": {"automation_gate": {"enabled": false}}}`, false},
		{"explicit false", `{"schema_version": 5, "engine": {"exit_source_age_report": false, "automation_gate": {}}}`, false},
		{"explicit true", `{"schema_version": 5, "engine": {"exit_source_age_report": true, "automation_gate": {}}}`, true},
		// mergeEngine 의 automation_gate 조기 반환 뒤에 두면 조용히 무시되는 모양(변이 ⑲).
		{"true without automation_gate", `{"schema_version": 5, "engine": {"exit_source_age_report": true}}`, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := writeConfig(t, tc.body).Load(context.Background())
			if err != nil {
				t.Fatalf("Load: %v", err)
			}
			if cfg.Engine.ExitSourceAgeReport != tc.want {
				t.Fatalf("ExitSourceAgeReport = %v, want %v", cfg.Engine.ExitSourceAgeReport, tc.want)
			}
		})
	}
	if DefaultFile().Engine.ExitSourceAgeReport {
		t.Fatal("a new config must never turn the report on — a human flips it (§0.7)")
	}
}

func TestA163AnOffToggleLeavesConfigShowByteIdentical(t *testing.T) {
	status, err := writeConfig(t, `{"schema_version": 5}`).Status(context.Background())
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	got, err := json.Marshal(status.Engine)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != a163EngineStatusGolden {
		t.Fatalf("config show engine block moved with the toggle off:\n got: %s\nwant: %s", got, a163EngineStatusGolden)
	}
	on, err := writeConfig(t, `{"schema_version": 5, "engine": {"exit_source_age_report": true}}`).Status(context.Background())
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if shown, _ := json.Marshal(on.Engine); !json.Valid(shown) || !containsKey(t, shown, "exit_source_age_report") {
		t.Fatalf("an on toggle is not shown: %s", shown)
	}
}

func containsKey(t *testing.T, raw []byte, key string) bool {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	_, ok := m[key]
	return ok
}

func TestA163ThePublishedSchemaDeclaresTheToggleOffByDefault(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "schemas", "config.schema.json"))
	if err != nil {
		t.Fatalf("reading the published schema: %v", err)
	}
	var schema struct {
		Properties map[string]struct {
			Properties map[string]struct {
				Type    string          `json:"type"`
				Default json.RawMessage `json:"default"`
			} `json:"properties"`
		} `json:"properties"`
	}
	if err := json.Unmarshal(data, &schema); err != nil {
		t.Fatalf("parsing the published schema: %v", err)
	}
	prop, ok := schema.Properties["engine"].Properties["exit_source_age_report"]
	if !ok || prop.Type != "boolean" || string(prop.Default) != "false" {
		t.Fatalf("schema engine.exit_source_age_report = type %q default %s (declared %v), want a boolean defaulting to false — the engine block is additionalProperties:false", prop.Type, prop.Default, ok)
	}
}
