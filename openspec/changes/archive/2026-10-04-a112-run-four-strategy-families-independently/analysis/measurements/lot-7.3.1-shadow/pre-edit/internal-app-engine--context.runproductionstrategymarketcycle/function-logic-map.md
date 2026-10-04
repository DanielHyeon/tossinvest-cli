# Function Logic Map (편집 전): `runProductionStrategyMarketCycle`

- Source: `internal/app/engine/strategy_entry_supervisor.go`
- Source SHA-256: `6f1f6804cfd437116c16a48d1526327c28360433f09e0ba3567eb3537442ee5b`
- Signature: `Context.runProductionStrategyMarketCycle(params=3, results=1)`
- Source range: `514:1`–`573:2`
- AST evidence: `ast.json` — 편집 **전**(HEAD 9e5f3ccf).

## Inputs and invariants

- 편집 계획: 브리프 §5 핀 ①: evaluate 호출에 인자 index 5 `fresh.<shadow 필드>.forMarket(market)` 하나 추가; 그 외 무변경, 마지막 문장 dispatch 유지.

## Branches and early returns

- Exact AST return nodes: `517:3`, `541:3`, `547:3`, `550:3`, `572:2`.

| Branch | AST kind | Source location | Condition |
|---|---|---|---|
| B1 | if | 516:2 | `if err != nil {` |
| B2 | if | 540:2 | `if err != nil {` |
| B3 | if | 543:2 | `if err := lanes.evaluate(ctx, market,` |
| B4 | if | 549:2 | `if fresh.dispatch == nil {` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `c.refreshPairedStrategyEntryProductionAssembly` | 515:16 |
| `c.productionStrategyLanes` | 539:16 |
| `lanes.evaluate` | 543:12 |
| `restore.Activation.Generation` | 544:3 |
| `fresh.schedule.forMarket` | 544:3 |
| `familyActivation` | 545:3 |
| `fresh.proposals.forMarket` | 545:3 |
| `strategyLaneInputs` | 546:3 |
| `fresh.proposals.forMarket` | 546:36 |
| `dispatchStrategyMarketHandoffs` | 572:9 |
| `dispatchHandoffs` | 572:72 |
| `fresh.proposals.forMarket` | 572:72 |

## Safety conclusion

- High-risk(시장 주기) — dispatch 문장 · 원천 · 순서 무변경(:169-175 핀), Args[2] 잠금 복구 세대 핀 보존, dispatch 앞 shadow 일은 값 운반 하나.
