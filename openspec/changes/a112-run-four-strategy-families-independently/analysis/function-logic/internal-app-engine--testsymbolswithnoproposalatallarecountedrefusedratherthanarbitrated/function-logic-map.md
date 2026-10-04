# Function Logic Map: `TestSymbolsWithNoProposalAtAllAreCountedRefusedRatherThanArbitrated (시험)`

- Source: `internal/app/engine/a112_arbitration_test.go`
- Source SHA-256: `44c3a7a4270c2c5b5e965da58dadff94b51df649d3aaf90d786141b1b189bb78`
- Signature: `TestSymbolsWithNoProposalAtAllAreCountedRefusedRatherThanArbitrated(params=1, results=0)`
- Source range: `281:1`–`297:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 7.3.1 SHADOW).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 시험 코드 — 생산 판정 없음.

## Branches and early returns

- Exact AST return nodes: `285:3`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 290:2 | 시험 자신의 갈래 |
| B2 | if | 293:2 | 시험 자신의 갈래 |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `time.Date` | 282:9 |
| `testStrategyProposalLoader` | 283:12 |
| `strategyproposal.ProductionBatchAuthorityMultiLaneForTest` | 285:10 |
| `loader.collect` | 287:13 |
| `context.Background` | 287:28 |
| `routeReadySchedulePair` | 287:50 |
| `arbitrationRoutePair` | 288:3 |
| `familyScoresForTest` | 288:32 |
| `proposalFXPair` | 289:3 |
| `t.Fatalf` | 291:3 |
| `t.Fatalf` | 294:3 |

## State mutations and fallbacks

- 시험 fixture.

## Safety conclusion

- 시험 코드 — 실주문 · 원장 쓰기 없음(fixture 원장 · Gateway 스파이).
