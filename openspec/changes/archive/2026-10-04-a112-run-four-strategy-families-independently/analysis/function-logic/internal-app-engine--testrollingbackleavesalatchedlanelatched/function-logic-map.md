# Function Logic Map: `TestRollingBackLeavesALatchedLaneLatched (시험)`

- Source: `internal/app/engine/a112_family_rollback_test.go`
- Source SHA-256: `c1b6888e6bccd1b9ca9e54c43abbae8063f5e9cf12178e98b9170634d4c73979`
- Signature: `TestRollingBackLeavesALatchedLaneLatched(params=1, results=0)`
- Source range: `231:1`–`268:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 7.3.1 SHADOW).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 시험 코드 — 생산 판정 없음.

## Branches and early returns

- Exact AST return nodes: ``.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | range | 234:2 | 시험 자신의 갈래 |
| B2 | if | 235:3 | 시험 자신의 갈래 |
| B3 | if | 238:3 | 시험 자신의 갈래 |
| B4 | if | 249:2 | 시험 자신의 갈래 |
| B5 | if | 254:2 | 시험 자신의 갈래 |
| B6 | range | 257:2 | 시험 자신의 갈래 |
| B7 | if | 258:3 | 시험 자신의 갈래 |
| B8 | if | 265:2 | 시험 자신의 갈래 |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `time.Date` | 232:9 |
| `familyGateFixture` | 233:18 |
| `runtime.lanesFor` | 234:23 |
| `lane.Key` | 235:6 |
| `lane.Fail` | 238:19 |
| `t.Fatal` | 239:4 |
| `testStrategyProposalLoader` | 242:11 |
| `rollbackLoader` | 243:12 |
| `strings.Repeat` | 244:60 |
| `forMarket` | 247:12 |
| `a112PairOnly` | 247:12 |
| `loader.collect` | 247:25 |
| `context.Background` | 247:40 |
| `routeReadySchedulePair` | 247:62 |
| `rollbackRoutes` | 247:91 |
| `proposalFXPair` | 248:3 |
| `t.Fatalf` | 250:3 |
| `append` | 252:11 |
| `(unnamed)` | 252:18 |
| `sort.Strings` | 253:2 |
| `strings.Join` | 254:5 |
| `string` | 254:33 |
| `string` | 254:75 |
| `t.Fatalf` | 255:3 |
| `runtime.lanesFor` | 257:23 |
| `lane.Key` | 258:6 |
| `lane.Health` | 258:60 |
| `t.Fatalf` | 259:4 |
| `lane.Health` | 259:71 |
| `collectUnderGate` | 263:15 |
| `selectedFamily` | 265:12 |
| `t.Fatalf` | 266:3 |

## State mutations and fallbacks

- 시험 fixture.

## Safety conclusion

- 시험 코드 — 실주문 · 원장 쓰기 없음(fixture 원장 · Gateway 스파이).
