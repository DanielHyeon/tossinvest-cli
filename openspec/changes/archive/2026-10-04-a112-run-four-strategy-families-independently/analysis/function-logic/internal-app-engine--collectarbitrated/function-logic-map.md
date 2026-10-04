# Function Logic Map: `collectArbitrated (시험)`

- Source: `internal/app/engine/a112_arbitration_test.go`
- Source SHA-256: `44c3a7a4270c2c5b5e965da58dadff94b51df649d3aaf90d786141b1b189bb78`
- Signature: `collectArbitrated(params=5, results=1)`
- Source range: `146:1`–`156:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 7.3.1 SHADOW).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 시험 코드 — 생산 판정 없음.

## Branches and early returns

- Exact AST return nodes: `152:3, 154:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `t.Helper` | 149:2 |
| `testStrategyProposalLoader` | 150:12 |
| `arbitrationBatch` | 152:10 |
| `a112PairOnly` | 154:9 |
| `loader.collect` | 154:22 |
| `context.Background` | 154:37 |
| `routeReadySchedulePair` | 154:59 |
| `arbitrationRoutePair` | 155:3 |
| `proposalFXPair` | 155:69 |

## State mutations and fallbacks

- 시험 fixture.

## Safety conclusion

- 시험 코드 — 실주문 · 원장 쓰기 없음(fixture 원장 · Gateway 스파이).
