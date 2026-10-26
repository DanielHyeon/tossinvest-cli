# Function Logic Map: `TestCandidateFilterCardsContainExactMarketMetricMatrix`

- Source: `internal/console/optimization_candidate_filters_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 인증된 후보 필터 HTML | KR·US와 3 registry key를 포함한 읽기 전용 문서 | `newDashboardHarness`, `candidate.CandidateFilterMarkets` | DOM 누락은 `t.Fatal`/`t.Error`로 시험 실패 |
| 행렬 행·열 | key 3, 시장 2, 공통 행은 값 동일 시만 병합 | a175 표시 요구 목록 | 카드 수·시장별 값 혼동을 시험 실패로 표시 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 문서에 후보 필터 section이 없음 | 없음 | `t.Fatal` | `TestCandidateFiltersRenderAsMatrix` |
| B2 | 행렬 key·시장 수가 registry와 다름 | 없음 | `t.Error` | `TestCandidateFiltersRenderAsMatrix` |
| B3 | KR·US 값·사유가 빠지거나 공통/시장별 표시가 섞임 | 없음 | `t.Error` | `TestCandidateFiltersRenderAsMatrix` 및 서로 다른 시장 값 단위 시험 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `newDashboardHarness`·`h.page` | 인증된 로컬 렌더 수집 | in-process, retry 없음 | `ast.json` calls |
| `candidateFilterNode`·DOM helper | 실제 `<table>` 역할과 셀을 확인 | parse 실패는 즉시 시험 실패 | `ast.json` calls |

## State mutations and fallbacks

- 테스트 fixture와 DOM만 읽음. 주문·설정·레지스트리 mutation 없음.

## Safety conclusion

- Safe edit boundary: 카드 개수 검사만 행렬의 실제 행·열·사유 검사로 교체함.
- High-risk impact: no — UI 계약 회귀를 잡는 시험이며 production 동작을 바꾸지 않음.
