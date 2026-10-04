# Function Logic Map: `TestStrategyProposalAuthorityLoadsKRUSConcurrently (시험)`

- Source: `internal/app/engine/strategy_proposal_authority_test.go`
- Source SHA-256: `d05d11be42f3538802baec08718c10782cea1007f608bac95ad206b16a9b57b0`
- Signature: `TestStrategyProposalAuthorityLoadsKRUSConcurrently(params=1, results=0)`
- Source range: `23:1`–`47:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 7.3.1 SHADOW).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 시험 코드 — 생산 판정 없음.

## Branches and early returns

- Exact AST return nodes: `31:3`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 38:2 | 시험 자신의 갈래 |
| B2 | if | 43:2 | 시험 자신의 갈래 |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `time.Date` | 24:9 |
| `testStrategyProposalLoader` | 25:12 |
| `make` | 26:13 |
| `make` | 27:13 |
| `StrategyMarket` | 29:14 |
| `testProposalBatch` | 31:10 |
| `make` | 33:10 |
| `(unnamed)` | 34:5 |
| `a112PairOnly` | 35:11 |
| `loader.collect` | 35:24 |
| `context.Background` | 35:39 |
| `routeReadySchedulePair` | 35:61 |
| `proposalRoutePair` | 35:90 |
| `proposalFXPair` | 35:117 |
| `t.Fatalf` | 39:3 |
| `close` | 41:2 |
| `pair.ResultAuthority` | 44:4 |
| `pair.ResultAuthority` | 44:40 |
| `t.Fatalf` | 45:3 |

## State mutations and fallbacks

- 시험 fixture.

## Safety conclusion

- 시험 코드 — 실주문 · 원장 쓰기 없음(fixture 원장 · Gateway 스파이).
