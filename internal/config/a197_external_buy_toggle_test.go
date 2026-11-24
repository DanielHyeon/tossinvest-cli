package config

// a197 design D8 — 외부 반대 매수 취소 토글은 기본 꺼짐이고, 키 없음 · 명시 false 는 꺼짐, automation_gate 키가 없는 파일에서도 읽힘
// (mergeEngine 의 조기 반환 앞). 꺼져 있으면 `tossctl config show` 의 engine 블록 바이트가 편집 전과 같음(omitempty — a163 골든 재사용).

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestA197TheExternalBuyToggleDefaultsOff(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		want       bool
	}{
		{"no engine block", `{"schema_version": 5}`, false},
		{"no key", `{"schema_version": 5, "engine": {"automation_gate": {"enabled": false}}}`, false},
		{"explicit false", `{"schema_version": 5, "engine": {"cancel_external_opposite_buy": false, "automation_gate": {}}}`, false},
		{"explicit true", `{"schema_version": 5, "engine": {"cancel_external_opposite_buy": true, "automation_gate": {}}}`, true},
		{"true without automation_gate", `{"schema_version": 5, "engine": {"cancel_external_opposite_buy": true}}`, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := writeConfig(t, tc.body).Load(context.Background())
			if err != nil {
				t.Fatalf("Load: %v", err)
			}
			if cfg.Engine.CancelExternalOppositeBuy != tc.want {
				t.Fatalf("CancelExternalOppositeBuy = %v, want %v", cfg.Engine.CancelExternalOppositeBuy, tc.want)
			}
		})
	}
	if DefaultFile().Engine.CancelExternalOppositeBuy {
		t.Fatal("a new config must never turn the external-buy cancel on — a human flips it (§0.7)")
	}
}

func TestA197AnOffToggleLeavesConfigShowByteIdentical(t *testing.T) {
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
}

func TestA197ThePublishedSchemaDeclaresTheToggleOffByDefault(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "schemas", "config.schema.json"))
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	engineProps := doc["properties"].(map[string]any)["engine"].(map[string]any)["properties"].(map[string]any)
	prop, ok := engineProps["cancel_external_opposite_buy"].(map[string]any)
	if !ok {
		t.Fatal("the published schema does not declare engine.cancel_external_opposite_buy")
	}
	if prop["type"] != "boolean" || prop["default"] != false || !strings.Contains(prop["description"].(string), "Defaults off") {
		t.Fatalf("schema entry = %+v, want a boolean defaulting off", prop)
	}
}
