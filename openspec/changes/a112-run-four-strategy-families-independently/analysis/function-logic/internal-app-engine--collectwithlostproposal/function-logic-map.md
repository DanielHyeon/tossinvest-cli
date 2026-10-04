# Function Logic Map: `collectWithLostProposal (시험)`

- Source: `internal/app/engine/a112_lost_proposal_test.go`
- Source SHA-256: `18c18285132166da53dc877a066324b9c848541850247af220aba98af7f511b0`
- Signature: `collectWithLostProposal(params=3, results=1)`
- Source range: `50:1`–`91:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 7.3.1 SHADOW).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 시험 코드 — 생산 판정 없음.

## Branches and early returns

- Exact AST return nodes: `87:3, 90:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | range | 57:2 | 시험 자신의 갈래 |
| B2 | if | 59:3 | 시험 자신의 갈래 |
| B3 | if | 65:3 | 시험 자신의 갈래 |
| B4 | if | 84:3 | 시험 자신의 갈래 |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `t.Helper` | 53:2 |
| `familyScoresForTest` | 54:12 |
| `arbitrationLineageDigests` | 55:34 |
| `make` | 56:13 |
| `strategyrouter.NewOwnerKey` | 58:15 |
| `t.Fatal` | 60:4 |
| `strategyrouter.MultiCandidateRouteFixture` | 62:19 |
| `t.Fatal` | 66:4 |
| `append` | 68:13 |
| `strategy.ApprovedSnapshotForTest` | 69:14 |
| `strategyrouter.ProductionRouteAuthorityFromRequestForTest` | 70:14 |
| `len` | 75:18 |
| `testStrategyProposalLoader` | 79:12 |
| `strategyflow.AcceptedResultForAuthorityTest` | 82:18 |
| `riskLoaderDescriptor` | 82:62 |
| `now.Add` | 83:71 |
| `now.Add` | 83:94 |
| `t.Fatal` | 85:4 |
| `strategyproposal.ProductionBatchAuthorityWithFaultForTest` | 87:10 |
| `a112PairOnly` | 90:9 |
| `loader.collect` | 90:22 |
| `context.Background` | 90:37 |
| `routeReadySchedulePair` | 90:59 |
| `proposalFXPair` | 90:96 |

## State mutations and fallbacks

- 시험 fixture.

## Safety conclusion

- 시험 코드 — 실주문 · 원장 쓰기 없음(fixture 원장 · Gateway 스파이).
