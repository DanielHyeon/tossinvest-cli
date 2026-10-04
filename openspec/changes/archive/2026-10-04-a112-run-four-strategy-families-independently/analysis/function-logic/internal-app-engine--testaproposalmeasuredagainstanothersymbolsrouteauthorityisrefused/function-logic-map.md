# Function Logic Map: `TestAProposalMeasuredAgainstAnotherSymbolsRouteAuthorityIsRefused (시험)`

- Source: `internal/app/engine/a112_arbitration_test.go`
- Source SHA-256: `44c3a7a4270c2c5b5e965da58dadff94b51df649d3aaf90d786141b1b189bb78`
- Signature: `TestAProposalMeasuredAgainstAnotherSymbolsRouteAuthorityIsRefused(params=1, results=0)`
- Source range: `238:1`–`277:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 7.3.1 SHADOW).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 시험 코드 — 생산 판정 없음.

## Branches and early returns

- Exact AST return nodes: `269:3`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 242:2 | 시험 자신의 갈래 |
| B2 | if | 249:2 | 시험 자신의 갈래 |
| B3 | if | 266:3 | 시험 자신의 갈래 |
| B4 | if | 273:2 | 시험 자신의 갈래 |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `time.Date` | 239:9 |
| `strategyrouter.NewOwnerKey` | 241:14 |
| `t.Fatal` | 243:3 |
| `arbitrationLineageDigests` | 245:34 |
| `strategyrouter.MultiCandidateRouteFixture` | 246:18 |
| `t.Fatal` | 250:3 |
| `strategy.ApprovedSnapshotForTest` | 252:49 |
| `strategyrouter.ProductionRouteAuthorityFromRequestForTest` | 253:10 |
| `familyScoresForTest` | 254:4 |
| `testStrategyProposalLoader` | 260:12 |
| `strategyflow.AcceptedResultForAuthorityTest` | 264:18 |
| `riskLoaderDescriptor` | 264:62 |
| `now.Add` | 265:78 |
| `now.Add` | 265:101 |
| `t.Fatal` | 267:4 |
| `strategyproposal.ProductionBatchAuthorityMultiLaneForTest` | 269:10 |
| `loader.collect` | 272:13 |
| `context.Background` | 272:28 |
| `routeReadySchedulePair` | 272:50 |
| `proposalFXPair` | 272:87 |
| `string` | 273:70 |
| `t.Fatalf` | 275:3 |

## State mutations and fallbacks

- 시험 fixture.

## Safety conclusion

- 시험 코드 — 실주문 · 원장 쓰기 없음(fixture 원장 · Gateway 스파이).
