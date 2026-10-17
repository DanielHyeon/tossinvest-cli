# Function Logic Map: `ExitObserver.checkOutage`

- Source: `internal/app/engine/exitloop.go` (767-804)
- AST evidence: `ast.json` — branches 5, returns 4, calls 19, assignments 5,
  **defers 0, go_statements 0**
- Risk scan: `risk-pattern-report.md`

**H1의 자리다.** 한 사이클에서 `obs.Notifier.Notify`에 **두 번** 도달하는 유일한 함수이고,
두 번째는 `o.alert`을 지나지 않는다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `o.lastObserved` | 마지막 성공 관측 시각, zero 허용 | `ObserveOnce:438`·`:449`가 stamp | zero면 `o.startedAt`으로 대체(B1 `:769`) |
| `o.outageAfter()` | > 0 | `:333-338` — 0 이하면 `DefaultExitObservationOutage` **60s**(`:104`) | 없음 |
| `o.outageRaised` | bool 래치 | 이 함수(`:778`)와 `ObserveOnce:439`·`:450` | 이미 true면 B3이 반환 |
| `o.opts.Escalate` | nil 허용 | `ExitObserverOptions` | nil이면 B4가 반환 |
| `o.opts.AccountRef` | 비어 있음 허용 | 같은 위 | 공백이면 B4가 반환 |
| `o.opts.Announcer` | nil 허용 | 프로덕션은 `ectx.Notifier`(`cmd/tossctl/engine.go:351`) | nil이면 journal이 알리지 않는다 |
| `cycle` | 포인터 | 호출자 `ObserveOnce` | `:803`만 씀 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return | `Notify` 도달 |
|---|---|---|---|---|
| B1 `:769` | `since.IsZero()` | `since = o.startedAt` `:770` | — | — |
| B2 `:772` | 경과 < `outageAfter()` (**60s**) | 없음 | `return` `:773` | ❌ |
| B3 `:775` | `o.outageRaised` | 없음 | `return` `:776` | ❌ |
| — `:778` | — | `o.outageRaised = true` | — | — |
| — `:780` | — | **`o.alert(...)` — `Notify` 도달 #1 (P4)** | — | ✅ **≤34s** |
| B4 `:793` | `Escalate == nil \|\| AccountRef 공백` | 없음 | `return` `:794` | ❌ |
| — `:796` | — | **`EscalateOperatingMode(..., o.opts.Announcer)` — `Notify` 도달 #2 (P1a)** | — | ✅ **≤34s** |
| B5 `:798` | 승격이 오류 | `o.logErr(EventOperatingMode, ...)` `:799` | `return` `:801` | — |
| — `:803` | — | `cycle.Escalated = changed` | 암묵 `:804` | — |

**`:780`과 `:796` 사이의 이탈은 B4 하나뿐이다.** 프로덕션은 `Escalate`와 `AccountRef`를
둘 다 채우므로 **B4는 프로덕션에서 통과한다** — 다만 **두 값은 같은 자리에서 오지 않는다**:
`Escalate: ectx.Journal`은 `cmd/tossctl/engine.go:350`이고 `AccountRef`는 그 파일에 없으며
`exitwiring.go:336`(`opts.AccountRef = c.AccountRef`)이 채운다. 7판까지 이 문장은 둘 다
`engine.go:351`이라고 적었는데 **`:351`은 `Announcer`다**(7라운드 H6) —
즉 두 번의 `Notify`가 **연달아** 일어난다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `o.clk.Now` `:772`·`:783`·`:789` (3회 — `ast.json` 열거) | 경과 계산 | 즉시 | AST calls |
| `o.outageAfter` `:772` | 임계 조회 | 즉시 | AST calls |
| `o.alert` `:780` | 두절 통지 | **동기·기한 없음.** normal 10s / critical **34s** — 이 이벤트는 critical(`event.go` `criticalEvents`) | `internal-app-engine--exitobserver.alert` FLM |
| `o.opts.Escalate.EscalateOperatingMode` `:796` | 운영 모드 승격 | **동기.** journal 트랜잭션 + 커밋 **뒤** `Announcer.AnnounceOperatingMode`(`operating_mode.go:466-479`) → `obs/mode.go:57` `Notify` → critical **34s**. **알림은 커밋 밖이므로 34초가 트랜잭션 안에 갇히지는 않는다** | `analysis/notify-reach.md` |
| `o.logErr` `:799` | 실패 기록 | 즉시 | AST calls |

