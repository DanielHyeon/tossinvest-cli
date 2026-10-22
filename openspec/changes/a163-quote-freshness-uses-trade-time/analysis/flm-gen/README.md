# a163 FLM 번들 생성 입력

`gen.py` 는 `analysis/function-logic/*/ast.json`(tools/logic-map 산출)과 소스에서 분기 좌표 · 원문을 **기계적으로** 읽고, 의미 칸(입력 ·
호출 · 편집 · 시험 인용)만 `spec-*.json` 에서 채워 `function-logic-map.md` · `branch-test-map.md` 를 씀. 재생성(저장소 루트):

```bash
python3 -I openspec/changes/a163-quote-freshness-uses-trade-time/analysis/flm-gen/gen.py . \
  openspec/changes/a163-quote-freshness-uses-trade-time/analysis/flm-gen/spec-current.json
```

- `spec-base.json` — 편집 전 판(커밋 `78efe1951`, 소스 = base `f7d80c348` 와 `internal`·`cmd` diff 0).
- `spec-current.json` — 편집 뒤 판(GREEN 로트). `ast.json` 은 `GOFLAGS=-trimpath go run ./tools/logic-map --file … --func …` 로 다시 뽑음.

## 편집 전 · 뒤 AST 대조 (design D7 · tasks 1.1 단언)

| 함수 | 분기 | 반환 | defer | 범위(줄) |
| --- | --- | --- | --- | --- |
| `official.adaptPrices` | 1 → 1 | 1 → 1 | 0 → 0 | 168–180 → 168–183 |
| `ExitObserver.observe` | 15 → 15 | 6 → 6 | 0 → 0 | 842–913 → 860–931 |
| `ExitObserver.ObserveOnce` | 8 → 8 | 5 → 5 | 2 → 2 | 457–526 → 469–541 |
| `cmd/tossctl.engineRuntime` | 6 → 6 | 7 → 7 | 0 → 0 | 655–750 → 655–752 |
| `config.mergeEngine` | **5 → 6** | 2 → 2 | 0 → 0 | 324–358 → 330–368 |
| `engine.recordGateSettings` | 6 → 6 | 2 → 2 | 0 → 0 | 739–811 → 739–813 |
| `engine.Context.ExitObserver` | 6 → 6 | 4 → 4 | 0 → 0 | 319–362 → 319–364 |
| `engine.NewContext` | 15 → 15 | 11 → 11 | 0 → 0 | 437–607 → 437–607 |
| a074 시험 함수 넷 | 4·4·3·5 → 같음 | 0 | 0 | 같음 |

`mergeEngine` 재번호: 새 분기 `if raw.ExitSourceAgeReport != nil {` 가 B4 자리에 들어와 옛 B4(`if raw.AutomationGate == nil {` 조기 반환)·
옛 B5(`if gate.Enabled != nil {`)가 B5·B6 이 됨 — 줄 원문으로 대조(번호는 위치라 손 재번호 금지 규칙, 기억 「분기 번호는 위치다」).
