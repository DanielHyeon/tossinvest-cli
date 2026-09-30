# Function Logic Map: `validateStrategyFirstLegResult`

- Source: `internal/app/engine/strategy_first_leg_admission.go`
- Current-base source SHA-256: `c31f12fd07855ab32d29c815c8b7b21e14c83add0cd3e1c46503bf01c15eda22`
- Signature: `validateStrategyFirstLegResult(params=1, results=2)`
- Source range: `110:1`–`144:2`
- AST evidence: `ast.json`, generated from frozen base `016da6245feb60e13971388be386c2c2041469a8`.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- Inputs/results are the exact AST signature above; this L0 map does not infer undocumented state.
- Any later edit must preserve OFF defaults, the owner key without family/horizon, and zero exposure-raising dispatch while a prerequisite is missing.

## Branches and early returns

- Exact AST return nodes: `120:3, 136:3, 140:4, 143:2`.

| Branch | AST kind | Source location | Required test disposition |
|---|---|---|---|
| B1 | switch | 114:2 | planned targeted RED before any edit; not run by L0 |
| B2 | case | 115:2 | planned targeted RED before any edit; not run by L0 |
| B3 | case | 117:2 | planned targeted RED before any edit; not run by L0 |
| B4 | case | 119:2 | planned targeted RED before any edit; not run by L0 |
| B5 | if | 135:2 | planned targeted RED before any edit; not run by L0 |
| B6 | range | 138:2 | planned targeted RED before any edit; not run by L0 |
| B7 | if | 139:3 | planned targeted RED before any edit; not run by L0 |

## Calls and live bindings

| Callee expression | Source location | Current-base evidence/requirement |
|---|---|---|
| strategyFirstLegRefusal | 120:38 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| lineage.Valid | 123:102 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| terms.Valid | 123:122 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| strings.ToUpper | 124:73 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| strings.TrimSpace | 124:89 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| terms.AccountRef | 131:3 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| terms.Market | 131:47 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| terms.Symbol | 131:83 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| terms.CampaignID | 132:3 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| terms.LegOrdinal | 132:47 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| terms.Quantity | 132:74 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| terms.LineageIdentity | 133:3 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| Identity | 133:50 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| terms.Policy | 133:50 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| strategyFirstLegRefusal | 136:38 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| terms.Entry | 138:55 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| terms.EffectiveStop | 138:70 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| terms.Target | 138:93 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| price.Currency | 139:6 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| price.MinorScale | 139:47 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| price.UnitVersion | 139:87 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| strategyFirstLegRefusal | 140:39 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |

## State mutations and fallbacks

- The AST is the exhaustive current-base record of assignments, calls, branches, defers and returns. Before a function body edit, the owning lot must update this map with changed condition semantics and concrete RED/GREEN test evidence.

## Safety conclusion

- L0 status: pre-edit evidence only; no production function was edited and no branch test is claimed as run by L0.
- A named targeted RED or explicit evidence-backed not-applicable rationale is required for every edited branch before GREEN.

a112 5.2.2.2: 같은 파일의 다른 함수 편집으로 줄만 이동(본문 불변)
