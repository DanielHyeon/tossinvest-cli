# Function Logic Map: `Runtime.Run`

- Source: `internal/app/engine/runtime.go` (261-332)
- AST evidence: `ast.json` — branches 4, returns 3, calls 29, assignments 7,
  **defers 3, go_statements 2**
- Risk scan: `risk-pattern-report.md`

**17판이 배달 루프를 붙이는 자리.** a092는 이 함수를 편집하지 않는다 —
루프는 `RuntimeOptions.Loops`로 들어가고 B3의 `range`가 그것을 그대로 띄운다.
**편집 없이 동작이 바뀌는 함수**이므로 그 결합을 여기 적는다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `r.opts.Recover` | nil이면 건너뜀 | `NewRuntime` 배선 | B1 `:262`, B2 `:267` |
| `r.opts.Loops` | `NewRuntime`이 이미 검증함 | 같은 위 | 검증 없음 — 여기서는 신뢰한다 |
| `ctx` | 호출자(엔진 명령)의 것 | `tossctl engine run` | 취소가 정상 종료 경로다 |

**계약이 doc comment `:252-260`에 있다**: 반환 시점에 **이 함수가 띄운 모든
goroutine이 이미 반환했다.** 그래서 호출자가 곧바로 journal을 닫아도 아직 쓰는
루프와 경합하지 않는다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return |
|---|---|---|---|
| B1 `:262` | `Recover != nil` | 복구 시퀀스 실행 `:267` | — |
| B2 `:267` | 복구 실패 | **루프를 하나도 안 띄운다** | 오류 `:268` |
| **B3 `:277`** | **`range r.opts.Loops`** | 루프마다 goroutine `:279` | — |
| B4 `:304` | `gracefulStop` — 첫 정지가 취소였다 | heartbeat 로그 `:305` | `nil` `:308` |
| — `:315-331` | 루프가 스스로 반환했다 | **critical 알림 `:316` + 로그 `:330`** | `ErrLoopFailed` 감싼 오류 |

**B3이 17판의 결합 지점이다.** 루프를 하나 더하면 goroutine이 하나 더 뜬다.
그리고 `:297` `first := <-stops`가 **첫 정지 하나로 전부를 내린다** — `:298`
`cancel()`이 모든 루프의 ctx를 끊는다.

`:293-296`의 주석이 그 이유를 적는다: 부분 생존 금지는 정돈의 문제가 아니라,
청산을 관측하지 않으면서 대사하는 엔진이 멈춘 엔진보다 나쁘다는 판단이다.

**그러므로 배달 루프가 반환하면 엔진 전체가 내려간다.** 17판이 받아들이는 대가다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `r.opts.Recover` `:267` | 재시작 복구 | 실패하면 아무것도 시작 안 함 | AST calls |
| `context.WithCancel` `:272` | 루프 ctx | `defer cancel()` `:273` | AST defers |
| `loop.Run` `:281` | **각 루프 본체 — goroutine 안** | 반환이 곧 정지 신호 | AST go_statements **2** |
| `r.superviseHealth` `:290` | 건강 폴링 — goroutine 안 | 두 번째 go | AST go_statements |
| `cancel` `:298` | 전부 내리기 | — | AST calls |
| `wg.Wait` `:300`·`supervisorWG.Wait` `:301` | **전부 반환할 때까지 대기** | 무기한 — 루프가 안 끝나면 여기서 멈춘다 | AST calls |
| `drain` `:302` | 나머지 정지 사유 수집 | — | AST calls |
| `r.alert` `:316` | critical 알림 | 이 함수의 유일한 알림 | AST calls |

**`go_statements`가 2개**(`:279` 루프 fan-out, `:288` 감독자). `defers`가 3개
(`:273` cancel, `:280` wg.Done, `:289` supervisorWG.Done).

**`wg.Wait()`에 기한이 없다.** 배달 루프가 ctx 취소에 반응하지 않으면
엔진 종료가 그 루프를 기다리며 멈춘다 — 17판 구현이 지켜야 할 제약이다.

## State mutations and fallbacks

| 대상 | 자리 | 성격 |
|---|---|---|
| `stops` 채널 | `:275`, `:281` | 버퍼 = 루프 수. **루프를 늘리면 버퍼도 는다**(`len` 사용) |
| `loopCtx` | `:272`, `:298` | 취소 전파 |
| 구조화 로그 | `:305`·`:330` | — |
| critical 알림 | `:316` | `EventEngineLoopFailed` |

- fallback 없음. 정지는 되돌리지 않는다 — **아무것도 재시작하지 않는다**(`:258`).

## Safety conclusion

- **Safe edit boundary**: a092는 편집하지 않는다. 배달 루프는 `Loops` 슬라이스로
  들어가고 B3가 그것을 그대로 처리한다.
- **High-risk impact**: yes — 엔진 생사의 자리다. 17판이 더하는 결합 두 가지를
  명시한다:
  1. **배달 루프가 반환하면 엔진 전체가 정지한다.** 그러므로 배달 루프의 `Run`은
     ctx가 끝날 때까지 반환하면 안 된다 — 전송 실패는 반환 사유가 아니다.
  2. **배달 루프가 ctx 취소에 반응하지 않으면 종료가 멈춘다**(`wg.Wait` `:300`).
     주기 대기는 취소 가능한 sleep이어야 한다.
- **§0.3에 대한 영향 없음**: 이 함수는 손절 경로에 없다. 배달 루프가 죽어
  엔진이 내려가는 경우는 알림이 아니라 **엔진 정지**의 문제이고,
  그때 손절이 사라진다는 사실은 이미 기록된 별도 리스크다.
- **§6.0 R17-8**(best-effort도 루프를 안 붙잡는다)과 **R17-10**(감독 아래)이
  위 두 제약을 관측한다.
