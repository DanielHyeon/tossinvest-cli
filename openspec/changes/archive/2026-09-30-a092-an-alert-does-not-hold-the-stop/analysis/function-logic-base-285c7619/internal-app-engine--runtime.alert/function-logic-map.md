# Function Logic Map: `Runtime.alert`

- Source: `internal/app/engine/runtime.go` (444-456)
- AST evidence: `ast.json` — branches 2, returns 1, calls 7, assignments 2,
  **defers 1, go_statements 0**
- Risk scan: `risk-pattern-report.md`

**이 함수는 선례다.** `ExitObserver.alert`·`ReconcileDriver.alert`와 하는 일이 같은데
혼자만 기한을 씌운다. 이 change의 논거 절반이 그 차이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `ctx` | 종료 중이면 **이미 취소된 상태일 수 있다** | 감독 루프 | `context.WithoutCancel`로 취소를 **떼어낸다**(`:451`) — 주석 `:448-450`: "sending them through the cancelled context would drop the one alert that explains the shutdown" |
| `e` | `obs.Event` | 감독 계약 2층(방어적 종료·지속 열화) | — |
| `r.opts.Alerts` | nil 허용 (`runtime.go:141`) | 조립 | nil이면 B1이 조용히 반환 |
| `alertDeliveryBound` | **30s** 상수 (`runtime.go:461`) | 같은 파일 | 초과 시 ctx 만료 → `Notify`가 오류로 돌아온다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return |
|---|---|---|---|
| B1 `:445` | `r.opts.Alerts == nil` | 없음 | `return` `:446` |
| B2 `:453` | `Notify`가 오류 | `r.log(EventAlertUndelivered, warn=true, …)` `:454` | 암묵 `:456` |

## Calls and live bindings

| Callee | Why called | **Error/timeout/retry contract** | Evidence |
|---|---|---|---|
| `context.WithoutCancel` `:451` | 종료 중인 ctx의 취소를 떼어냄 | 오류 없음 | AST calls |
| `context.WithTimeout` `:451` | **기한 30초 부여** | 초과 시 하위 publish가 `ctx` 만료로 실패 | AST calls + `runtime.go:461` |
| `cancel` `:452` (**defer**) | 누수 방지 | — | AST defers 1 |
| `r.opts.Alerts.Notify` `:453` | 통지 | **동기.** 상한은 30초로 **닫혀 있다** — `ExitObserver.alert`의 무기한과 여기가 다르다 | AST calls |
| `r.log` `:454` | 실패 기록 | 네트워크 없음 | AST calls |

### 30초는 어디서 왔는가

`runtime.go:458-461`이 직접 말한다:

```go
// alertDeliveryBound is how long the runtime waits for an alert it raises while
// stopping. Generous enough for obs.Notifier's three bounded publish attempts,
// finite so a dead transport cannot hold the shutdown open.
const alertDeliveryBound = 30 * time.Second
```

**"three bounded publish attempts"** — 저장소는 `deliver`의 예산을 이미 알고 있었고
그것을 근거로 상수를 골랐다. 다만 30초는 실제 최악(3×10 + 2×2 = **34초**)보다 **4초
짧다**: 마지막 publish가 10초를 다 쓰면 ctx가 먼저 만료되어 `Publish`가 오류로 끊긴다.
그것이 이 상수의 의도이므로 결함이 아니라 **의도된 절단**이다.

## State mutations and fallbacks

- 상태 변경 없음. assignments 2는 `alertCtx, cancel :=`(`:451`)와 `err :=`(`:453`)뿐.
- fallback: `Notify` 실패는 `EventAlertUndelivered` warn 한 줄. 종료는 계속된다.
- **goroutine 없음**(`go_statements: 0`) — 발송은 여전히 동기다. 이 함수가 고른 것은
  "비동기"가 아니라 "**유계**"다.

## Safety conclusion

- **Safe edit boundary**: 이 change는 이 함수를 **바꾸지 않는다.** 근거로만 쓴다.
- **High-risk impact**: no — 종료 경로이고 주문을 내지 않는다.
- **이 함수가 증명하는 것**: "죽은 transport가 루프를 붙잡으면 안 된다"는 판단은 이미
  이 저장소의 것이다. 새 원칙이 아니다. 적용되지 않은 곳이 **손절이 사는 루프**라는
  것이 이 change가 여는 문제다.
