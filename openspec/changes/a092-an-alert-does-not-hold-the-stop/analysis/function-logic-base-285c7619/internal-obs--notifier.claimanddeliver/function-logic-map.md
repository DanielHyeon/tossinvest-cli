# Function Logic Map: `Notifier.claimAndDeliver`

- Source: `internal/obs/notifier.go` (238-277)
- AST evidence: `ast.json` — branches 4, returns 3, calls 9, assignments 1,
  **defers 1, go_statements 0**
- Risk scan: `risk-pattern-report.md`

**a092가 존재하는 이유의 자리.** 관측 사이클이 알림 하나를 올릴 때 지나는 경로이고,
`:241`의 `n.mu.Lock()`부터 `:276`의 `n.deliver(...)`까지가 **한 잠금 안**이다.
17판은 그 구간에서 `deliver`를 들어낸다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `record` | `journal.Alert` | 호출자 `notifyCritical` | 키·타입 검증은 journal이 한다 |
| `e` | `Event` | 같은 위 | 로그 필드로만 쓰인다 |
| `n.remindAfter()` | `> 0` (기본 1시간) | `:280-285` | 0 이하를 기본값으로 올린다 |
| `n.Log`·`n.Gate` | nil일 수 있다 | 배선 | B2 `:258`, B3 `:261` |
| `n.mu` | 배달 뮤텍스 | `:241`, defer `:242` | **claim과 send를 한 원자 단위로 묶는다** |

**`:230-234`의 주석이 뮤텍스의 근거를 적는다**: 잠금 밖에서 claim하면 같은 조건의
두 관측이 아직 배달 안 된 같은 행을 읽고 둘 다 발행한다(a096 라운드 1, blocker 1).

**17판은 그 근거를 옮긴다.** 발행이 이 함수 밖으로 나가면 "claim과 send를 묶는다"는
목적이 사라지고, 남는 배제는 claim 사이의 배제뿐이다 — 그리고 그것은
`claimOwed`의 `PENDING → rearm=false`와 `MarkAlert*`의 CAS가 이미 제공한다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return |
|---|---|---|---|
| **B1 `:245`** | **claim 트랜잭션 실패** | 로그 `:259` + **게이트 래치 `:262`** | `:264` `(false, false, err)` |
| B2 `:258` | `n.Log != nil` | 오류 로그 한 줄 | — |
| B3 `:261` | `n.Gate != nil` | `Gate.Block(ReasonAlertUndelivered, detail)` | — |
| **B4 `:266`** | **`!owed` — 이미 알렸고 창이 안 지났다** | **없음** | `:274` `(false, false, nil)` |
| — `:276` | — | **`n.deliver(ctx, id, e)`** | `(sent, true, nil)` |

**`:276`이 17판이 잘라내는 한 줄이다.** `deliver`는 재시도 루프와 시도 간 대기를
가진 함수이므로(`:341-407`), **이 한 줄이 관측 사이클에 원격 왕복 × 시도 수를
집어넣는다.**

**B1은 a097이 만든 분기다.** 원장 기록 자체가 실패하면 행이 없으므로 나중에
집어갈 것도 없다 — 그래서 **그 자리에서 동기로** 래치한다. `:255`가 명시한다:
*"Entries only. Exits are untouched: no alert failure may slow a stop."*
**17판도 이 분기는 동기로 남긴다** — 미룰 대상인 행이 존재하지 않기 때문이고,
exit-policy 델타의 "durable 기록이 실패한다" 시나리오가 그것을 적었다.

**B4가 재알림 억제다.** 억제해도 로그 줄은 이미 쓰였다(`:271-273`) — 조건이
얼마나 지속됐는지의 기록이 발송 여부에 의존하지 않는다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `n.mu.Lock` `:241` | claim+send 원자화 | **17판이 축소한다** | AST calls |
| `n.mu.Unlock` `:242` | **defer** | — | AST defers **1** |
| `n.Journal.ClaimAlertForDelivery` `:244` | id + owed 판정 | 로컬 SQLite | AST calls |
| `n.remindAfter` `:244` | 재알림 창 | 순수 | AST calls |
| `fmt.Sprintf` `:256` | 래치 detail | 순수 | AST calls |
| `n.Log.Error` `:259` | 구조화 로그 | — | AST calls |
| `n.Gate.Block` `:262` | **진입 래치** | 실패 불가(void) | AST calls |
| **`n.deliver` `:276`** | **원격 전송 + 재시도 + 대기 + 래치 + 승격** | **여기서 사이클이 멈춘다** | AST calls |

## State mutations and fallbacks

| 대상 | 자리 | 성격 |
|---|---|---|
| `alert_outbox` 행 | `:244` 안 | 생성 또는 재무장 |
| 진입 게이트 래치 | B3 `:262` | 인메모리, claim 실패 시 |
| 나머지 전부 | `:276` `deliver` 안 | **17판이 옮기는 것** |

- fallback: B1이 유일하다 — 오류를 돌려주면서 **동시에** 래치한다.
  `:250-253`이 근거를 적는다: 호출자가 오류를 검사하는지에 결과가 의존하면 안 되고,
  `internal/flatten/flatten.go:694`가 실제로 버린다.

## Safety conclusion

- **Safe edit boundary**: **a092가 편집한다**(§8 GREEN). `:276`이 동기 호출에서
  "outbox에 남기고 즉시 반환"으로 바뀐다.
- **High-risk impact**: yes — exit 관측 사이클이 이 함수를 통해 알림을 올린다.
  §0.3(손절 즉시성)에 가장 가까운 자리다.
- **바뀌지 않아야 하는 것 둘**:
  1. **B1의 동기 래치.** 행이 없으면 미룰 대상이 없다.
  2. **B4의 억제 판정.** 등급이나 경로로 좁히면 a096의 계약이 깨진다.
- **바뀌는 것 하나**: `:276`. 그 한 줄이 사이클에 넣던 시간이
  17판이 없애는 전부다.
- **미해결**: 잠금을 좁힌 뒤 `TestFlushCannotPublishBesideASend`가 지키던 성질을
  무엇이 지키는가. 답은 CAS이고, **그 답이 참인지는 R17-3이 관측한다.**
