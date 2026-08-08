# Function Logic Map: `Notifier.Notify`

- Source: `internal/obs/notifier.go` (107-116)
- AST evidence: `ast.json` — branches 1, returns 2, calls 4, assignments 1,
  **defers 0, go_statements 0**
- Risk scan: `risk-pattern-report.md`

세 호출자(`ExitObserver.alert`·`ReconcileDriver.alert`·`Runtime.alert`)와
`notifierAlerter` 두 메서드가 전부 여기로 들어온다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `ctx` | 호출자의 컨텍스트 | 호출자 5곳 | 그대로 하위로 전달 |
| `e.Type` | `obs.EventType` | `event.go` | 미등록 종류는 `SeverityNormal`(`SeverityOf` 기본값) |
| `n.Log` | nil 허용 | 조립 | `logEvent`가 nil을 흡수 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return | 소요 상한 |
|---|---|---|---|---|
| B1 `:111` | `severity != SeverityCritical` | `n.publishBestEffort(ctx, e, severity)` `:112` | `return nil` `:113` | **10s** |
| — | critical | `n.notifyCritical(ctx, e)` `:115` | `return` `:115` | **34s** |

**이탈 둘 다 반환값은 nil이거나 outbox 쓰기 오류뿐이다.** 전달 실패는 이 함수의 오류가
아니다 — 주석 `:103-106`이 그것을 계약으로 못박는다: "A failed send is not an error to
the caller: it has already been handled here, by latching the gate."

## Calls and live bindings

| Callee | Why called | **Error/timeout/retry contract** | Evidence |
|---|---|---|---|
| `SeverityOf` `:108` | 등급 판정 | 순수 함수, map 조회 1회 (branches 1) | AST calls + `event.go:309-314` |
| `n.logEvent` `:109` | 구조화 로그 | 네트워크 없음 | AST calls |
| `n.publishBestEffort` `:112` | normal 발송 | publish **1회**, 상한 **10s** | `internal-obs--notifier.publishbesteffort/ast.json` |
| `n.notifyCritical` `:115` | critical 경로 | outbox 쓰기 + publish **최대 3회**, 상한 **34s**, `n.mu` 보유 | `internal-obs--notifier.notifycritical/ast.json` |

**두 경로 모두 동기다.** AST `go_statements: 0` — 이 함수는 어떤 것도 백그라운드로
넘기지 않는다.

## State mutations and fallbacks

- 이 함수 자체는 상태를 바꾸지 않는다(assignments 1 = `severity :=`).
- 로그는 등급과 무관하게 **먼저** 나간다(`:109`가 분기 `:111`보다 위). 그래서
  `engine.log`의 이벤트 줄 시각은 **발송 시작 시각**이고, 이것이 위 실측의 기준점이다.

## Safety conclusion

- **Safe edit boundary**: 이 함수의 계약(전달 실패는 호출자 오류가 아니다)은 유지된다.
  바뀔 수 있는 것은 **언제 반환하는가**다.
- **High-risk impact**: **yes** — exit 관측 루프가 이 함수 안에 머무는 동안 손절 판정이
  진행되지 않는다.
- 편집 시 지켜야 할 것: **durable 기록은 반환 전에 끝나야 한다.** `notifyCritical`
  주석 `:175-176`이 이유를 쓴다 — "a record that only exists in memory is a record that
  does not survive the crash it is warning about". 비동기화는 **발송**에만 적용 가능하고
  **기록**에는 적용 불가다.
