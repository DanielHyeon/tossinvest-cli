# Function Logic Map: `Notifier.publishBestEffort`

- Source: `internal/obs/notifier.go` (138-150)
- AST evidence: `ast.json` — branches 2, returns 1, calls 6, assignments 1,
  **defers 0, go_statements 0**
- Risk scan: `risk-pattern-report.md`

normal 등급의 전 경로. **"최선노력"은 발송 결과에 대한 말이고 소요 시간에 대한 말이
아니다** — 이 함수도 동기다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `n.Publisher` | nil 허용 | 조립 | B1 `:139`이 **로그 없이** 반환 — 조용한 구멍 |
| `severity` | normal 또는 critical | `Notify:112` normal / `notifyCritical:163` critical(강등) | 메시지 우선순위에만 쓰임 |
| `n.Log` | nil 허용 | 조립 | B2가 nil을 흡수 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return | 소요 |
|---|---|---|---|---|
| B1 `:139` | `n.Publisher == nil` | **없음 — 로그도 없다** | `return` `:140` | 0 |
| B2 `:142` | `Publish` 오류 **and** `n.Log != nil` | `Log.Warn(EventAlertUndelivered, event/severity/error)` `:145-148` | 암묵 `:150` | — |
| — | 성공 | 없음 | 암묵 `:150` | **최대 10s** |

**outbox 행을 만들지 않는다.** 재시도도 게이트도 없다. 주석 `:143-144`가 그것을 계약으로
쓴다: "treating its failure as an incident would make the grading meaningless".

## Calls and live bindings

| Callee | Why called | **Error/timeout/retry contract** | Evidence |
|---|---|---|---|
| `notificationFor` `:142` | 메시지 조립 | 순수 | AST calls |
| **`n.Publisher.Publish` `:142`** | **원격 발송** | **1회, 재시도 없음, 상한 10s**(`ntfy.go:96` `timeout <= 0 → 10s`) | `internal-obs--ntfy.publish/ast.json` |
| `n.Log.Warn` `:145` | 실패 기록 | 네트워크 없음 | AST calls |

**normal 알림의 동기 체류 상한은 10초다.** critical의 34초보다 작지만
exit 관측 주기(**5초**)보다는 크다.

### 이 함수의 체류는 **측정된 적이 없다** — 아래 숫자는 체류가 아니다

`engine.log`에서 연속 `exit.position_unmanaged` 줄의 간격을 재면 이런 값이 나온다:
0.202 / 0.208 / 0.623 / 0.644 / 0.656 / 0.660 / 0.688 / 1.811 / 1.836 초 (표본 9),
그리고 두 건이 연달아 나간 사이클 하나는 2.499초(`14:47:41.509 → 14:47:44.008`).

**이 값들은 `publishBestEffort`의 동기 체류가 아니다.**
`analysis/delivery-latency.md` §5.3이 그 이유를 적는다 — `exit.position_unmanaged`는
**조립 자리가 4곳**이고 그중 셋이 대사 루프 goroutine이다(`runtime.go:277-283`).
줄 간격은 **어느 루프의 체류도 재지 못한다.** 예: 표본의 1.836초를 만드는 뒤쪽 줄
(`engine.log` 7814→7815)은 `adopted_quantity`를 가진 `adoption.go:455` 생산이고
바로 다음이 `reconcile.clean`이다 — **대사 goroutine의 줄**이다.

그래서 이 산출물은 이 숫자를 **예산 유도의 입력으로 쓰지 않는다.** 유도가 쓰는 유일한
표본은 §3의 6건(0.1983 ~ 0.7540초)이고, 그것은 짝지은 줄이 모두 인접함을 확인한
critical 경로의 값이다. 여기 남겨 두는 이유는 **지웠다가 다음 판에서 같은 방법으로
다시 만들어지는 것을 막기 위해서**다: 이 방법은 무효이고, 무효인 이유가 여기 있다.

**normal 등급의 실제 체류를 재려면 계측이 필요하고 그것이 a090이다.**

## State mutations and fallbacks

- 상태 변경 없음(assignments 1 = `err :=`).
- fallback 없음. B1은 **fallback도 로그도 없이** 반환한다 — nil publisher에서 normal은
  흔적이 0이다(a091 proposal이 인용한 성질).
- **goroutine 없음**.

## Safety conclusion

- **Safe edit boundary**: 이 함수의 계약은 "결과를 사건으로 만들지 않는다"이고
  그것은 유지된다. 바꿀 수 있는 것은 **누가 언제 부르는가**다.
- **High-risk impact**: **yes**(간접) — exit 관측 루프의 normal 알림 2종
  (`EventExitProposalCapped` `exitloop.go:1431`, `EventExitPositionUnmanaged` `:1501`)이
  이 경로로 최대 10초 동기 체류한다.
- **비동기화가 여기서 더 쉬운 이유**: outbox가 없으므로 durability 계약이 없다.
  "최선노력"은 이미 **유실 허용**을 뜻하므로 유계 큐에서 넘치면 버려도 계약 위반이 아니다.
  critical과 정확히 반대 성질이고, 그래서 두 등급의 비동기화 방식이 달라야 한다.
