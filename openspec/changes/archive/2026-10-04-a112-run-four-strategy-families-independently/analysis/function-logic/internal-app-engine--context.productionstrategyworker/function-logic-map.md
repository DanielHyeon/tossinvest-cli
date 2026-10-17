# Function Logic Map: `productionStrategyWorker`

- Source: `internal/app/engine/strategy_entry_supervisor.go`
- Source SHA-256: `9cb510c1c9c7f44ac109de5fd8ddac8e6c559adb4d654b1acd65dce19cc9e577`
- Signature: `Context.productionStrategyWorker(params=10, results=1)`
- Source range: `415:1`–`426:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 7.3.1 SHADOW).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 래퍼 하나 — 위와 같은 핀.

## Branches and early returns

- Exact AST return nodes: `421:3, 423:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 420:2 | Context nil → 빈 worker |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `buildProductionStrategyMarketWorker` | 423:9 |
| `c.productionStrategyCycle` | 425:80 |

## State mutations and fallbacks

- 없음(값 구성).

## Safety conclusion

- High-risk(worker 배선) — wiringReady 판정 · 인자 불변.
