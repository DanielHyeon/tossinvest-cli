# Function Logic Map (편집 전): `NewRefreshingPairedStrategyEntrySupervisor`

- Source: `internal/app/engine/strategy_entry_supervisor.go`
- Source SHA-256: `6f1f6804cfd437116c16a48d1526327c28360433f09e0ba3567eb3537442ee5b`
- Signature: `Context.NewRefreshingPairedStrategyEntrySupervisor(params=1, results=2)`
- Source range: `382:1`–`410:2`
- AST evidence: `ast.json` — 편집 **전**(HEAD 9e5f3ccf).

## Inputs and invariants

- 편집 계획: 브리프 §5: cycle 클로저 본문을 새 helper(성공 플래그 · 경과 판정 · recover 없는 defer → invalidateShadow · nil 반환 뒤 shadow 단계 비동기 기동) 호출로 바꿈; 반환값은 주기 함수 오류 그대로.

## Branches and early returns

- Exact AST return nodes: `384:3`, `388:3`, `396:5`, `404:3`, `409:2`.

| Branch | AST kind | Source location | Condition |
|---|---|---|---|
| B1 | if | 383:2 | `if c == nil // clk == nil {` |
| B2 | if | 387:2 | `if c.Entry == nil {` |
| B3 | range | 391:2 | `for _, market := range []StrategyMarket{StrategyMarketKR, StrategyMarketUS} {` |
| B4 | if | 403:2 | `if err != nil {` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `errors.New` | 384:15 |
| `fmt.Errorf` | 388:15 |
| `make` | 390:13 |
| `append` | 393:13 |
| `c.runProductionStrategyMarketCycle` | 396:12 |
| `NewStrategyEntrySupervisor` | 400:21 |
| `c.strategyProjectionMu.Lock` | 406:2 |
| `c.strategyProjectionMu.Unlock` | 408:2 |

## Safety conclusion

- High-risk(supervisor 배선) — 주기 오류 · panic 전파 · invokeStrategyCycle 회복 · latch · 삼킴 경로 무변경(핀 (iv) · (v) · central-integrity 신원 시험).
