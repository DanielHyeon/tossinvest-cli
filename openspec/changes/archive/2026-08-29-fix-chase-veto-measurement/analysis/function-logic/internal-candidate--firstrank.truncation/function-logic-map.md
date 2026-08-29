# Function Logic Map: `FirstRank.Truncation`

- Source: `internal/candidate/store.go`
- Change: `fix-chase-veto-measurement`
- AST evidence: `ast.json` (revision `current`, L1430–1430, 분기 0개)
- Risk scan: `risk-pattern-report.md`

**신규 메서드**(무분기 한 줄). 저장된 요청 행 수가 그 읽기에 대해 말하는 것을 돌려준다.

`truncationOf` **같은 함수**를 쓰는 것이 요점이다. `ScanResult.ReadingFacts.Truncation`도
같은 함수를 쓰므로, 스캔 리포트가 "whole"이라고 말한 읽기를 veto가 절단이라고 판정하는
불일치가 구조적으로 불가능하다 — 이 저장소가 네 번째 그림자 밴드에서 한 번 지불한 형태다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `f.Requested` | 0(미기록) 또는 양수 | 저장 칼럼 | 0이면 `unknown` |
| `f.Total` | 도착한 행 수 | 저장 칼럼 | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | (무분기) `truncationOf(f.Requested, f.Total)` | 없음 | 3-상태 `Truncation` | `TestTruncationComparesTheTwoNumbersTheSourceDeclared` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `truncationOf(f.Requested, f.Total)` | 요청과 도착의 비교 | 순수, 오류 없음 | ast.json calls |

## State mutations and fallbacks

- 없음 — 순수 파생.
- fallback 없음. 미기록은 `unknown`이고 `MeasureFirstSighting`이 그것을 `REQUEST_UNRECORDED`로 거부한다.

## Safety conclusion

- Safe edit boundary: 신규 메서드. 기존 동작 없음.
- High-risk impact: no (순수 파생). 재는 성질은 High-risk 인접 — 이것이 `whole`로 접히면 100행 요청에 3행이 온 읽기의 1위가 백분위 66.7로 측정된다.