### 두절 사이클의 최악 동기 체류 = 2 × 34s = **68s — 단 NORMAL 계정에서만**

**오늘 `:796`은 알림을 만들지 않는다.** `TransitionOperatingMode`는 계정이 이미 목표
모드면 `direction == 0`에서 **announce 전에** `changed=false`로 반환한다
(`operating_mode.go:409-415`). `ModeTriggerExitObservationOutage`의 목표는
`ENTRY_BLOCKED`이고(`:537-549`) 계정은 2026-07-31부터 그 상태다.
로그 전체에서 `AnnounceOperatingMode`가 쓴 모양의 줄은 **line 372 하나**다
(`analysis/delivery-latency.md` §0). **오늘의 두절 사이클은 34초다.**

`ObserveOnce`는 `checkOutage`를 두 자리(`:423` 유예 · `:446` 가격 읽기 실패)에서 부른다.
두 자리 다 그 사이클이 **가격을 못 읽은** 사이클이다. 즉 관측이 이미 끊긴 상태에서
알림이 68초를 더 먹고, `Run:359`가 거기에 5초를 더한다 → **다음 관측 시도까지 73초**.

## State mutations and fallbacks

| 대상 | 자리 | 성격 |
|---|---|---|
| `o.outageRaised` | `:778` | **래치** — 두절당 1회만 알린다. 복구 시 `ObserveOnce:439`·`:450`이 푼다 |
| `cycle.Escalated` | `:803` | 보고용 |

- fallback 없음. 승격 실패는 로그 한 줄이고(B5) 사이클은 계속한다.
- goroutine 없음(AST `go_statements: 0`), defer 없음.

## Safety conclusion

- **Safe edit boundary**: a092는 이 함수를 **편집하지 않는다.** 여기 있는 것은 예산이
  **두 번** 걸린다는 사실이다.
- **High-risk impact**: **yes** — §0.4. 이 함수가 도는 사이클은 이미 가격을 못 읽은
  사이클이고, 거기서 68초를 더 쓰면 보호되지 않는 포지션의 무관측 구간이 그만큼 늘어난다.
- **기한을 호출자에 씌우면 안 되는 이유가 여기 있다 — 단 이유는 "잘림"이 아니다.**
  6판까지 이 산출물은 *"`:796`의 `ctx`는 journal 트랜잭션과 공유되므로 짧은 기한을
  씌우면 원장 트랜잭션이 잘린다"*고 적었다. **틀렸다.** `TransitionOperatingMode`의
  AST가 `tx.Commit()` **B25 `:468`** → `AnnounceOperatingMode` **B27/B28 `:478-479`**를
  열거한다 — **announce는 커밋 뒤이고 트랜잭션 밖이다.** 같은 파일의
  「Calls and live bindings」가 이미 *"알림은 커밋 밖이므로 34초가 트랜잭션 안에 갇히지는
  않는다"*고 적고 있었고, `design.md` D1도 2라운드 M5에서 이것을 정정했다.
  **정정이 design에만 착지하고 이 산출물에는 네 판째 오지 않은 것**이 6라운드 차단 B3다.

  실제 이유는 둘이고, design D1 안 B가 정본이다.

  1. **호출자 기한은 그 아래 이탈들에 누적된다.** `:780`의 알림이 기한을 다 쓰면
     `:796`의 `EscalateOperatingMode`는 **이미 만료된 ctx로 `BeginTx`
     (`operating_mode.go:391`)를 부른다** — 트랜잭션이 **잘리는** 게 아니라
     **시작되지 않는다.** 운영 모드 승격 자체가 사라진다.
  2. **경로를 다 덮지 못한다.** `Notify`에 도달하는 자리는 7곳이고 그중 넷은
     exit 루프의 함수를 지나지 않는다(`analysis/notify-reach.md`).

  그러므로 예산은 **`Notify` 아래 transport 계층**에서 정해져야 두 경로에 다 걸리고
  원장에는 안 걸린다. 결론은 6판과 같고 **근거만 참으로 바뀐다.**
