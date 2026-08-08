# Function Logic Map: `ReconcileDriver.alert`

- Source: `internal/app/engine/reconcileloop.go` (552-560)
- AST evidence: `ast.json` — branches 2, returns 1, calls 3, assignments 1,
  **defers 0, go_statements 0**
- Risk scan: `risk-pattern-report.md`

대사 루프의 알림 자리. `ExitObserver.alert`와 **구조가 같고 기한도 없다.**

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `ctx` | 대사 루프의 컨텍스트 | `ReconcileDriver`의 주기 루프 | **그대로 전달**(AST calls에 `context.*` 없음) |
| `e` | `obs.Event` | 대사 사이클 | — |
| `d.opts.Alerts` | nil 허용 (`reconcileloop.go:139`) | `reconcileloop.go:367-368`이 `c.Notifier` 주입 | nil이면 B1이 반환 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return |
|---|---|---|---|
| B1 `:553` | `d.opts.Alerts == nil` | 없음 | `return` `:554` |
| B2 `:556` | `Notify` 오류 **and** `d.opts.Log != nil` | `Log.Error(EventAlertUndelivered, …)` `:557-558` | 암묵 `:560` |

## Calls and live bindings

| Callee | Why called | **Error/timeout/retry contract** | Evidence |
|---|---|---|---|
| `d.opts.Alerts.Notify` `:556` | 통지 | **동기·무기한** — `ExitObserver.alert`와 동일. normal 10s / critical 34s | AST calls |
| `d.opts.Log.Error` `:557` | 실패 기록 | 네트워크 없음 | AST calls |
| `string` `:558` | 이벤트 종류 문자열화 | — | AST calls |

## State mutations and fallbacks

- 상태 변경 없음(assignments 1 = `err :=`).
- **goroutine 없음**, **defer 없음**.

## Safety conclusion

- **Safe edit boundary**: 대사 루프의 주기는 **60초**(실측 `reconcile.clean` 간격
  62.6~63.5s, `engine.log`)이므로 34초 체류는 exit 루프(5초 주기)만큼 치명적이지 않다.
  그러나 같은 `*obs.Notifier`를 공유하고 `deliver`가 `n.mu`를 잡으므로
  (`internal-obs--notifier.deliver/ast.json` defers 1 = `n.mu.Unlock`),
  **대사 루프의 critical 알림이 exit 루프의 critical 알림을 그 시간만큼 막는다.**
- **High-risk impact**: **yes**(간접) — 자기 루프가 아니라 **exit 루프를 막는 경로**로서
  High-risk다. 이 경로가 exit 루프와 뮤텍스를 공유한다는 사실이 이 change가 두 호출자를
  같이 다뤄야 하는 이유다.
- 이 change의 범위 결정: `ReconcileDriver.alert` **본문은 바꾸지 않는다.** 공유 자원
  (`Notifier`)의 성질을 바꾸면 이 호출자도 자동으로 혜택을 받는다.
