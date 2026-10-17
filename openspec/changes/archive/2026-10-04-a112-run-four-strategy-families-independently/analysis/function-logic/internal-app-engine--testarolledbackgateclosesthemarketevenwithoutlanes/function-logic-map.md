# Function Logic Map: `TestARolledBackGateClosesTheMarketEvenWithoutLanes (시험)`

- Source: `internal/app/engine/a112_family_rollback_test.go`
- Source SHA-256: `c1b6888e6bccd1b9ca9e54c43abbae8063f5e9cf12178e98b9170634d4c73979`
- Signature: `TestARolledBackGateClosesTheMarketEvenWithoutLanes(params=1, results=0)`
- Source range: `401:1`–`416:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 7.3.1 SHADOW).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 시험 코드 — 생산 판정 없음.

## Branches and early returns

- Exact AST return nodes: ``.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 410:2 | 시험 자신의 갈래 |
| B2 | if | 413:2 | 시험 자신의 갈래 |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `time.Date` | 402:9 |
| `testStrategyProposalLoader` | 403:11 |
| `rollbackLoader` | 404:12 |
| `strings.Repeat` | 405:60 |
| `forMarket` | 408:15 |
| `a112PairOnly` | 408:15 |
| `loader.collect` | 408:28 |
| `context.Background` | 408:43 |
| `routeReadySchedulePair` | 408:65 |
| `rollbackRoutes` | 408:94 |
| `proposalFXPair` | 409:3 |
| `t.Fatalf` | 411:3 |
| `strings.Join` | 413:5 |
| `string` | 413:60 |
| `t.Fatalf` | 414:3 |

## State mutations and fallbacks

- 시험 fixture.

## Safety conclusion

- 시험 코드 — 실주문 · 원장 쓰기 없음(fixture 원장 · Gateway 스파이).
