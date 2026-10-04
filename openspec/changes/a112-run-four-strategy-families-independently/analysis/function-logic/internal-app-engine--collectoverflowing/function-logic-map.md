# Function Logic Map: `collectOverflowing (시험)`

- Source: `internal/app/engine/a112_coordinator_test.go`
- Source SHA-256: `8581d7275bae005df65081299b204c1eeecb6f2ad6234aa4b00748538f04ad73`
- Signature: `collectOverflowing(params=4, results=1)`
- Source range: `85:1`–`130:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 7.3.1 SHADOW).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 시험 코드 — 생산 판정 없음.

## Branches and early returns

- Exact AST return nodes: `124:3, 129:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | range | 90:2 | 시험 자신의 갈래 |
| B2 | if | 92:3 | 시험 자신의 갈래 |
| B3 | if | 98:3 | 시험 자신의 갈래 |
| B4 | range | 115:3 | 시험 자신의 갈래 |
| B5 | if | 119:4 | 시험 자신의 갈래 |
| B6 | range | 126:2 | 시험 자신의 갈래 |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `t.Helper` | 86:2 |
| `familyScoresForTest` | 87:12 |
| `arbitrationLineageDigests` | 88:34 |
| `make` | 89:13 |
| `len` | 89:52 |
| `strategyrouter.NewOwnerKey` | 91:15 |
| `t.Fatal` | 93:4 |
| `strategyrouter.MultiCandidateRouteFixture` | 95:19 |
| `t.Fatal` | 99:4 |
| `append` | 101:13 |
| `strategy.ApprovedSnapshotForTest` | 102:14 |
| `strategyrouter.ProductionRouteAuthorityFromRequestForTest` | 103:14 |
| `len` | 108:18 |
| `testStrategyProposalLoader` | 112:12 |
| `make` | 114:13 |
| `len` | 114:52 |
| `strategyflow.AcceptedResultForAuthorityTest` | 116:19 |
| `riskLoaderDescriptor` | 116:63 |
| `target.Approved.Symbol` | 117:24 |
| `now.Add` | 118:5 |
| `now.Add` | 118:28 |
| `t.Fatal` | 120:5 |
| `target.Approved.Symbol` | 122:11 |
| `strategyproposal.ProductionBatchAuthorityMultiLaneForTest` | 124:10 |
| `change` | 127:3 |
| `a112PairOnly` | 129:9 |
| `loader.collect` | 129:22 |
| `context.Background` | 129:37 |
| `routeReadySchedulePair` | 129:59 |
| `proposalFXPair` | 129:96 |

## State mutations and fallbacks

- 시험 fixture.

## Safety conclusion

- 시험 코드 — 실주문 · 원장 쓰기 없음(fixture 원장 · Gateway 스파이).
