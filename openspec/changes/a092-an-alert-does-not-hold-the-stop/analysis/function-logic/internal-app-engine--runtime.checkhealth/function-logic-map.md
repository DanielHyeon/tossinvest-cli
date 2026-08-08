# Function Logic Map: `Runtime.CheckHealth`

- Source: `internal/app/engine/runtime.go` (368-388)
- AST evidence: `ast.json` — branches 5, **returns 0**, calls 4, assignments 1,
  defers 0, go_statements 0
- Risk scan: `risk-pattern-report.md`

**1초 루프의 몸통이다.** `superviseHealth`(`:346-360`)가 `r.interval`마다 이 함수를 부르고,
프로덕션은 override를 주지 않으므로 그 주기는 `DefaultHealthInterval = time.Second`
(`runtime.go:84`) — **저장소에서 가장 짧은 루프 주기**다. a092는 편집하지 않는다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `r.opts.Loops` | 프로덕션 4개(reconcile·exit·filldetect·strategy-entry) | `cmd/tossctl/engine.go` | 빈 슬라이스면 B1이 즉시 끝난다 |
| `loop.Health` | nil 허용 | `LoopSpec` | nil이면 B2가 건너뛴다 |
| `r.threshold` | > 0 | `NewRuntime` | — |
| `r.escalated` | 래치 맵 | `takeLatch`·`clearLatch` | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return | `Notify` 도달 |
|---|---|---|---|---|
| B1 `:369` | `range r.opts.Loops` | — | — | — |
| B2 `:370` | `loop.Health == nil` | 없음 | `continue` | ❌ |
| B3 `:374` | `consecutive < r.threshold` | — | `continue` `:381` | ❌ |
| B4 `:375` | `consecutive == 0` (복구) | `r.clearLatch(loop.Name)` `:379` | — | ❌ |
| B5 `:383` | `!r.takeLatch(loop.Name)` — 이미 승격함 | 없음 | `continue` | ❌ |
| — `:386` | — | **`r.escalate(ctx, loop, consecutive)`** | — | ✅ |

**AST returns 0이다.** 이 함수는 반환문이 하나도 없고 이탈은 전부 `continue`다.
즉 **한 번의 호출이 루프 4개를 모두 훑고, 승격 자격을 얻은 루프마다 `escalate`을 부른다** —
최악은 1회 호출에 `escalate` 4회다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `loop.Health.ConsecutiveFailures` `:373` | 연속 실패 수 | 즉시 | AST calls |
| `r.clearLatch` `:379` | 복구 시 래치 해제 | 즉시 | AST calls |
| `r.takeLatch` `:383` | 중복 승격 차단 | 즉시 | `internal-app-engine--runtime.takelatch` FLM |
| `r.escalate` `:386` | **알림 + 운영모드 승격** | **동기·부분적으로만 유계** | `internal-app-engine--runtime.escalate` FLM |

## State mutations and fallbacks

- 이 함수 자체는 상태를 바꾸지 않는다. AST assignments 1은 `consecutive :=`다.
- 상태 변경은 전부 `takeLatch`/`clearLatch` 안에 있다.
- goroutine 없음, defer 없음.

## Safety conclusion

- **Safe edit boundary**: a092는 편집하지 않는다.
- **High-risk impact**: no(직접) — 하지만 **engine-safety 예산 기준을 정할 때
  이 함수의 주기가 반례가 된다.**
- **a092가 여기서 얻는 사실 두 가지**
  1. 이 루프의 주기는 **1초**이고 exit 관측 주기(5초)보다 짧다. 그러므로
     "가장 짧은 루프 주기"라는 표현으로 예산 기준을 쓰면 거짓이다(2라운드 C2).
  2. **B5의 래치가 있어서** 1초 루프가 매초 전송을 부르지는 않는다. 두절 1회당 1회다.
     이것이 감독자를 예산 기준에서 면제할 수 있는 근거이고, **유일한 근거다.**
