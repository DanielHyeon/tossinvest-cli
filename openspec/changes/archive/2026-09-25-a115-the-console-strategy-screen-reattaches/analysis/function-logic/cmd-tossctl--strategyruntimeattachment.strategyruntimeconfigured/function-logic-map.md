# Function Logic Map: `strategyRuntimeAttachment.StrategyRuntimeConfigured`

- Source: `cmd/tossctl/httpapi_strategy_attach.go`
- AST evidence: `ast.json` (base `8688f74f` = HEAD — 이 파일은 a115 가 편집하지 않는다, :113–117, 분기 0)
- Risk scan: `risk-pattern-report.md`

> **인용 전용 번들**(a115 freeze 리뷰 P1-4) — a115 는 이 함수를 편집하지 않는다. design D1 의 재사용·펌프 논거가 이 분기들에 기댄다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 자리 | nil(부재)·sentinel·client | wrapper | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| (분기 없음) | 본문 전체 (:113–117) | 자리 상태를 읽고 **무조건 wake** 한 뒤 `reader != nil` 을 답한다(a109 G2) | — | `TestTheAbsenceJudgementIsOneForBothPackages` · `TestTheAbsenceSignalIsOneJudgementForEveryConsumer` · `TestAnUnconfiguredWrapperStillRendersDormant`(콘솔 쪽 판정) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `a.state` | 자리 조회 | 잠금 | AST |
| `a.wake` | 시도 깨우기 | 비차단, rate limit | AST |

## State mutations and fallbacks

- 없음(부작용은 wake). 콘솔 화면이 `strategyprojection.StrategyRuntimeAbsent` 로 이것을 물으면 그 질문이 재부착 시도를 깨운다 — 그래서 페이지는 한 번만 묻는다(변이 K16).

## Safety conclusion

- Safe edit boundary: 편집 없음(인용 전용).
- High-risk impact: no — 조회 전용 projection 재부착, 주문 경로 아님.
