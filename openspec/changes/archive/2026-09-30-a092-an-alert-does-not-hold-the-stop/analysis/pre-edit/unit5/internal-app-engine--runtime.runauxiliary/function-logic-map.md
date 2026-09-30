# Function Logic Map: `Runtime.runAuxiliary`

- Source: `internal/app/engine/auxiliary.go`
- AST evidence: `ast.json` — **편집 전**, :87–114, 분기 2, source_sha256 `b1e02baf829f…`, 추출 HEAD `b01e0cd0`.
- Risk scan: `risk-pattern-report.md`
- 편집 목적(단위 ⑤ · 23.3 K13 · 24.3 M15): 보조 실행자의 정지 로그 이벤트 타입을 실행자별로 — 오늘은 모든 보조 실행자의 정지를 `EventAlertUndelivered` 로 적음(주석이 「a second one would have to bring its own event type」). 일반 등급 이관 실행자(C8)가 두 번째 보조 실행자가 되므로 `AuxiliaryExecutor.StopEvent`(빈 값이면 오늘 값)를 씀.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `aux` | 이름 · Run 필수, OnStop 선택 | 런타임 옵션 | — |
| `ctx` | 런타임 loopCtx | Run | 정상 종료 판정 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `r.gracefulStop(ctx, err)` (:89) | — | 반환(정상 종료) | `TestEveryAuxiliaryExecutorIsStarted` |
| B2 | `aux.OnStop == nil` (:110) | — | 반환 | `TestAnAuxiliaryExecutorThatReturnsDoesNotStopTheEngine` |
| 종단 | — | `r.log(EventAlertUndelivered, …)` :104 · `aux.OnStop` | — | 배달 실행자 정지 시험 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `runAuxiliaryBody` | recover 경계 | 패닉 → `ErrAuxiliaryPanicked` | AST |
| `r.log(obs.EventAlertUndelivered, …)` | 정지 기록 | — | AST — 편집 대상(이벤트 타입) |

## State mutations and fallbacks

- 로그 한 줄 · OnStop 호출.

## Safety conclusion

- Safe edit boundary: 분기 불변, 로그 이벤트 타입 인자만.
- High-risk impact: 중간 — 배달 실행자 정지 기록(그 실행자는 StopEvent 빈 값 → 오늘 값 그대로).
