# Function Logic Map: `reportEngineEndpointDegraded`

- Source: `cmd/tossctl/engine.go` (468-505)
- AST evidence: `ast.json` — AST 분기 1 · return 1 · go 문 1 · defer 0
  (source_sha256 `3d941a4616837f73ce2a2acf87ed4bfada5eaccdf4b77fa966b46ac4cd972209`,
  a109 §2.2 편집 후 생성)
- Risk scan: `risk-pattern-report.md`
- **유래**: a108 의 `reportStrategyProjectionDegraded` 를 endpoint 좌표를 받는 형태로
  일반화한 것이다(D3a). 이름이 바뀐 이유는 하나다 — 일반화한 함수가 「strategy
  projection」을 이름에 달고 있으면 그 이름이 거짓말을 한다. **분기 구조와 금지 3종은
  그대로**이고, 기준(base) 판의 맵은
  `cmd-tossctl--reportstrategyprojectiondegraded/` 에 base revision 으로 남아 있다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `ctx` | 부팅 ctx (SIGTERM 에서 끊긴다) | `runEngineRun` | **그대로 쓰면 안 된다** — `context.WithoutCancel` 로 떼어 낸다. 상속하면 「종료하면서 남기는 마지막 말」이 늘 유실된다 |
| `ectx` / `ectx.Notifier` | nil 가능 | 조립 결과 | nil 이면 stderr 만 남기고 즉시 return (B1) |
| `errOut` | 기동 stderr | cobra | 동기 한 줄 — **이것이 1차 표면이다** |
| `endpoint` | 네 좌표 중 하나 | `engine*Endpoint(dir)` 생성자 | 표면 이름·잃는 것·control 경로·이벤트·제목을 싣는다 |
| `cause` | 기동 실패 원인 | 해당 Start | 문구에 그대로 싣는다 — **운영자가 제거해야 할 대상**이다 |
| 이벤트 등급 | **Normal 이어야 한다** | 두 이벤트 타입 모두 obs 등급표에 없다 | 등급표 미등재 → `obs.SeverityOf` 가 Normal → 원장에 닿지 않는다 |

**금지 3종 (불변 — engine.go:358–397 주석이 정본)**:
1. 이 이벤트 타입들을 `internal/obs` 의 `criticalEvents` 등급표에 올리지 마라.
2. severity 를 critical 로 올리지 마라.
3. 이 보고를 원장 outbox(`Journal.EnqueueAlert`)에 싣지 마라.

미전달 PENDING 행은 다음 부팅의 진입 게이트를 잠그고(`UndeliveredCount` →
`restoreAlertEntryLatch` → `ReasonAlertUndelivered`), 해제는 운영자 ack 뿐이다.

**비차단 불변식**: `Notify` 는 Publisher 가 있으면 그 자리에서 발행하고 상한이
`obs.DefaultPublishTimeout`(10s)다. 이 함수는 `rt.Run` **앞**에서 불리므로 동기 호출은
**손절 루프가 시작되지 않는 10초**가 된다. 그래서 stderr 만 동기, 발행은 goroutine.

**a109 가 더한 불변식**: 보고는 **어느 표면인지** 말한다. 강등이 넷으로 늘면 구별할 수
없는 보고는 운영자에게 재시작밖에 안내하지 못한다. 그리고 안내는 「재시작하라」가 아니라
**「원인을 제거한 뒤 재시작하라」**다 — 잔여 원인은 이물·환경 이상이고 결정적이다(D3b).

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (:475) | `ectx == nil \|\| ectx.Notifier == nil` | 없음 (stderr 는 이미 찍혔다) | 조기 return — obs 이벤트 없음 | 직접 핀 없음; non-nil 경로는 `TestAFailedSiblingEndpointDoesNotStopTheEngine` 이 지난다 |
| — (:499 go) | B1 을 지났다 | **떼어 낸 ctx** 로 goroutine 에서 `Notify` | 반환값 버림 — critical 전용 오류만 돌아온다 | `TestTheDegradedBootDoesNotWaitForTheNotifier` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `fmt.Sprintf` / `fmt.Fprintf` | 동기 stderr 세 줄 | 실패 없음 | AST · engine.go:469–474 |
| `context.WithoutCancel` | 부모 취소로부터 분리 | 값은 상속, 취소는 안 한다 | AST · :498 |
| `ectx.Notifier.Notify` | obs Normal 이벤트 | goroutine 안 — 반환 오류는 critical 전용 | AST · :500 |

호출자: `runEngineRun` 의 B16·B18·B20·B23 (engine.go:297·305·330·344) — 네 endpoint 가
같은 의례를 쓴다.

## State mutations and fallbacks

- **원장·디스크 상태 변경 없음.** 그래서 `ectx.Close()` 와 겹쳐도 안전하다.
- 부작용은 둘: 동기 stderr 세 줄, 비동기 obs 이벤트 하나.
- fallback: Notifier 가 없으면 stderr 가 전부다(B1) — **의도된 최소 표면**이다.

## Safety conclusion

- Safe edit boundary: 문구와 endpoint 좌표. `WithoutCancel` + goroutine 구조, B1 가드,
  이벤트 등급(Normal)은 바꾸지 않는다.
- High-risk impact: **yes** — 엔진 기동 경로에서 불리고, 잘못 바꾸면(등급 상승·outbox
  적재) **실계좌 신규 진입이 영구 차단된다**.
- 금지: 위 금지 3종. 동기 발행으로 되돌리는 것. 어느 표면인지 말하지 않는 보고.
