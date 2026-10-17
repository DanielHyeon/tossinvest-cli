# Function Logic Map (편집 전): `collectOverflowing`

- Source: `internal/app/engine/a112_coordinator_test.go`
- Source SHA-256: `157ac586f802f30b4089bdbf5f1bc075af45e1b3a78e7e554b07a15e3f1a5832`
- Signature: `collectOverflowing(params=3, results=1)`
- Source range: `68:1`–`110:2`
- AST evidence: `ast.json` — 편집 **전**(a112 base 재고정 1e25b3a3 — 8.5 응답 로트(6f5b0df6) 편집 전(178cc196)).

## Inputs and invariants

- 편집 계획: 8.5 응답 로트 ⑦: 판정 활성화 carry 단언 추가(B10/B12 — 보이스 3 P2-1); collectOverflowing 은 configure 가변 인자(관문 아래 실행).

## Branches and early returns

- Exact AST return nodes: `107:3`, `109:2`.

| Branch | AST kind | Source location | Condition |
|---|---|---|---|
| B1 | range | 73:2 | `for _, symbol := range symbols {` |
| B2 | if | 75:3 | `if err != nil {` |
| B3 | if | 81:3 | `if err != nil {` |
| B4 | range | 98:3 | `for _, target := range targets {` |
| B5 | if | 102:4 | `if err != nil {` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `t.Helper` | 69:2 |
| `familyScoresForTest` | 70:12 |
| `arbitrationLineageDigests` | 71:34 |
| `make` | 72:13 |
| `len` | 72:52 |
| `strategyrouter.NewOwnerKey` | 74:15 |
| `t.Fatal` | 76:4 |
| `strategyrouter.MultiCandidateRouteFixture` | 78:19 |
| `t.Fatal` | 82:4 |
| `append` | 84:13 |
| `strategy.ApprovedSnapshotForTest` | 85:14 |
| `strategyrouter.ProductionRouteAuthorityFromRequestForTest` | 86:14 |
| `len` | 91:18 |
| `testStrategyProposalLoader` | 95:12 |
| `make` | 97:13 |
| `len` | 97:52 |
| `strategyflow.AcceptedResultForAuthorityTest` | 99:19 |
| `riskLoaderDescriptor` | 99:63 |
| `target.Approved.Symbol` | 100:24 |
| `now.Add` | 101:5 |
| `now.Add` | 101:28 |
| `t.Fatal` | 103:5 |
| `target.Approved.Symbol` | 105:11 |
| `strategyproposal.ProductionBatchAuthorityMultiLaneForTest` | 107:10 |
| `loader.collect` | 109:9 |
| `context.Background` | 109:24 |
| `routeReadySchedulePair` | 109:46 |
| `proposalFXPair` | 109:83 |

## Safety conclusion

- 시험 코드 — 생산 경로 없음.
