# Function Logic Map: `collectOverflowing (시험 도우미)`

- Source: `internal/app/engine/a112_coordinator_test.go`
- Source SHA-256: `60d5adbcf550c3ff422046d010bfe037481c7fa39e5b1561afb5c1d19a1f2115`
- Signature: `collectOverflowing(params=4, results=1)`
- Source range: `85:1`–`130:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 repin-1e25b3a3).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- KR 에 종목을 원하는 수만큼 두고 각 종목이 지속형 한 레인으로만 제안하게 하는 시험 도우미 — configure 는 수집 직전에만 적재기를 바꾼다.

## Branches and early returns

- Exact AST return nodes: `124:3, 129:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | range | 90:2 | 종목마다 경로 항목 생성 |
| B2 | if | 92:3 | 소유자 열쇠 생성 실패 → Fatal |
| B3 | if | 98:3 | 후보 경로 픽스처 실패 → Fatal |
| B4 | range | 115:3 | 제안 적재 스텁의 대상 순회 |
| B5 | if | 119:4 | 수락 결과 픽스처 실패 → Fatal |
| B6 | range | 126:2 | **(새)** configure 적용 |

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
| `loader.collect` | 129:9 |
| `context.Background` | 129:24 |
| `routeReadySchedulePair` | 129:46 |
| `proposalFXPair` | 129:83 |

## State mutations and fallbacks

- 시험 코드.

## Safety conclusion

- 시험 코드 — 생산 경로 없음.
