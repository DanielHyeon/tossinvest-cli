# Function Logic Map: `fakePerformanceReader.Dashboard`

- Source: `internal/console/performance_history_test.go` (20-28) — 시험 전용 가짜 리더(생산 코드 아님). base 대비 오류 주입 분기 1 추가.
- Qualified: `fakePerformanceReader.Dashboard`
- AST evidence: `ast.json` (편집 뒤 추출, `source_sha256` 4b9faf0c3668d9b2…) — AST branches 1 · returns 2
- Risk scan: `risk-pattern-report.md` (configured risk pattern 0)
- 출처: `archive/2026-10-27-a172-console-anti-ai-redesign` 번들 복사, source sha 대조 일치 — a170 종결 주석(2026-10-28): a170 base `0d168e0a3` 가 그 change 착지보다 앞이라 창(base→워킹트리)에 들어온 함수. `sha256sum internal/console/performance_history_test.go` = `4b9faf0c3668…` 이 번들의 `source_sha256` 과 같아 번들을 그대로 복사함.
- 출처: `archive/2026-10-25-a173-console-ux-common-cleanup` 번들 복사, source sha 대조 일치 — a172 종결 주석(2026-10-28): a172 base `fd6afe23b` 가 그 change 착지보다 앞이라 창(base→워킹트리)에 들어온 함수. `sha256sum internal/console/performance_history_test.go` = `4b9faf0c3668…` 이 번들의 `source_sha256` 과 같아 번들을 그대로 복사함.
- a173 종결 주석(2026-10-25): a173 base `b27a1ea87` 가 a177 착지보다 앞이라 a173 창(base→HEAD)에 들어온 함수. a173 의 편집은 0 이고 소스 sha 가 a177 번들과 같아(`sha256sum` 대조) a177 아카이브 번들(`openspec/changes/archive/2026-10-25-a177-performance-window-single-source/analysis/function-logic/`)을 그대로 복사함.
- 출처: `archive/2026-10-28-a170-inspection-guard-reconciles-offline-assessment/analysis/function-logic/internal-console--fakeperformancereader.dashboard` 번들 복사, source sha 대조 일치 — a164 종결 주석(2026-10-28): a164 base `4cd7af19d` 가 그 함수를 고친 형제 착지보다 앞이라 창(base→워킹트리)에 들어온 함수. `sha256sum internal/console/performance_history_test.go` = `4b9faf0c3668…` 이 번들의 `source_sha256`(revision current)과 같아 번들을 그대로 복사함.

**역할.** 콘솔 시험의 `PerformanceReader` 가짜. 호출 수 · 받은 질의를 기록하고 고정 view 를 돌려줌. a177 은 `err` 필드와
`if f.err != nil { return performance.DashboardView{}, f.err }` 를 더해 LoadErr 갈래를 시험할 수 있게 함.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `f.err` | nil(기존 시험 전부) 또는 주입 오류 | 시험 픽스처 | 비-nil → 영값 view + 그 오류 |
| `query` | 핸들러가 넘긴 질의 | 호출자 | 기록만 함 |

## Branches and early returns

| Branch | Source line | AST kind | Condition / 효과 | a177 |
|---|---|---|---|---|
| B1 | `24:2` | if | `f.err != nil` → 영값 view + 오류 반환 | 추가 |

Exact AST return positions: 25:3, 27:2

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| (없음) | 필드 대입만 | — | AST calls 0 |

## State mutations and fallbacks

- `f.calls` 증가 · `f.query` · `f.view.Query` 기록(오류 분기 앞이라 오류 주입 시에도 기록됨). 기존 시험은 `err` 를 안 채우므로 동작 동일.

## Safety conclusion

- Safe edit boundary: 시험 파일 안 가짜. 생산 빌드에 들어가지 않음.
- High-risk impact: no.
