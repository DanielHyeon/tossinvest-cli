package config

// a208 design 8.2 T7 · T8 — engine.adoption.floor_only_exit 는 기본 거짓이고, 키 없음 · 명시 false 는 거짓,
// 참은 병합 · 원문 읽기 · 저장 모두에서 보존됨. 거짓이면 저장 바이트에 키가 없음(§0.2 — a208 이전과 같은 파일).
// 거부된 블록은 지금처럼 전면 0 이라 토글도 거짓.

import (
	"os"
	"testing"
)

// T7 — 키 없음 → 거짓 · 저장 바이트에 키 없음, 참 → merge · LoadRaw · Save 모두 보존.
func TestA208FloorOnlyExitRoundTrips(t *testing.T) {
	t.Run("absent key is off and a save does not write it", func(t *testing.T) {
		// 블록 값이 a208 이전 저장이 쓰는 직렬화 그대로라, 같은 블록을 다시 저장하면 파일 바이트가 한 글자도 안 바뀌어야 함.
		body := `{"schema_version":5,"trading":{},
		  "engine":{"adoption":{"enabled":true,"default_stop_pct":0.05}}}`
		svc := writeConfig(t, body)
		cfg, err := svc.Load(t.Context())
		if err != nil {
			t.Fatalf("Load: %v", err)
		}
		if cfg.Engine.Adoption.FloorOnlyExit {
			t.Fatal("an absent floor_only_exit key must load as false")
		}
		raw, _, err := svc.LoadRawEngineAdoption()
		if err != nil || raw.FloorOnlyExit {
			t.Fatalf("raw = %+v, err = %v, want false", raw, err)
		}
		if err := svc.SaveEngineAdoption(raw); err != nil {
			t.Fatalf("SaveEngineAdoption: %v", err)
		}
		after, err := os.ReadFile(svc.Path())
		if err != nil {
			t.Fatal(err)
		}
		if string(after) != body {
			t.Fatalf("an off toggle changed the saved bytes — the file is no longer what a pre-a208 save wrote:\n%s", after)
		}
	})

	t.Run("explicit false is off", func(t *testing.T) {
		svc := writeConfig(t, `{"schema_version":5,"trading":{},
		  "engine":{"adoption":{"enabled":true,"default_stop_pct":0.05,"floor_only_exit":false}}}`)
		cfg, err := svc.Load(t.Context())
		if err != nil || cfg.Engine.Adoption.FloorOnlyExit {
			t.Fatalf("explicit false: cfg = %+v, err = %v", cfg.Engine.Adoption, err)
		}
	})

	t.Run("true survives merge, raw read and save", func(t *testing.T) {
		svc := writeConfig(t, `{"schema_version":5,"trading":{},
		  "engine":{"adoption":{"enabled":true,"default_stop_pct":0.05,"floor_only_exit":true}}}`)
		cfg, err := svc.Load(t.Context())
		if err != nil {
			t.Fatalf("Load: %v", err)
		}
		if !cfg.Engine.Adoption.FloorOnlyExit {
			t.Fatal("merge dropped floor_only_exit=true")
		}
		raw, verdict, err := svc.LoadRawEngineAdoption()
		if err != nil || verdict != "" {
			t.Fatalf("LoadRawEngineAdoption: %v verdict=%q", err, verdict)
		}
		if !raw.FloorOnlyExit {
			t.Fatal("the raw read dropped floor_only_exit=true")
		}
		if err := svc.SaveEngineAdoption(raw); err != nil {
			t.Fatalf("SaveEngineAdoption: %v", err)
		}
		again, err := svc.Load(t.Context())
		if err != nil {
			t.Fatalf("Load after save: %v", err)
		}
		if !again.Engine.Adoption.FloorOnlyExit {
			t.Fatal("the save dropped floor_only_exit=true")
		}
	})
}

// T8 — 손절폭 범위 밖 + 참 → 거부 블록은 전면 0(토글도 거짓).
func TestA208ARefusedBlockClearsFloorOnly(t *testing.T) {
	svc := writeConfig(t, `{"schema_version":5,"trading":{},
	  "engine":{"adoption":{"enabled":true,"default_stop_pct":0.005,"floor_only_exit":true}}}`)
	cfg, err := svc.Load(t.Context())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Engine.Adoption.Rejected == "" {
		t.Fatal("a 0.005 fraction must be refused")
	}
	if cfg.Engine.Adoption.FloorOnlyExit {
		t.Fatal("a refused block kept floor_only_exit — a refused block is zeroed in full")
	}
	// 원문 읽기는 거부 블록도 쓰인 그대로 돌려줌(콘솔이 고쳐 저장할 수 있게) — 토글 포함.
	raw, verdict, err := svc.LoadRawEngineAdoption()
	if err != nil || verdict == "" || !raw.FloorOnlyExit {
		t.Fatalf("raw = %+v verdict = %q err = %v, want the block as written with its verdict", raw, verdict, err)
	}
}
