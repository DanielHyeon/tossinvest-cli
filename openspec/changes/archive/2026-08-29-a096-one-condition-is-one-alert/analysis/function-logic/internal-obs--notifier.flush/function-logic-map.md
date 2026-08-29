# Function Logic Map: `Notifier.Flush`

- Source: `internal/obs/notifier.go` (367–402)
- AST evidence: `ast.json` (sha256 `c9dee3479706…`, 6분기, 반환 4곳)
- Risk scan: `risk-pattern-report.md`
# Function Logic Map: `Notifier.Flush`

a096이 이 함수에 더한 것은 **`n.mu` 하나**다. 루프도 SQL도 반환값도 그대로다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `n.Journal` | nil 가능 | 조립 시점 | B1: 아무것도 하지 않고 0을 반환 |
| `n.Publisher` | nil 가능 | 조립 시점 | B3: 루프를 끊는다 |
| PENDING 행 목록 | 오래된 것부터 | `PendingAlerts` | B2: 오류를 그대로 올린다 |
| `n.mu` | **이 함수가 잡는다** | a096 | 루프 전체 동안 보유 |

불변식: **어느 순간에도 하나의 발행 경로만 동작한다.** 1판까지 이 함수는 `n.mu`를
전혀 잡지 않았고, 그래서 관측 경로의 전송과 flush가 같은 행을 같은 순간에 보낼 수 있었다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1@368 | `n.Journal == nil` | 없음 | `0, 0, nil` @369 | 기존 |
| B2@378 | `PendingAlerts` 오류 | 없음 | `0, 0, err` @379 | 없음(미진입) |
| B3@381 | `n.Publisher == nil` | break | — | 없음(미진입) |
| B4@382 | `Publish` 실패 | 시도 실패 기록 후 continue | — | 기존 flush 테스트 |
| B5@391 | 전달 표시 실패 | 없음 | `delivered, 0, merr` @392 | 없음(미진입) |
| B6@395 | `UndeliveredCount` 오류 | 없음 | 그 오류와 함께 반환 | 없음(미진입) |

## 배선에 관한 사실 하나

**`Notifier.Flush`에는 non-test 호출자가 없다.** 확인했다. 따라서
`execgw.Gateway.parkAlert`가 outbox에 넣는 행("the notifier's Flush picks the row up and
delivers it" — `replay.go:101`)은 현재 배선에서 아무도 집어 가지 않는다.

a096이 만든 문제도, a096이 고치는 문제도 아니다. 이 map은 기록만 하며 별도 change가
필요하다(tasks 7.4). 다만 a096이 이 함수에 잠금을 더한 것은 여전히 옳다 — 배선되는 날
경합이 생기는 것과, 지금 안전하게 만들어 두는 것은 비용이 다르다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `n.Journal.PendingAlerts` | 백로그 | B2 | AST :377 |
| `n.Publisher.Publish` | 실제 전송 | 실패는 continue | AST :385 |
| `n.Journal.MarkAlertAttemptFailed` | 실패 기록 | **오류를 버린다** — 아래 | AST :386 |
| `n.Journal.MarkAlertDelivered` | 성공 기록 | 오류면 즉시 반환 | AST :390 |
| `n.Journal.UndeliveredCount` | 남은 수 | B6 | AST :395 |

`MarkAlertAttemptFailed`의 오류를 버리는 것(`_ =`)은 a096 이전부터의 동작이며, 기록에
실패하고도 깨끗한 결과를 보고할 수 있다. 독립 리뷰 1라운드 concern 4이고, a096이 만든
것이 아니므로 여기서 고치지 않는다(tasks 7.4).

## State mutations and fallbacks

- 행마다 DELIVERED 또는 attempts 증가. 되돌림 없음.
- gate는 건드리지 않는다 — 백로그를 비우는 것과 차단을 푸는 것은 다른 일이고,
  후자는 `Acknowledge`가 한다.

## Safety conclusion

- Safe edit boundary: 함수 진입부의 `n.mu.Lock()` 한 쌍.
- High-risk impact: no — 주문·손절·익절·사이징·체결 경로에 닿지 않는다.
- 잠금을 더해 생기는 유일한 비용은 flush와 관측 전송이 서로를 기다린다는 것이며,
  그것이 바로 목적이다.
