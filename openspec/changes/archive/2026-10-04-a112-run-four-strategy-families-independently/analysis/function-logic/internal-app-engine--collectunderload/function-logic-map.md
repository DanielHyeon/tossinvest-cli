# Function Logic Map: `collectUnderLoad (시험)`

- Source: `internal/app/engine/a112_family_gate_test.go`
- Source SHA-256: `49511cdebc1ba5506737d375b775aa2a533b358701fb531f33f41051f2400fcf`
- Signature: `collectUnderLoad(params=6, results=1)`
- Source range: `62:1`–`82:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 7.3.1 SHADOW).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 시험 코드 — 생산 판정 없음.

## Branches and early returns

- Exact AST return nodes: `71:3, 76:3, 81:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `t.Helper` | 65:2 |
| `time.Date` | 66:9 |
| `testStrategyProposalLoader` | 67:12 |
| `arbitrationBatch` | 71:10 |
| `collect` | 78:13 |
| `loader.withStrategyLanes` | 78:13 |
| `context.Background` | 78:53 |
| `routeReadySchedulePair` | 78:75 |
| `arbitrationRoutePair` | 79:3 |
| `familyScoresForTest` | 79:32 |
| `proposalFXPair` | 80:3 |
| `pair.forMarket` | 81:9 |

## State mutations and fallbacks

- 시험 fixture.

## Safety conclusion

- 시험 코드 — 실주문 · 원장 쓰기 없음(fixture 원장 · Gateway 스파이).
