# CodeGraph baseline — a113

- 날짜: 2026-09-25 · 기준 HEAD `54004f44` · codegraph **1.6.0** (`.mcp.json` `--no-watch`, 갱신자는 PostToolUse sync 훅)

| 질의 | 결과 |
|---|---|
| `codegraph callers projectionSocketAccepts` | 3: `TestProjectionLivenessClausesEachDecideOnTheirOwn`(a108_publication_is_total_test.go:453) · `reclaimStaleControlDirectory`(transport_unix.go:243) · `Dial`(:402) |
| `codegraph callees projectionSocketAccepts` | 2: `projectionProbeTimeout`(:29) · `Close` → **`internal/journal/readonly_schema_arms_test.go:51`** (오해석 — reconciliation 참조) |
| `codegraph callers verifyStaleSocketShape` | 1: `reclaimStaleControlDirectory` |
| `codegraph callers reclaimStaleControlDirectory` | 1: `Start`(:79) |
| `codegraph impact projectionSocketAccepts` | 5: projectionSocketAccepts · reclaimStaleControlDirectory · Start · Dial · 판정 표 테스트 — 패키지 밖으로 나가지 않는다 |
| `codegraph affected transport_unix.go` (기본) | 1: `auth-helper/tests/test_cli.py` — Go 테스트 0 (하네스 주석의 `isTestPath` 결함 그대로) |
| `codegraph affected transport_unix.go --filter '*_test.go'` | 3: `a108_leftover_recovery_test.go` · `a108_publication_is_total_test.go` · `transport_unix_test.go` |
