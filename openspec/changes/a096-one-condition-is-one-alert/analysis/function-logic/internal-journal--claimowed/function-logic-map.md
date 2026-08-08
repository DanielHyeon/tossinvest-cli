# Function Logic Map: `claimOwed`

- Source: `internal/journal/outbox.go` (237–268)
- AST evidence: `ast.json` (sha256 `c6612c641a3a…`, 7분기, 반환 6곳)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `state` | PENDING·DELIVERED·ACKNOWLEDGED, **또는 그 밖** | `alert_outbox.state` — **CHECK 제약이 없다** | B7 default: owed |
| `deliveredAt` | RFC3339 또는 NULL | `MarkAlertDelivered` | 파싱 실패는 없는 것으로 친다 |
| `acknowledgedAt` | RFC3339 또는 NULL | `AcknowledgeAlert` | 위와 같음 |
| `now` | 주입 시계 | `Journal.clk` | — |
| `remindAfter` | ≤0이면 재무장 없음 | 호출자의 정책 | — |

불변식: **`owed`는 "운영자가 아직 못 받았을 수 있다"를 뜻한다.** 확실히 받았다고
말할 수 있을 때만 false다. 그래서 모르는 상태는 owed이고, 날짜를 못 읽는 종결 행도 owed다.

`AcknowledgeAlert`의 UPDATE가 `WHERE state = 'PENDING'`이므로 ACKNOWLEDGED는 DELIVERED를
거치지 않고 PENDING에서 온다(`outbox.go:353-357`). 두 상태는 "이 episode의 빚은 끝났다"를
같은 강도로 말하며, 그래서 같은 창을 쓴다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1@243 | `switch state` | 없음 | — | 열거만 |
| B2@244 | `state == AlertPending` | 없음 | `true, false` @246 | `TestClaimingAFreshAlertOwesDelivery` |
| B3@247 | `state == Delivered\|\|Acknowledged` | 없음 | 아래 | — |
| B4@248 | `remindAfter <= 0` | 없음 | `false, false` @249 | `TestAZeroReminderWindowNeverReArms` |
| B5@252 | 종결인데 읽을 시각이 없다 | 없음 | `true, true` @255 | 없음 — 정상 mutator로 도달 불가 |
| B6@257 | `now - settled < remindAfter` | 없음 | `false, false` @258 | `TestClaimingADeliveredAlertInsideTheWindowOwesNothing` |
| B7@261 | `default` — 모르는 상태 | 없음 | `true, true` @266 | `TestAnUnrecognisedAlertStateOwesDelivery` |

B6을 지나거나 B7이면 `true, true` — owed이면서 **재무장**이다. B7을 재무장하지 않으면
발행 뒤 `MarkAlertDelivered`가 실패해 다음 관측이 다시 발행한다. 재무장은 이 함수가 아니라
호출자가 같은 트랜잭션 안에서 수행한다. 이 함수는 순수하다(부작용 0).

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `latestStamp` | 종결 시각 중 가장 최근 것 | 파싱 실패는 "없음"으로 흡수 | AST :252 |
| `now.Sub` | 창 판정 | 없음 | AST :257 |

호출자(CodeGraph): `Journal.ClaimAlertForDelivery` 한 곳(`outbox.go:194`).

## State mutations and fallbacks

- **없다.** 순수 함수다. 판정과 쓰기를 나눈 것이 이 분리의 목적이다 — 판정은 테스트하기
  쉽고, 쓰기는 트랜잭션 경계 안에 있어야 하며, 둘을 섞으면 어느 쪽도 따로 볼 수 없다.
- `latestStamp`가 `delivered_at`과 `acknowledged_at` 중 **더 나중 것**을 고른다. 재무장이
  이전 episode의 시각을 남겨 두므로, 둘 다 있을 수 있고 그때 창의 기준은 나중 것이다.

## Safety conclusion

- Safe edit boundary: 이 함수 전체가 a096 2판의 신규 코드다.
- High-risk impact: **no** — 순수 판정이며 원장에 쓰지 않는다. 다만 이 함수가 틀리면
  critical 알림이 삼켜지므로, 실질 위험은 **false를 과하게 돌려주는 것**이다.
  그래서 모르는 것은 전부 true로 떨어뜨린다(B5·B7).
