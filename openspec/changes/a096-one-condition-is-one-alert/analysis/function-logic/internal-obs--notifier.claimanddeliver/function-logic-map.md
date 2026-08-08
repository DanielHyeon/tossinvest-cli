# Function Logic Map: `Notifier.claimAndDeliver`

- Source: `internal/obs/notifier.go` (220–241)
- AST evidence: `ast.json` (sha256 `c9dee3479706…`, 2분기, 반환 3곳)
- Risk scan: `risk-pattern-report.md`
# Function Logic Map: `Notifier.claimAndDeliver`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `record` | critical alert 한 건 | `notifyCritical`이 조립 | — |
| `n.mu` | 이 함수가 잡는다 | — | claim부터 send까지 보유 |
| `n.remindAfter()` | >0 | `RemindAfter` 또는 `DefaultRemindAfter`(1시간) | 0 이하면 기본값 |

불변식: **claim과 send가 한 잠금 안에 있다.** 이것이 이 함수의 존재 이유 전부다.
claim이 잠금 밖에 있으면 같은 조건을 관측한 두 goroutine이 아직 전달되지 않은 같은 행을
읽고, 둘 다 owed로 판정하고, 둘 다 발행한다. 두 번째는 이미 DELIVERED인 행을 표시하려다
실패하며, 그것이 2026-08-08 로그에 53번 찍힌 `no such alert`다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1@227 | claim이 오류 | 없음 | `false, false, err` @228 | 없음(원장 실패 주입 하니스 없음) |
| B2@230 | `!owed` | **없음 — `deliver`를 부르지 않는다** | `false, false, nil` @238 | `TestOneConditionIsOneSend` |
| — | owed | `deliver` → `Publish` | `sent, true, nil` @240 | 기존 다수 |

반환이 `(sent, owed, err)` 두 bool인 이유는 호출자가 **구별해야 하기** 때문이다:
보내지 않은 것이 "필요 없어서"인지 "실패해서"인지에 따라 escalate 여부가 갈린다.
필요 없던 전송은 전달 실패가 아니므로 진입을 차단해서는 안 된다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `n.Journal.ClaimAlertForDelivery` | 기록 + 전달 필요 여부 + 창 지난 행의 재무장 | 오류는 호출자로 | AST :226 |
| `n.remindAfter` | 창 | 없음 | AST :226 |
| `n.deliver` | 재시도 예산 안의 전송 | 소진 시 false | AST :240 |

호출자(CodeGraph): `Notifier.notifyCritical` 한 곳(`notifier.go:194`).
`escalate`는 **이 함수 밖**에 있다 — 잠금 재진입 교착을 피하기 위해서다.

## State mutations and fallbacks

- 이 함수 자체는 원장에 쓰지 않는다. 쓰기는 `ClaimAlertForDelivery`(행 생성·재무장)와
  `deliver` 안의 표시가 한다.
- 잠금 보유 시간은 claim 한 트랜잭션 + 전송 예산(최대 3회×`RetryDelay`)이다. 1판 대비
  늘어난 것은 claim 한 트랜잭션분이며, 전송 예산은 이전에도 같은 잠금 안에 있었다.

## Safety conclusion

- Safe edit boundary: 이 함수 전체가 a096 2판의 신규 코드다.
- High-risk impact: **no** — 주문·손절·익절·사이징·체결 경로에 닿지 않는다.
- 잠금이 관측 사이클을 붙잡는 문제는 a092의 소관이며 a096이 새로 만든 것이 아니다.
  a092가 착지하면 이 경계를 비동기 경로로 옮기는 판단이 필요하다(tasks 7.5).
