# Function Logic Map: `Notifier.Flush`

- Source: `internal/obs/notifier.go` (427-462)
- AST evidence: `ast.json` — branches 6, returns 4, calls 9, assignments 3,
  **defers 1, go_statements 0**
- Risk scan: `risk-pattern-report.md`

**17판이 가장 크게 손대는 함수이자, AST가 위험을 먼저 잡은 자리.**
`:424`의 doc comment는 이 함수를 *"감독 루프가 주기적으로 부르는 것"*이라고 적지만,
**프로덕션 호출자는 0이다**(`internal/`·`cmd/` 전체 `rg` 확인 — `Flush()` 일치는
전부 `bufio`/`csv`/`http.Flusher`다). 오늘 이 함수는 테스트에서만 돈다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `n.Journal` | nil이면 무동작 | Notifier 배선 | B1 `:428` → `(0,0,nil)` |
| `n.Publisher` | **nil이면 루프를 끊는다** | 같은 위 | B4 `:442` `break` — **시도를 기록하지 않는다** |
| `n.mu` | 배달 뮤텍스 | 같은 위 | `:434` 획득, `:435` defer 해제 |
| 처리 대상 | `PendingAlerts(ctx, 0)` — **무제한** | `:437` | 밀린 만큼 전부 |
| `ctx` | 호출자의 것 | 오늘은 테스트 | 전송 기한이 여기서 온다 |

**전송 기한이 이 함수 안에 없다.** `Publish`에 넘기는 것은 호출자의 ctx뿐이고,
`:451`은 `n.Publisher.Publish(ctx, msg)` 하나다 — 시도별 상한도, 시도 간 대기도,
재시도 횟수도 없다. **HEAD의 `Flush`는 "행마다 한 번, 기한은 남의 것"이다.**

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return |
|---|---|---|---|
| B1 `:428` | `n.Journal == nil` | 없음 | `:429` `(0, 0, nil)` |
| B2 `:438` | `PendingAlerts` 실패 | 없음 | `:439` `(0, 0, err)` |
| **B3 `:441`** | **`range pending` — 상한 없음** | 행마다 `Publish` | — |
| **B4 `:442`** | **`n.Publisher == nil` → `break`** | **아무것도 기록 안 함** | — |
| B5 `:451` | `Publish` 실패 | `MarkAlertAttemptFailed` `:452` → `continue` `:453` | — |
| B6 `:455` | `MarkAlertDelivered` 실패 | 없음 | `:456` `(delivered, 0, merr)` — **루프 중단** |
| — `:461` | — | `UndeliveredCount` `:460` | `(delivered, remaining, err)` |

**B3 × B5가 17판이 반드시 고쳐야 하는 조합이다.** `n.mu`를 잡은 채(`:434`)
밀린 행 전부를 순회하며 행마다 원격 전송을 기다린다. 밀린 행 9개에 전송 상한
3.5초면 **한 번의 `Flush`가 뮤텍스를 31.5초 쥔다** — 그동안 `claimAndDeliver`
(`:241`이 같은 뮤텍스를 잡는다)는 진행하지 못하고, 그것은 **exit 관측 사이클이
멈춘다**는 뜻이다.

즉 **`Flush`를 그대로 감독 루프에 배선하면 a092가 없애려는 결함이 자리만 옮긴다.**
그래서 17판 spec은 네 가지를 SHALL로 못박는다: 사이클당 작업량 상한,
기록 경로가 배달 잠금을 잡지 않을 것, 행당 주기당 1시도, 해제 경로.

**B4는 별도의 결함이다.** publisher가 없으면 `break`하므로
`MarkAlertAttemptFailed`가 한 번도 불리지 않는다 — `attempts`도 안 늘고
`last_error`도 비어 있다. 운영자가 outbox를 읽으면 "아무도 시도 안 했다"와
"전송기가 없었다"가 구별되지 않는다. 17판은 이것을 **실패한 시도로 센다**로 바꾼다.

**B6도 위험하다.** 이미 `DELIVERED`인 행에 `MarkAlertDelivered`를 부르면
CAS가 0행을 갱신해 `ErrAlertNotFound`가 되고, `:456`이 그것을 **오류로 올려
루프를 끊는다.** 그러면 뒤에 밀린 행들이 그 주기에 처리되지 않는다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `n.mu.Lock` `:434` | 기록 경로와 배제 | **17판이 떼려는 것** | AST calls |
| `n.mu.Unlock` `:435` | **defer** | — | AST defers **1** |
| `n.Journal.PendingAlerts` `:437` | 대상 행 | `limit = 0` — 무제한 | AST calls |
| `EventType` `:446` | 타입 변환 | 순수 | AST calls |
| **`n.Publisher.Publish` `:451`** | **원격 전송 — 유일한 네트워크 호출** | **기한 없음** | AST calls |
| `n.Journal.MarkAlertAttemptFailed` `:452` | 실패 기록 | 오류를 버린다(`_ =`) | AST calls |
| `n.Journal.MarkAlertDelivered` `:455` | 성공 기록 | 오류가 루프를 끊는다 | AST calls |
| `n.Journal.UndeliveredCount` `:460` | 남은 수 | 오류를 반환값에 실어 올린다 | AST calls |

## State mutations and fallbacks

| 대상 | 자리 | 성격 |
|---|---|---|
| `alert_outbox` 행 | `:452`·`:455` | 시도/배달 기록 |
| `n.mu` | `:434`-`:435` | **전 순회 구간 점유** |
| `delivered` 카운터 | `:458` | 지역 |

- fallback: B5의 `continue` 하나 — 한 행의 실패가 다음 행을 막지 않는다.
  **B6에는 그 fallback이 없다**(`return`).
- **게이트 래치도 운영 모드 승격도 이 함수에는 없다.** 그것들은
  `deliver:403`과 `escalate`에 있다. 17판의 배달 루프는 **그 둘도 가져가야 한다** —
  안 가져가면 실패가 결과를 낳지 않는다.

## Safety conclusion

- **Safe edit boundary**: **a092가 편집한다**(§8 GREEN). 17판의 배달 루프는
  이 함수를 그대로 부르지 않고, 여기 있는 순회를 상한 있는 형태로 다시 쓴다.
- **High-risk impact**: yes — 알림 전송·시도 기록·미전달 수의 전부가 여기다.
- **AST가 문서보다 먼저 잡은 것**: 분기 6개, `n.mu` + defer, 행마다 `Publish`,
  `PendingAlerts(ctx, 0)` 무제한. 이 넷을 보기 전의 17판 초안은
  *"Flush를 감독 루프에 배선한다"*였고, **그 배선이 결함을 옮길 뿐이라는 것을
  이 열거가 보여줬다.** 설계 D0의 네 속성은 이 표에서 나왔다.
- **RED로 잡아야 하는 것**: R17-1(사이클이 전송을 안 기다린다),
  R17-2(기록 경로가 배달 잠금을 안 기다린다), R17-4(nil publisher가 시도로
  세어진다), R17-5(사이클당 행 수 상한), R17-6(행당 주기당 1시도).
