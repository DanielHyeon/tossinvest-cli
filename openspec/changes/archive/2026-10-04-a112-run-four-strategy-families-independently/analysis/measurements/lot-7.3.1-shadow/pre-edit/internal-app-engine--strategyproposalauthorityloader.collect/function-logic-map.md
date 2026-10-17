# Function Logic Map (편집 전): `collect`

- Source: `internal/app/engine/strategy_proposal_authority.go`
- Source SHA-256: `2c546898bc4178d0ee44238d51fb8e05a4c908cee721041ee24713c4c98ac385`
- Signature: `strategyProposalAuthorityLoader.collect(params=4, results=1)`
- Source range: `257:1`–`291:2`
- AST evidence: `ast.json` — 편집 **전**(HEAD 9e5f3ccf).

## Inputs and invariants

- 편집 계획: 브리프 §4: 시장별 goroutine 결과에 shadow 묶음 동반, 짝 구조 옆 strategyShadowPair 로 반환; recover 갈래는 부재 값.

## Branches and early returns

- Exact AST return nodes: `259:3`, `290:2`.

| Branch | AST kind | Source location | Condition |
|---|---|---|---|
| B1 | if | 258:2 | `if loader == nil // ctx == nil // schedule.observedAt.IsZero() // !schedule.observedAt.Equal(routes.observedAt) // !schedule.observedAt.Equal(fx.observedAt) {` |
| B2 | range | 266:2 | `for _, market := range []StrategyMarket{StrategyMarketKR, StrategyMarketUS} {` |
| B3 | if | 272:6 | `if recover() != nil {` |
| B4 | range | 282:2 | `for range 2 {` |
| B5 | if | 284:3 | `if result.market == StrategyMarketKR {` |
| B6 | else | 286:10 | `} else {` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `schedule.observedAt.IsZero` | 258:36 |
| `schedule.observedAt.Equal` | 258:69 |
| `schedule.observedAt.Equal` | 258:118 |
| `failedStrategyProposalPair` | 259:10 |
| `make` | 265:14 |
| `(unnamed)` | 268:6 |
| `(unnamed)` | 270:4 |
| `(unnamed)` | 271:11 |
| `recover` | 272:9 |
| `loader.collectMarket` | 276:13 |
| `schedule.forMarket` | 276:39 |
| `routes.forMarket` | 276:67 |
| `fx.forMarket` | 276:93 |

## Safety conclusion

- High-risk(제안 권한 파도) — authority 짝 · recover · join 무변경; shadow 짝은 별도 반환값.
