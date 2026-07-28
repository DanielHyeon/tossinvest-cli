# Function Logic Map: `newlyListed`

- Source: `internal/candidatesrc/candidatesrc.go`
- Change: `fix-chase-veto-measurement`
- AST evidence: `ast.json` (revision `current`, L340–348, 분기 2개)
- Risk scan: `risk-pattern-report.md`

**신규 함수**. 한 심볼에 대한 3-상태 답이다.

직전 읽기가 없으면 `unknown`이고, 그것이 design D2의 전부다 — "직전 읽기에 없었다"는 첫
읽기의 **모든** 행에 대해 공허하게 참이므로, 그것을 `yes`로 적으면 프로세스가 시작될 때마다
패널 전체가 신규 진입으로 선언된다. 이 자리에 있던 `bool`은 그 구분을 말할 수 없었다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `previous` | 직전 읽기의 심볼 집합 또는 nil | `rememberRead` | nil은 `had=false`와 짝 |
| `had` | 그 집합을 써도 되는가 | `previousReading.usableAt` | false면 `unknown` |
| `symbol` | 행이 보고되는 식별자 | `Read`(공식) / `wtsSymbol`(WTS) | 집합의 키와 **같은 문자열**이어야 한다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `!had` | 없음 | `candidate.NewlyListedUnknown()` | `TestASourcesFirstReadingHasNoAnswerAboutNewEntrants` |
| B2 | `previous[symbol]` | 없음 | `NewlyListedNo()` / 아니면 `NewlyListedYes()` | `TestTheSecondReadingSeparatesTheSymbolsThatJoinedFromTheOnesThatStayed` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| (없음) | — | — | ast.json calls=null |

## State mutations and fallbacks

- 없음. 순수 함수.
- fallback 없음 — 미상은 zero value이므로 누락이 정직한 답을 낸다.

## Safety conclusion

- Safe edit boundary: 신규 함수. 기존 동작 없음.
- High-risk impact: no (순수 판정). `unknown`을 `no`로 접으면 `MeasureFirstSighting`의 D3 거부가 통째로 무력화된다 — 이 세 줄이 그 거부의 입력 전부다.
