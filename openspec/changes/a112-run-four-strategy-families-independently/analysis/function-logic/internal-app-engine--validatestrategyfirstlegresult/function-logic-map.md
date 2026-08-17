# Function Logic Map: `validateStrategyFirstLegResult`

- Source: `internal/app/engine/strategy_first_leg_admission.go`
- Current-base source SHA-256: `a08618229629b30fd7f4f45b19b3773cb9b1e84f9dc3eebf6654e44ea4e72894`
- Signature: `validateStrategyFirstLegResult(params=1, results=2)`
- Source range: `102:1`–`136:2`
- AST evidence: `ast.json`, generated from frozen base `016da6245feb60e13971388be386c2c2041469a8`.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- Inputs/results are the exact AST signature above; this L0 map does not infer undocumented state.
- Any later edit must preserve OFF defaults, the owner key without family/horizon, and zero exposure-raising dispatch while a prerequisite is missing.

## Branches and early returns

- Exact AST return nodes: `112:3, 128:3, 132:4, 135:2`.

| Branch | AST kind | Source location | Required test disposition |
|---|---|---|---|
| B1 | switch | 106:2 | planned targeted RED before any edit; not run by L0 |
| B2 | case | 107:2 | planned targeted RED before any edit; not run by L0 |
| B3 | case | 109:2 | planned targeted RED before any edit; not run by L0 |
| B4 | case | 111:2 | planned targeted RED before any edit; not run by L0 |
| B5 | if | 127:2 | planned targeted RED before any edit; not run by L0 |
| B6 | range | 130:2 | planned targeted RED before any edit; not run by L0 |
| B7 | if | 131:3 | planned targeted RED before any edit; not run by L0 |

## Calls and live bindings

| Callee expression | Source location | Current-base evidence/requirement |
|---|---|---|
| strategyFirstLegRefusal | 112:38 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| lineage.Valid | 115:102 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| terms.Valid | 115:122 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| strings.ToUpper | 116:73 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| strings.TrimSpace | 116:89 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| terms.AccountRef | 123:3 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| terms.Market | 123:47 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| terms.Symbol | 123:83 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| terms.CampaignID | 124:3 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| terms.LegOrdinal | 124:47 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| terms.Quantity | 124:74 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| terms.LineageIdentity | 125:3 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| Identity | 125:50 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| terms.Policy | 125:50 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| strategyFirstLegRefusal | 128:38 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| terms.Entry | 130:55 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| terms.EffectiveStop | 130:70 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| terms.Target | 130:93 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| price.Currency | 131:6 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| price.MinorScale | 131:47 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| price.UnitVersion | 131:87 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| strategyFirstLegRefusal | 132:39 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |

## State mutations and fallbacks

- The AST is the exhaustive current-base record of assignments, calls, branches, defers and returns. Before a function body edit, the owning lot must update this map with changed condition semantics and concrete RED/GREEN test evidence.

## Safety conclusion

- L0 status: pre-edit evidence only; no production function was edited and no branch test is claimed as run by L0.
- A named targeted RED or explicit evidence-backed not-applicable rationale is required for every edited branch before GREEN.
