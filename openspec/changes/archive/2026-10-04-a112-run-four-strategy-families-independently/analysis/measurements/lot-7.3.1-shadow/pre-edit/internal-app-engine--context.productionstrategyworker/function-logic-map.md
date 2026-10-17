# Function Logic Map (편집 전): `productionStrategyWorker`

- Source: `internal/app/engine/strategy_entry_supervisor.go`
- Source SHA-256: `6f1f6804cfd437116c16a48d1526327c28360433f09e0ba3567eb3537442ee5b`
- Signature: `Context.productionStrategyWorker(params=10, results=1)`
- Source range: `412:1`–`424:2`
- AST evidence: `ast.json` — 편집 **전**(HEAD 9e5f3ccf).

## Inputs and invariants

- 편집 계획: 브리프 §5: 두 번째 cycle 클로저를 같은 helper 호출로 바꿈.

## Branches and early returns

- Exact AST return nodes: `418:3`, `420:2`, `423:42`.

| Branch | AST kind | Source location | Condition |
|---|---|---|---|
| B1 | if | 417:2 | `if c == nil {` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `buildProductionStrategyMarketWorker` | 420:9 |
| `c.runProductionStrategyMarketCycle` | 423:49 |

## Safety conclusion

- High-risk(worker 배선) — buildProductionStrategyMarketWorker 인자 · wiringReady 판정 무변경.
