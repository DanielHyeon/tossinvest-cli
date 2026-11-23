# Function Logic Map: `Gateway.withCancelRefusal`

- Source: `internal/execgw/cancel_refusal.go:44-112`
- AST evidence: `ast.json` (**post-edit**, 4b.7 뒤 — source_sha256 `36d892ae…`, 분기 15). Pre-edit 판은 base 와 같은 파일(sha `acd4d7de…`, 분기 13 — 아래 표의 번호).
- **재번호(pre → post, difflib)**: B1 동일 · **새 B2 L53 `if policy.Definitive`·새 B3 L54 `settledCancelRefusal(first)` 삽입**(D7 — 참이면 확정 거절 `dispatch_rejected_by_broker` 반환) · 옛 B2~B13 → 새 B4~B15(조건·반환 무변). 아래 표는 pre-edit 번호다.
- Risk scan: `risk-pattern-report.md` (적중 1: cancel_refusal.go:154 float64 변환 — `retryAfter` 안, 이 함수 밖)
- 판: **pre-edit**. 편집 범위(design D7, Manager 결정 2026-11-24): 첫 dispatch 결과 직후, **엔진 정책(`policy.Definitive`)일 때만** 새 판정 함수(새 파일)를 불러 `already-filled`·`already-canceled`·`already-rejected` 의 모호 결과를 확정 거절(`dispatch_rejected_by_broker`)로 반환. B1(정책 없음·비취소)·a165 흐름(B2~B13) 무편집.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `g.cancelRefusal` | nil(시험) · `EngineCancelRefusal`(Definitive) · `FlattenCancelRefusal`(재시도 2·5s·8s) | 게이트웨이 조립(engine gateway.go:329, cmd/tossctl/flatten.go:251) | nil → 원 dispatch |
| `plan.kind` | CANCEL 만 손댐 | submit | 비취소 → 원 dispatch(B1) |
| `first` | 원 dispatch 결과(`classifyMutation`) | gateway | — |
| 불변식 | 「설정이 없으면 기록은 이전과 같다」(정본 order-execution:540) · 엔진은 같은 attempt 안에서 재전송하지 않는다(:531) · already-processing 판정은 HTTP 409 와 code 가 함께 맞을 때만(:542) | 정본 | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 L45 | 정책 nil 또는 비취소 | 없음 | 원 dispatch | TestA165NilPolicyAndNonCancelKeepTheOriginalOutcome |
| B2 L53 | already-processing 아님 또는 모호 아님 | 없음 | `first` | TestA165CancelRefusalDoesNotOverrideNonAmbiguousDispatch · TestA165EnginePolicyLeavesOther409Ambiguous |
| B3 L56 | 엔진 정책(Definitive) | 꼬리표·`broker_rejected` | 확정 거절 | TestA165EngineAlreadyProcessingIsOneConfirmedRefusalWithoutDelay |
| B4 L61 | 정책 값 0 | 꼬리표 | 원 class | (정책 상수 시험) |
| B5 L70 | 재시도 루프 | 대기·재전송 | — | TestA165FlattenRetriesWithinCumulativeBudget |
| B6~B10 | 대기 상한·창·Sleep 취소 | — | 꼬리표 붙은 첫 결과 | TestA165WindowStartsBeforeFirstDispatch · TestA165CallerCancellationStopsRetry 등 |
| B11 L92 | 창 context 종료 + 재전송 모호/미전송 | — | 첫 결과 | TestA165RetryDeadlineKeepsFirstAmbiguousAsCancelled |
| B12 L96 | 재전송이 already-processing 모호 아님 | — | 재전송 결과(꼬리표 retried) | TestA165FlattenRetriesWithinCumulativeBudget |
| B13 L97 | 재전송 미전송 | — | 첫 결과 | TestA165FlattenRetryNotSentKeepsTheFirstAmbiguousOutcome |
| (편집 예정) | 엔진 정책 && 첫 결과 모호 && code ∈ 세 code(두 자리 일치) | 없음 | 확정 거절 `dispatch_rejected_by_broker` | a198 §4-10~13b |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `dispatch` | 원 전송 | — | AST |
| `alreadyProcessingRefusal` | a165 판정 | 409 + code | AST |
| `decorateCancelRefusal` | latch 가 읽는 꼬리표 | — | AST |
| `g.clk.Sleep`·`Since` | 재시도 창(주입 시계만 — a165 구조 시험) | — | TestA165CancelRefusalUsesOnlyInjectedClockForJudgement |
| (편집 예정) 새 판정 함수 | D7 | `refusalCodes` 판독 재사용 | 새 파일 |

## State mutations and fallbacks

- 없음(결과 값만 바꿔 반환). a165 의 꼬리표는 확정 미수락 latch 의 근거 — D7 결과에는 꼬리표를 붙이지 않는다(그 주문은 이미 종결돼 진입 latch 대상이 아님).

## Safety conclusion

- Safe edit boundary: `first := dispatch(...)` 직후 엔진 정책 갈래에 한 줄짜리 판정 호출. B1~B13 조건·반환 무편집.
- High-risk impact: yes — 취소 attempt 의 종결 상태(IN_DOUBT ↔ FAILED_CONFIRMED)를 바꾼다. 확정 쪽으로 바꾸는 것은 「취소 미성립」이 계약상 확정인 세 code 뿐.
