# Function Logic Map: `TestCandidateFilterCardsRemainMobileAndAccessibilityFriendly`

- Source: `internal/console/optimization_candidate_filters_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 후보 필터 section HTML | 읽기 전용, 390px에서도 셀 접근 가능 | `newDashboardHarness`와 a175 CSS | 표 scope·scroll 역할 누락은 시험 실패 |
| 행렬 표 | `.veto-matrix` 1개, 열·행 머리 scope 보유 | a175 design §2 | 카드 전용 검사 제거 후 실제 접근성 역할 검사 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `.veto-matrix`가 없거나 복수임 | 없음 | `t.Fatal` | `TestCandidateFilterCardsRemainMobileAndAccessibilityFriendly` |
| B2 | `scope=col` 머리 수가 registry 시장 수와 다름 | 없음 | `t.Error` | 같은 시험 |
| B3 | `scope=row` 행 머리 수가 시장 합집합 key 수와 다름 | 없음 | `t.Error` | 같은 시험 |
| B4 | 접근성·키보드 marker 반복 | 없음 | `t.Error` | 같은 시험 |
| B5 | 시장 deep link·region·`dl` marker 누락 | 없음 | `t.Error` | 같은 시험 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `newDashboardHarness`·`h.page` | 로컬 DOM 획득 | in-process, retry 없음 | `ast.json` calls |
| `candidateFilterNode`·DOM helper | 표·scope·region을 구조적으로 확인 | parse 실패는 즉시 시험 실패 | `ast.json` calls |

## State mutations and fallbacks

- 테스트 fixture와 HTML DOM만 읽음. 외부 요청·계정·설정 mutation 없음.

## Safety conclusion

- Safe edit boundary: "표 0개" 고정 리터럴을 시장 합집합으로 유도한 행렬 표의 scope·스크롤 역할 검증으로 교체함.
- High-risk impact: no — 변경된 마크업 역할을 정직하게 검증함.
