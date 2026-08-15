# Function Logic Map: `reportStrategyProjectionDegraded`

- Source: `cmd/tossctl/engine.go` (369-405)
- AST evidence: `ast.json` — AST 분기 1 · return 1 · go 문 1 · defer 0
  (source_sha256 `8111c1c9e20f501b6221e231836fb02d7d03d127b3592892175c1beb38788381`,
  a109 base `016da624`)
- Risk scan: `risk-pattern-report.md`
- **a109 T2 편집 대상: 시그니처의 일반화** — 오늘 이 함수는 strategy projection **하나**
  에 묶여 있다(`strategyprojectionrpc.ControlDirectory(dir)` 를 스스로 계산하고 문구가
  「전략 화면」을 단정한다). a109 D3a 는 endpoint 이름을 받는 형태로 일반화해 네
  endpoint 가 같은 의례를 쓰게 한다. **분기 구조와 금지 3종은 그대로다.**

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `ctx` | 부팅 ctx (SIGTERM 에서 끊긴다) | `runEngineRun` | **그대로 쓰면 안 된다** — `context.WithoutCancel` 로 떼어 낸다(:396). 상속하면 「종료하면서 남기는 마지막 말」이 늘 유실된다 |
| `ectx` | nil 가능 | `runEngineRun` | nil 이면 stderr 만 남기고 즉시 return (B1) |
| `ectx.Notifier` | nil 가능 | 조립 결과 | 같음 (B1) |
| `errOut` | 기동 stderr | cobra | 동기 한 줄 — **이것이 1차 표면이다** |
| `dir` | 엔진 디렉터리 | `runEngineRun` | control 디렉터리 경로 계산에만 쓴다 |
| `cause` | 기동 실패 원인 | 해당 Start | 문구에 그대로 싣는다 — 원인이 결정적이므로 **운영자가 제거해야 할 대상**이다(design D3 ③) |
| 이벤트 등급 | **Normal 이어야 한다** | `engineStrategyProjectionDegradedEvent` 주석 (:330–359) | 등급표 미등재 → `obs.SeverityOf` 가 Normal → `Notify` 는 로그+best-effort 발행만 하고 **원장에 닿지 않는다** |

**금지 3종 (불변 — 주석 :330–359 가 정본)**:
1. 이 이벤트 타입을 `internal/obs` 의 `criticalEvents` 등급표에 올리지 마라.
2. severity 를 critical 로 올리지 마라.
3. 이 보고를 원장 outbox(`Journal.EnqueueAlert`)에 싣지 마라.

이유는 하나다: outbox 의 **미전달 PENDING 행은 다음 부팅의 진입 게이트를 잠근다**
(`Journal.UndeliveredCount` 는 Type 무필터, `restoreAlertEntryLatch` 가 >0 이면
`execgw.ReasonAlertUndelivered` 로 latch, 해제는 운영자 ack 뿐). publisher 없는
배포에서 그 행은 영원히 PENDING 이므로 결과는 「표면 하나를 잃었다는 보고가 실계좌의
신규 진입을 영구 차단한다」가 된다.

**비차단 불변식**: `Notify` 는 Publisher 가 붙어 있으면 그 자리에서 발행하고 한 번의
상한이 `obs.DefaultPublishTimeout`(10s)다. 이 함수는 `rt.Run` **앞**에서 불리므로 동기
호출은 **손절 루프가 시작되지 않는 10초**가 된다. 그래서 stderr 만 동기이고 발행은
goroutine 이다(:397).

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (:375) | `ectx == nil \|\| ectx.Notifier == nil` | 없음 (stderr 는 이미 찍혔다) | 조기 return (:376) — obs 이벤트 없음 | 미고정 직접 핀 없음; `TestAFailedStrategyProjectionDoesNotStopTheEngine` 이 non-nil 쪽 경로를 지난다 |
| — (:397 go) | B1 을 지났다 | **떼어 낸 ctx** 로 goroutine 에서 `Notify` | 반환값 버림 — `Notify` 는 critical 을 durable 하게 못 만들었을 때만 오류를 준다 | `TestTheDegradedBootDoesNotWaitForTheNotifier` (① 기한 내 부팅 ② Notifier 에 닿음 ③ 종료가 끊지 않음) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strategyprojectionrpc.ControlDirectory` | 보고 문구의 scope 값 | 순수 함수 | AST · engine.go:371 |
| `fmt.Sprintf` / `fmt.Fprintf` | 동기 stderr 한 줄 | 실패 없음 | AST · engine.go:372–374 |
| `context.WithoutCancel` | 부모 취소로부터 분리 | 값은 상속, 취소는 안 한다 | AST · engine.go:396 |
| `ectx.Notifier.Notify` | obs Normal 이벤트 | **goroutine 안** — 반환 오류는 critical 전용이라 여기서 발생하지 않는다(`obs.Notifier.notifyCritical`) | AST · engine.go:398 |

호출자: `runEngineRun` B18 (engine.go:304). a109 이후에는 B15·B16·B20 도 같은 함수를
부른다 — 그것이 이 함수를 일반화하는 이유다.

## State mutations and fallbacks

- **원장·디스크 상태 변경 없음.** 그래서 `ectx.Close()` 와 겹쳐도 안전하다: Normal
  경로는 로그 한 줄과 best-effort 발행뿐이라 journal handle 에 닿지 않는다.
- 부작용은 둘: 동기 stderr 한 줄, 비동기 obs 이벤트 하나.
- fallback: Notifier 가 없으면 stderr 한 줄이 전부다(B1). 그것이 **의도된 최소 표면**
  이다 — 보고 수단이 없다고 기동을 실패시키지 않는다.
- a109 가 더하는 것은 「어느 endpoint 인가」를 문구와 scope 필드가 말하게 하는 것뿐이다.
  강등이 세 개로 늘면 **어느 표면이 없는지 구별할 수 없는 보고**는 운영자에게 쓸모가 없다.

## Safety conclusion

- Safe edit boundary: **시그니처(endpoint 이름·control 경로를 인자로)와 문구**.
  `WithoutCancel` + goroutine 구조, B1 가드, 이벤트 등급(Normal)은 바꾸지 않는다.
- High-risk impact: **yes** — 엔진 기동 경로에서 불리고, 잘못 바꾸면(등급 상승·outbox
  적재) **실계좌 신규 진입이 영구 차단된다**. 코드가 아니라 등급표 한 줄이 사고를
  만드는 자리다.
- 금지: 위 금지 3종. 그리고 이 함수를 동기 발행으로 되돌리는 것 — 보호의 시작이 늦는다.
