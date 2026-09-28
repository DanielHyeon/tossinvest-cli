# Function Logic Map: `Gateway.parkAlert`

- Source: `internal/execgw/replay.go` (534-559)
- AST evidence: `ast.json` — branches 2, returns 0, calls 5, assignments 2,
  **defers 0, go_statements 0**
- Risk scan: `risk-pattern-report.md`

**살아 있는 결함의 자리.** a092는 이 함수를 편집하지 않지만, **이 함수가 만든 행을
아무도 보내지 않는다는 사실**이 17판이 고치는 것 중 하나다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `g.entry` | nil일 수 있다 | Gateway 배선 | B1 `:535` — nil이면 래치를 건너뛴다 |
| `rec.ID` | attempt id | `journal.AttemptRecord` | 없으면 event key가 잘린다 |
| `intent.Symbol`·`Market`·`AccountRef` | 주문 의도 | `journal.Intent` | 검증 없음 |
| `detail` | 사람이 읽는 사유 | 호출자 | 그대로 `Body`로 |
| `ctx` | 호출자의 것 | replay 경로 | `EnqueueAlert`가 쓴다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return |
|---|---|---|---|
| B1 `:535` | `g.entry != nil` | **`entry.Block(ReasonUnresolvedInDoubt, …)` `:536`** | 없음 (void) |
| B2 `:548` | `json.Marshal` 실패 | `payload = nil` `:549` — **계속 진행한다** | 없음 |
| — `:551` | — | `EnqueueAlert` — **반환값 둘 다 버린다** | 없음 (void) |

`return` 문이 0개다(AST returns null). **이 함수는 실패를 보고할 수 없다.**
`:551`의 `_, _ =`가 그것을 명시한다. doc comment `:530-533`이 근거를 적었다 —
park 자체는 이미 내구적이고, 알림을 잃는 것은 나쁘지만 안전하지 않은 것은 아니다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `g.entry.Block` `:536` | 진입 래치 | 실패 불가 (void, 오류 없음) | AST calls |
| `fmt.Sprintf` `:536` | detail 조립 | 순수 | AST calls |
| `json.Marshal` `:539` | payload | 실패는 B2가 삼킨다 | AST calls |
| **`g.journal.EnqueueAlert` `:551`** | **outbox 기록** | **오류를 버린다** | AST calls |
| `string` `:557` | payload 변환 | 순수 | AST calls |

**`EnqueueAlert`이지 `ClaimAlertForDelivery`가 아니다.** 즉 이 경로는
**기록만 하고 배달을 시작하지 않는다.** `Notifier`를 거치지 않으므로 발행도,
시도 기록도, 재시도도 없다.

## State mutations and fallbacks

| 대상 | 자리 | 성격 |
|---|---|---|
| `EntryGate` 래치 | B1 `:536` | 인메모리, 재시작으로 소멸 |
| `alert_outbox` 행 | `:551` | 내구, `state = PENDING` |

- fallback: B2 하나. payload를 못 만들면 **빈 payload로 계속 간다** —
  알림 자체를 포기하지 않는다.

## Safety conclusion

- **Safe edit boundary**: a092는 편집하지 않는다.
- **High-risk impact**: yes — 미해소 주문(`UNRESOLVED_IN_DOUBT`)의 통지 경로다.
- **결함(측정된 것).** `:551`이 만든 `order.unresolved_in_doubt` 행을
  프로덕션에서 집어가는 코드가 없다. `PendingAlerts`를 부르는 프로덕션 경로는
  `Notifier.Flush`(`notifier.go:437`)와 `Notifier.Acknowledge`(`:491`)뿐이고
  **둘 다 프로덕션 호출자가 0이다**(`internal/`·`cmd/` 전체 `rg` 확인).
  그러므로 이 알림은 **DB에 쌓이기만 하고 사람에게 도달하지 않는다.**
- **`:530-533`의 doc comment가 그 반대를 단언한다** — *"losing the notification
  is bad but not unsafe"*는 배달이 다른 곳에서 일어난다는 전제 위에 있고,
  그 전제가 오늘 거짓이다.
- 17판이 배달 루프를 세우면 이 행이 실제로 나간다. **§6.0 R17-7**이 그 회귀
  테스트다: 이 함수가 넣은 행이 배달 루프를 통해 발행되는지 관측한다.
