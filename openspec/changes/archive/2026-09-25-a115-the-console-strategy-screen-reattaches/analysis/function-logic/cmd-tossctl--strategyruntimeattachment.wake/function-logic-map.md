# Function Logic Map: `strategyRuntimeAttachment.wake`

- Source: `cmd/tossctl/httpapi_strategy_attach.go`
- AST evidence: `ast.json` (base `8688f74f` = HEAD — 이 파일은 a115 가 편집하지 않는다, :166–177, 분기 1)
- Risk scan: `risk-pattern-report.md`

> **인용 전용 번들**(a115 freeze 리뷰 P1-4) — a115 는 이 함수를 편집하지 않는다. design D1 의 재사용·펌프 논거가 이 분기들에 기댄다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `a.interval`·`a.lastTry`·`a.trying`·`a.ctx` | 간격>0 운영 30s(콘솔 전용 변수) | wrapper 필드 | early return |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if a.trying \|\| tooSoon \|\| a.ctx.Err() != nil {` (:170) | single-flight(trying)·rate limit(tooSoon)·수명 종료(ctx.Err) 중 하나면 early return; 아니면 trying=true·lastTry=now·`go attempt()` | — | `TestTheAttemptIsSingleFlight` · `TestTheAttemptIsRateLimited` · `TestTheConsoleStrategyPumpStopsWithTheConsole` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `a.now` | 시계 | 주입 가능 | AST |
| `a.attempt` | 시도(goroutine) | 요청 경로 밖 | AST |

## State mutations and fallbacks

- trying·lastTry 갱신. 무조건 wake 의 비용이 「간격당 시도 1회」로 고정되는 근거(:110–112 주석과 일치) — 콘솔 펌프 결정(freeze P1-2)의 근거. 주의: interval≤0 이면 tooSoon 이 항상 false — 콘솔 펌프 쪽 가드는 별도(`pumpConsoleStrategyRuntime`, 변이 K9).

## Safety conclusion

- Safe edit boundary: 편집 없음(인용 전용).
- High-risk impact: no — 조회 전용 projection 재부착, 주문 경로 아님.
