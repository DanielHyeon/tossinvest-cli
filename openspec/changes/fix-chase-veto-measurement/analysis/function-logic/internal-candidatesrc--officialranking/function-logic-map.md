# Function Logic Map: `OfficialRanking`

- Source: `internal/candidatesrc/candidatesrc.go`
- Change: `fix-chase-veto-measurement`
- AST evidence: `ast.json` (revision `current`, L208–224, 분기 3개)
- Risk scan: `risk-pattern-report.md`

공식 순위 어댑터의 생성자. 이 change가 **`clk clock.Clock` 인자 하나**를 더했다(리뷰 F1) —
기억의 나이를 잴 시계가 없으면 `previousReading`의 두 조건 중 하나가 성립할 수 없다.

`count` 상한 100은 기존 동작이고, 이 change에서 **의미가 하나 늘었다**: `Row.RankRequested`로
실려 나가는 것이 호출자의 인자가 아니라 `o.count`(상한 적용 후)여야 한다. 500을 요청해
100으로 잘린 호출자가 500을 보고하면 모든 읽기가 영구히 절단으로 판정된다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `reader` | `RankingReader` — `Rankings` 한 메서드 | 호출자 배선 | nil이면 `Read`에서 패닉 — 생성자는 검사하지 않는다(기존 동작) |
| `budget` | `BudgetReader` 또는 nil | official client | nil은 예산 미보고 |
| `typ` | `rankingSourceID`의 세 키 | 이 파일의 상수 | 미등록이면 error |
| `count` | 1..100, 그 밖은 100 | 이 함수 | 상한 적용 후 값이 `o.count`가 된다 |
| `clk` | `clock.Clock` 또는 nil | 호출자 | nil은 `clock.System()` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `!ok` — `rankingSourceID`에 없는 ranking type | 없음 | `nil, error` | `TestAnUnknownRankingTypeIsRefused` |
| B2 | `count <= 0 || count > 100` | `count = 100` | — | `TestTheRequestedCountIsCappedAtTheDocumentedMaximum`(500→100) · `TestTheRequestedCountIsTheCappedOneRatherThanTheOneTheCallerAsked` |
| B3 | `clk == nil` | `clk = clock.System()` | — | `TestTheOfficialRankingAsksForTheRealtimeList` 등 nil을 넘기는 모든 테스트 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `clock.System` | 기본 시계 | 오류 없음 | ast.json calls |
| `fmt.Errorf` | 미등록 type 거부 | — | ast.json calls |

## State mutations and fallbacks

- `officialRanking` 하나를 새로 만든다. `seen` 맵은 lazily 생성되므로 여기서는 nil이다 — nil이 곧 '직전 읽기 없음'이고 그것이 `unknown`의 zero value다.
- fallback: `clk == nil` → 시스템 시계. 이것은 미측정을 값으로 접는 것이 아니라 주입 지점의 기본값이다.

## Safety conclusion

- Safe edit boundary: 시그니처 가산(`clk`) + 문서. 본문 로직 무변경.
- High-risk impact: no (읽기 전용 어댑터 생성). `count` 상한이 풀리면 `RankRequested`가 도달할 수 없는 수가 되어 모든 읽기가 절단 판정된다 — 노출을 줄이는 방향이라 안전 실패다.
