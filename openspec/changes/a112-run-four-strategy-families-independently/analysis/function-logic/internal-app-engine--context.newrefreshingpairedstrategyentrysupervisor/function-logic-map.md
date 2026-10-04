# Function Logic Map: `NewRefreshingPairedStrategyEntrySupervisor`

- Source: `internal/app/engine/strategy_entry_supervisor.go`
- Source SHA-256: `9cb510c1c9c7f44ac109de5fd8ddac8e6c559adb4d654b1acd65dce19cc9e577`
- Signature: `Context.NewRefreshingPairedStrategyEntrySupervisor(params=1, results=2)`
- Source range: `385:1`–`413:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 7.3.1 SHADOW).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 두 생산 자리가 같은 래퍼를 쓴다 — AST 핀 `TestTheCycleClosureStartsTheShadowOnlyAfterANilCycleAndDiscardsOtherwise`.

## Branches and early returns

- Exact AST return nodes: `387:3, 391:3, 407:3, 412:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 386:2 | Context · 시계 nil |
| B2 | if | 390:2 | 진입 관문 없음 |
| B3 | range | 394:2 | 시장 순회 |
| B4 | if | 406:2 | supervisor 생성 오류 |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `errors.New` | 387:15 |
| `fmt.Errorf` | 391:15 |
| `make` | 393:13 |
| `append` | 396:13 |
| `c.productionStrategyCycle` | 400:11 |
| `NewStrategyEntrySupervisor` | 403:21 |
| `c.strategyProjectionMu.Lock` | 409:2 |
| `c.strategyProjectionMu.Unlock` | 411:2 |

## State mutations and fallbacks

- supervisor 등록.

## Safety conclusion

- High-risk(supervisor 배선) — 주기 오류 · panic 전파 · 회복 · latch · 삼킴 경로 불변(핀 (iv) · (v) · central-integrity 신원).
