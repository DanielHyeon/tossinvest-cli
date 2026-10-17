# Function Logic Map: `strategyDispatchGatewaySpy.ObserveStrategyEntryGate`

- Source: `internal/app/engine/strategy_dispatch_cycle_test.go`
- Source SHA-256: `b0b9734d75c5e4fafa2b2033bd9c3d9660af813aa7d485b1840d2a1f8ebcd958`
- Signature: `strategyDispatchGatewaySpy.ObserveStrategyEntryGate(params=3, results=2)`
- Source range: `46:1`–`57:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 5.2.2.2).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 시험 코드 — 생산 동작 없음(비례 원칙: FLM 의무의 무거운 규율 대상 아님, 게이트 모양만 채움)

## Branches and early returns

- Exact AST return nodes: `51:3, 54:3, 56:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 50:2 | 시장 단위 거절(편집 불변) |
| B2 | if | 53:2 | **(새)** 종목 단위 거절 |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `spy.mu.Lock` | 47:2 |
| `spy.mu.Unlock` | 48:8 |
| `execgw.StrategyEntryGateAuthorityForTest` | 56:9 |
| `strings.Repeat` | 56:63 |

## State mutations and fallbacks

- 스파이의 관측 횟수 map 만 쓴다.

## Safety conclusion

- High-risk 아님(시험 스파이). 실주문 없음.
