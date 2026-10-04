# Function Logic Map: `TestAClosedMarketStillCarriesTheGatesActivation (시험)`

- Source: `internal/app/engine/a112_family_gate_test.go`
- Source SHA-256: `49511cdebc1ba5506737d375b775aa2a533b358701fb531f33f41051f2400fcf`
- Signature: `TestAClosedMarketStillCarriesTheGatesActivation(params=1, results=0)`
- Source range: `260:1`–`295:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 7.3.1 SHADOW).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 시험 코드 — 생산 판정 없음.

## Branches and early returns

- Exact AST return nodes: `274:3, 280:3`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | range | 265:2 | 시험 자신의 갈래 |
| B2 | if | 266:3 | 시험 자신의 갈래 |
| B3 | if | 286:2 | 시험 자신의 갈래 |
| B4 | if | 289:2 | 시험 자신의 갈래 |
| B5 | if | 292:2 | 시험 자신의 갈래 |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `familyGateFixture` | 261:25 |
| `time.Date` | 262:9 |
| `familyScoresForTest` | 264:10 |
| `testStrategyProposalLoader` | 270:12 |
| `arbitrationBatch` | 274:10 |
| `collect` | 282:13 |
| `loader.withStrategyLanes` | 282:13 |
| `context.Background` | 282:55 |
| `routeReadySchedulePair` | 282:77 |
| `arbitrationRoutePair` | 283:3 |
| `proposalFXPair` | 284:75 |
| `pair.forMarket` | 285:15 |
| `t.Fatalf` | 287:3 |
| `Verified` | 289:6 |
| `authority.familyActivation` | 289:6 |
| `t.Fatal` | 290:3 |
| `Generation` | 292:5 |
| `authority.familyActivation` | 292:5 |
| `activation.Generation` | 292:50 |
| `t.Fatalf` | 293:3 |
| `Generation` | 293:34 |
| `authority.familyActivation` | 293:34 |
| `activation.Generation` | 293:77 |

## State mutations and fallbacks

- 시험 fixture.

## Safety conclusion

- 시험 코드 — 실주문 · 원장 쓰기 없음(fixture 원장 · Gateway 스파이).
