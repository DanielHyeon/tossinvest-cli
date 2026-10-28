# Function Logic Map: `TestCandidateFilterCardsContainExactMarketMetricMatrix`

- Source: `internal/console/optimization_candidate_filters_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- 출처: `archive/2026-10-27-a172-console-anti-ai-redesign` 번들 복사, source sha 대조 일치 — a170 종결 주석(2026-10-28): a170 base `0d168e0a3` 가 그 change 착지보다 앞이라 창(base→워킹트리)에 들어온 함수. `sha256sum internal/console/optimization_candidate_filters_test.go` = `488fe6d51591…` 이 번들의 `source_sha256` 과 같아 번들을 그대로 복사함.
- 출처: `archive/2026-10-26-a175-optimization-and-runtime-matrix-views` 번들 복사, source sha 대조 일치 — a172 종결 주석(2026-10-28): a172 base `fd6afe23b` 가 그 change 착지보다 앞이라 창(base→워킹트리)에 들어온 함수. `sha256sum internal/console/optimization_candidate_filters_test.go` = `488fe6d51591…` 이 번들의 `source_sha256` 과 같아 번들을 그대로 복사함.
- 출처: `archive/2026-10-28-a170-inspection-guard-reconciles-offline-assessment/analysis/function-logic/internal-console--testcandidatefiltercardscontainexactmarketmetricmatrix` 번들 복사, source sha 대조 일치 — a164 종결 주석(2026-10-28): a164 base `4cd7af19d` 가 그 함수를 고친 형제 착지보다 앞이라 창(base→워킹트리)에 들어온 함수. `sha256sum internal/console/optimization_candidate_filters_test.go` = `488fe6d51591…` 이 번들의 `source_sha256`(revision current)과 같아 번들을 그대로 복사함.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 인증된 후보 필터 HTML | 현재 registry 시장·key를 포함한 읽기 전용 문서 | `newDashboardHarness`, `candidate.CandidateFilterMarkets` | DOM 누락은 `t.Fatal`/`t.Error`로 시험 실패 |
| 행렬 행·열 | registry 시장 수 + key 열, 시장 순서 합집합 key 수 행 | a175 표시 요구 목록 | 행렬 역할·키 누락을 시험 실패로 표시 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `.veto-matrix`가 없거나 복수임 | 없음 | `t.Fatal` | `TestCandidateFilterCardsContainExactMarketMetricMatrix` |
| B2 | native table에 `aria-readonly`가 남음 | 없음 | `t.Error` | 같은 시험 |
| B3 | `scope=col` 머리 수가 registry 시장 수와 다름 | 없음 | `t.Error` | 같은 시험 |
| B4 | `scope=row` 시장 합집합 registry 행 수와 다름 | 없음 | `t.Error` | 같은 시험 |
| B5 | 각 registry key가 한 번이 아니거나 누락됨 | 없음 | `t.Error` | 같은 시험 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `newDashboardHarness`·`h.page` | 인증된 로컬 렌더 수집 | in-process, retry 없음 | `ast.json` calls |
| `candidate.CandidateFilterMarkets`·`candidateFilterMatrixFromMarkets` | 현재 시장·필터 registry에서 기대 열·합집합 행·key를 유도 | in-process, retry 없음 | `ast.json` calls |
| `candidateFilterNode`·DOM helper | 실제 `<table>` 역할·scope·행과 native semantics를 확인 | parse 실패는 즉시 시험 실패 | `ast.json` calls |

## State mutations and fallbacks

- 테스트 fixture와 DOM만 읽음. 주문·설정·레지스트리 mutation 없음.

## Safety conclusion

- Safe edit boundary: 카드 개수 검사를 registry 유도 행렬·native table semantics·key 검사로 교체함.
- High-risk impact: no — UI 계약 회귀를 잡는 시험이며 production 동작을 바꾸지 않음.
