# Function Logic Map: `TestStrategyProposalAuthorityKeepsMarketFailureLocal (시험)`

- Source: `internal/app/engine/strategy_proposal_authority_test.go`
- Source SHA-256: `d05d11be42f3538802baec08718c10782cea1007f608bac95ad206b16a9b57b0`
- Signature: `TestStrategyProposalAuthorityKeepsMarketFailureLocal(params=1, results=0)`
- Source range: `49:1`–`65:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 7.3.1 SHADOW).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 시험 코드 — 생산 판정 없음.

## Branches and early returns

- Exact AST return nodes: `54:4, 56:3`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 53:3 | 시험 자신의 갈래 |
| B2 | if | 59:2 | 시험 자신의 갈래 |
| B3 | if | 62:2 | 시험 자신의 갈래 |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `time.Date` | 50:9 |
| `testStrategyProposalLoader` | 51:12 |
| `errors.New` | 54:56 |
| `testProposalBatch` | 56:10 |
| `loader.collect` | 58:13 |
| `context.Background` | 58:28 |
| `routeReadySchedulePair` | 58:50 |
| `proposalRoutePair` | 58:79 |
| `proposalFXPair` | 58:106 |
| `t.Fatalf` | 60:3 |
| `t.Fatalf` | 63:3 |

## State mutations and fallbacks

- 시험 fixture.

## Safety conclusion

- 시험 코드 — 실주문 · 원장 쓰기 없음(fixture 원장 · Gateway 스파이).
