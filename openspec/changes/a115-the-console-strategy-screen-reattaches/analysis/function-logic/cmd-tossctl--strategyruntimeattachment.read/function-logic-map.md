# Function Logic Map: `strategyRuntimeAttachment.Read`

- Source: `cmd/tossctl/httpapi_strategy_attach.go`
- AST evidence: `ast.json` (base `8688f74f` = HEAD — 이 파일은 a115 가 편집하지 않는다, :119–137, 분기 3)
- Risk scan: `risk-pattern-report.md`

> **인용 전용 번들**(a115 freeze 리뷰 P1-4) — a115 는 이 함수를 편집하지 않는다. design D1 의 재사용·펌프 논거가 이 분기들에 기댄다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 자리 상태 | reader·seat·failed | `a.state()` | 부재 → 오류(지어내지 않음) |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if wanted {` (:121) | 시도 대상(failed)이면 요청이 재부착을 깨운다(비차단) | — | `TestTheRequestPathNeverWaitsForADial` |
| B2 | `if reader == nil {` (:124) | 부재(reader nil) → 오류 — 부재를 스냅샷으로 짓지 않음(design 대안 2 기각의 근거) | — | `TestTheDaemonAttachesWhenTheEngineComesUpLater` |
| B3 | `if a.observe(ctx, seat, err) {` (:131) | 읽기 실패 판정 → 즉시 wake(다음 요청을 기다리지 않음) | — | `TestTheRequestPathNeverWaitsForADial` · `TestTheDaemonReattachesAfterTheEngineRestarts` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `a.state` | 자리 조회 | 잠금 | AST |
| `a.wake` | 시도 깨우기 | 비차단 | AST |
| `reader.Read` | 스냅샷 | 요청 ctx | AST |
| `a.observe` | 결과 판정 | seat 대조 | AST |

## State mutations and fallbacks

- 자리 상태는 observe 가 바꾼다. 요청 경로는 dial 하지 않는다(dial 은 attempt 에만) — spec 의 SHALL NOT(요청 경로 dial 금지)의 근거.

## Safety conclusion

- Safe edit boundary: 편집 없음(인용 전용).
- High-risk impact: no — 조회 전용 projection 재부착, 주문 경로 아님.
