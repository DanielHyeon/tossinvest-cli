# Function Logic Map: `Store.Clock`

- Source: `internal/candidate/store.go`
- AST evidence: `ast.json` (revision: current, source_sha256 bound)
- Risk scan: `risk-pattern-report.md`

이 change가 추가한 신규 함수다. watch 루프가 **기다리는** 시계와 **찍는** 시계를 하나로 만들려고 존재한다. 둘이면 루프는 벽시계로 자고 기록은 가짜 시계로 찍혀, 가짜를 앞으로 민 테스트가 루프에 존재하지 않은 주기를 측정한다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `s.clk` | `Options.Clock` 또는 `clock.System()` | `Open` | nil 불가 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| (무분기) | — | 없음 | `s.clk` | `TestTheWatchLoopWaitsOnTheInjectedClock` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| (없음) | — | — | ast.json calls null — 필드 반환 한 줄 |

## State mutations and fallbacks

- 상태 변경 없음.
- `Now()`가 아니라 `Clock` 자체를 내보내는 것이 요점이다 — 호출자가 `Sleep`/`After`를 그 시계 위에서 해야 하고, `TestNothingInThisPackageAsksTheWallClockWhatTimeItIs`가 `time.Sleep`을 포함한 11가지 철자를 막기 때문에 다른 선택지가 없다.

## Safety conclusion

- Safe edit boundary: 반환 타입을 구체 시계로 좁히거나 watch 루프가 별도 시계를 갖게 하는 것은 금지
- High-risk impact: no — 접근자. 주문 경로 무접촉.
