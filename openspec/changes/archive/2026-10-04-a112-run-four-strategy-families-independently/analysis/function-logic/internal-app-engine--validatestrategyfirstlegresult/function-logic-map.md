# Function Logic Map: `validateStrategyFirstLegResult`

- Source: `internal/app/engine/strategy_first_leg_admission.go`
- Current-base source SHA-256: `254b4a6abb0d95febd036b0f437829c391e61000fa71b9b9abac1241ee14444c`
- Signature: `validateStrategyFirstLegResult(params=1, results=2)`
- Source range: `119:1`–`153:2`
- AST evidence: `ast.json`, generated from frozen base `016da6245feb60e13971388be386c2c2041469a8`.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- Inputs/results are the exact AST signature above; this L0 map does not infer undocumented state.
- Any later edit must preserve OFF defaults, the owner key without family/horizon, and zero exposure-raising dispatch while a prerequisite is missing.

## Branches and early returns

- Exact AST return nodes: `129:3, 145:3, 149:4, 152:2`.

| Branch | AST kind | Source location | Required test disposition |
|---|---|---|---|
| B1 | switch | 123:2 | planned targeted RED before any edit; not run by L0 |
| B2 | case | 124:2 | planned targeted RED before any edit; not run by L0 |
| B3 | case | 126:2 | planned targeted RED before any edit; not run by L0 |
| B4 | case | 128:2 | planned targeted RED before any edit; not run by L0 |
| B5 | if | 144:2 | planned targeted RED before any edit; not run by L0 |
| B6 | range | 147:2 | planned targeted RED before any edit; not run by L0 |
| B7 | if | 148:3 | planned targeted RED before any edit; not run by L0 |

## Calls and live bindings

| Callee expression | Source location | Current-base evidence/requirement |
|---|---|---|
| strategyFirstLegRefusal | 129:38 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| lineage.Valid | 132:102 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| terms.Valid | 132:122 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| strings.ToUpper | 133:73 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| strings.TrimSpace | 133:89 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| terms.AccountRef | 140:3 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| terms.Market | 140:47 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| terms.Symbol | 140:83 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| terms.CampaignID | 141:3 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| terms.LegOrdinal | 141:47 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| terms.Quantity | 141:74 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| terms.LineageIdentity | 142:3 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| Identity | 142:50 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| terms.Policy | 142:50 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| strategyFirstLegRefusal | 145:38 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| terms.Entry | 147:55 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| terms.EffectiveStop | 147:70 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| terms.Target | 147:93 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| price.Currency | 148:6 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| price.MinorScale | 148:47 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| price.UnitVersion | 148:87 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| strategyFirstLegRefusal | 149:39 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |

## State mutations and fallbacks

- The AST is the exhaustive current-base record of assignments, calls, branches, defers and returns. Before a function body edit, the owning lot must update this map with changed condition semantics and concrete RED/GREEN test evidence.

## Safety conclusion

- L0 status: pre-edit evidence only; no production function was edited and no branch test is claimed as run by L0.
- A named targeted RED or explicit evidence-backed not-applicable rationale is required for every edited branch before GREEN.

a112 5.2.2.2: 같은 파일의 다른 함수 편집으로 줄만 이동(본문 불변)

a112 5.2.2.2 리뷰 수리: 같은 파일의 다른 함수 편집으로 줄만 이동(본문 불변)

a112 6.2: 같은 파일의 다른 함수 편집으로 줄만 이동(본문 불변)
