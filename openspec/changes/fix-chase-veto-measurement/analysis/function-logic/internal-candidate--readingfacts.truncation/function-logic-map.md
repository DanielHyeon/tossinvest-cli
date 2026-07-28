# Function Logic Map: `ReadingFacts.Truncation`

- Source: `internal/candidate/scan.go`
- Change: `fix-chase-veto-measurement`
- AST evidence: `ast.json` (revision `current`, L170–170, 분기 0개)
- Risk scan: `risk-pattern-report.md`

**신규 메서드**(무분기 한 줄). `ScanResult.Readings`의 두 숫자가 말하는 것을 돌려준다.

`FirstRank.Truncation`과 **같은 `truncationOf`**를 부르는 것이 요점이다. 스캔 리포트가
"whole"이라고 쓴 읽기를 veto가 절단으로 판정하는 불일치가 구조적으로 불가능해진다 —
같은 규칙을 두 번 구현하면 언젠가 갈라지고, 갈라진 결과는 조용해 보이는 화면이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `r.Requested` | 0(미기록) 또는 양수 | `Collect`가 행에서 읽는다 | 0이면 `unknown` |
| `r.Arrived` | 도착한 행 수 | 같은 행의 `RankTotal` | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | (무분기) `truncationOf(r.Requested, r.Arrived)` | 없음 | 3-상태 `Truncation` | `TestTheScanReportSaysWhatEachSourceAskedForAndWhatArrived` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `truncationOf(r.Requested, r.Arrived)` | 요청과 도착의 비교 | 순수 | ast.json calls |

## State mutations and fallbacks

- 없음 — 순수 파생.
- 두 숫자를 비교로 **대체하지 않는다**. 100 요청 99 도착과 100 요청 3 도착은 같은 `Truncation`이고 전혀 다른 대화이므로, 리포트는 두 숫자를 싣고 이 메서드는 라벨만 만든다.

## Safety conclusion

- Safe edit boundary: 신규 메서드. 기존 동작 없음.
- High-risk impact: no (순수 파생, 표시용).
