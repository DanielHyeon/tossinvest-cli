# Function Logic Map: `ExitObserver.ObserveOnce`

- Source: `internal/app/engine/exitloop.go` (**편집 뒤** `456`–`523`; 편집 전 `444`–`509`, base `381e73da`)
- Qualified: `ExitObserver.ObserveOnce`
- AST evidence: `ast.json` (**편집 뒤** `source_sha256` fba676d31e44590c… — 편집 전 0733bd8641ed8c41…) — branches 8 · returns 5 · **defers 1**(`:458` `o.endExitOrderCycle`) · go 0
- Risk scan: `risk-pattern-report.md` (파일 단위)
- 작성 시점: **편집 전**(tasks 0.5). 선행 번들: a090 아카이브 `…/2026-09-30-a090-an-unobserved-position-is-counted/analysis/function-logic/internal-app-engine--exitobserver.observeonce/`
  (같은 분기 8 · return 5, 좌표만 a091 뒤로 밀림).

**역할.** exit 관측 한 주기: 체결 감지 양보 → 작업 집합 → 일괄 가격 읽기 → 포지션별 판정 → a090 미관측 정리.
호출자(CodeGraph `callers ObserveOnce`): 생산 `ExitObserver.Run`(`exitloop.go:382`) · `Tracer.Run`(`tracer.go:273`) + 시험 9.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `ctx` | 취소 가능 | `Run` | 하위 호출이 ctx 오류를 돌려줌 |
| `o.opts.SLO` | nil 가능 | 생산: `detectorPressure` | nil 이면 양보 없음 |
| 관측자 메모리(`lastObserved`·`outageRaised`·a090 기록) | 단일 goroutine | 이 관측자 | 재시작 시 잃음(파일 머리 계약) |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `:448` `if o.opts.SLO != nil && …FillDetectionBehind()` | `cycle.Deferred`, `checkOutage` | `return cycle` `:454` | a129: 양보 주기에도 체결 따라가기가 돈다 |
| B2 | `:458` `if err != nil` (`workingSet`) | `cycle.Err` | `return` `:460` | 무편집 |
| B3 | `:462` `if len(states) == 0` | `settleUnobserved`, 두절 시계 리셋 | `return` `:472` | a129: 전량 청산 뒤(보유 0) 주기에도 체결 따라가기가 돈다 — 손절 체결의 주 경로 |
| B4 | `:475` `if err != nil` (`observe`) | `cycle.Err`, `checkOutage` | `return` `:478` | 무편집 |
| B5 | `:484` `for … range states` | 포지션별 판정 | — | 무편집(판정이 따라가기보다 먼저) |
| B6 | `:486` `if !ok` (응답 없음) | `noteUnobservedCause` | `continue` | 무편집 |
| B7 | `:494` `if !o.quoteUsable(quote)` | `noteUnobservedCause` | `continue` | 무편집 |
| B8 | `:502` `if err := o.judge(…); err != nil && cycle.Err == nil` | `cycle.Err` | — | 무편집 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `o.opts.SLO.FillDetectionBehind` | 양보 판정 | bool | AST `:448` |
| `o.checkOutage` | 계정 두절 사다리 | 내부 처리 | AST `:453` `:477` |
| `o.workingSet` | 포지션 · exit state | 오류 → B2 | AST `:457` |
| `o.observe` | **유일한 브로커 읽기**(가격) | 오류 → B4 | AST `:474` |
| `o.judge` | 판정 · 기록 · 제출 | 첫 오류만 `cycle.Err` | AST `:502` |
| `o.settleUnobserved` · `noteUnobservedCause` · `noteJudged` | a090 | 내부 처리 | AST |

Live config binding: 없음.

## State mutations and fallbacks

- `o.lastObserved` · `o.outageRaised`(B3 · B4 뒤), `cycle.*`.
- a129 가 더하는 것: 함수 첫 문장 `defer o.followExitFills(ctx)` — 모든 return 경로(B1 · B2 · B3 · B4 · 끝)에서 판정·미관측 정리 **뒤에** 돈다.
  원장 읽기(`LookupFill`)만 하고 `cycle` 을 바꾸지 않는다(반환값이 `defer` 전에 평가된 `cycle` 값 그대로). `ExitOrderLog` nil 이면 즉시 돌아간다.

## Safety conclusion

- Safe edit boundary: 첫 문장에 `defer` 한 줄. 분기·순서·반환값 무변경. 따라가기는 판정 뒤에만 돌고 브로커를 부르지 않는다.
- High-risk impact: yes — 손절 판정 주기. 판정 앞에 새 일이 서지 않음을 a129 시험(같은 주기에 손절 제출과 체결 줄 순서)으로 잰다.

## 편집 뒤 (a129)

분기 수·종류·순서 불변(8). 좌표만 +14(옵션 주석 · 구조체 필드 · defer 두 줄): B1 `:462` · B2 `:472` · B3 `:476` · B4 `:489` · B5 `:498` · B6 `:500` · B7 `:508` · B8 `:516`,
return `:468` `:474` `:486` `:492` `:522`. 유일한 편집은 첫 문장 `defer o.endExitOrderCycle(ctx)`(`:458`) — 이 주기 대기열의 줄 쓰기와 체결 따라가기.
결과 `ExitCycle` 은 이름 없는 반환값이라 defer 가 바꿀 수 없다. `ExitOrderLog` nil 이면 defer 는 첫 줄에서 돌아간다.
