# Function Logic Map: `validateStrategyFirstLegResult`

- Source: `internal/app/engine/strategy_first_leg_admission.go`
- Current-base source SHA-256: `3c82793300b39454ccac2ff41fe97c0b76ed2190534c5a76c0f9f7abd59652e5`
- Signature: `validateStrategyFirstLegResult(params=1, results=2)`
- Source range: `107:1`–`141:2`
- AST evidence: `ast.json`, generated from frozen base `016da6245feb60e13971388be386c2c2041469a8`.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- Inputs/results are the exact AST signature above; this L0 map does not infer undocumented state.
- Any later edit must preserve OFF defaults, the owner key without family/horizon, and zero exposure-raising dispatch while a prerequisite is missing.

## Branches and early returns

- Exact AST return nodes: `117:3, 133:3, 137:4, 140:2`.

| Branch | AST kind | Source location | Required test disposition |
|---|---|---|---|
| B1 | switch | 111:2 | planned targeted RED before any edit; not run by L0 |
| B2 | case | 112:2 | planned targeted RED before any edit; not run by L0 |
| B3 | case | 114:2 | planned targeted RED before any edit; not run by L0 |
| B4 | case | 116:2 | planned targeted RED before any edit; not run by L0 |
| B5 | if | 132:2 | planned targeted RED before any edit; not run by L0 |
| B6 | range | 135:2 | planned targeted RED before any edit; not run by L0 |
| B7 | if | 136:3 | planned targeted RED before any edit; not run by L0 |

## Calls and live bindings

| Callee expression | Source location | Current-base evidence/requirement |
|---|---|---|
| strategyFirstLegRefusal | 117:38 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| lineage.Valid | 120:102 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| terms.Valid | 120:122 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| strings.ToUpper | 121:73 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| strings.TrimSpace | 121:89 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| terms.AccountRef | 128:3 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| terms.Market | 128:47 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| terms.Symbol | 128:83 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| terms.CampaignID | 129:3 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| terms.LegOrdinal | 129:47 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| terms.Quantity | 129:74 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| terms.LineageIdentity | 130:3 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| Identity | 130:50 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| terms.Policy | 130:50 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| strategyFirstLegRefusal | 133:38 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| terms.Entry | 135:55 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| terms.EffectiveStop | 135:70 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| terms.Target | 135:93 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| price.Currency | 136:6 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| price.MinorScale | 136:47 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| price.UnitVersion | 136:87 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| strategyFirstLegRefusal | 137:39 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |

## State mutations and fallbacks

- The AST is the exhaustive current-base record of assignments, calls, branches, defers and returns. Before a function body edit, the owning lot must update this map with changed condition semantics and concrete RED/GREEN test evidence.

## Safety conclusion

- L0 status: pre-edit evidence only; no production function was edited and no branch test is claimed as run by L0.
- A named targeted RED or explicit evidence-backed not-applicable rationale is required for every edited branch before GREEN.

a112 5.2.2.2: 같은 파일의 다른 함수 편집으로 줄만 이동(본문 불변)

a112 5.2.2.2 리뷰 수리: 같은 파일의 다른 함수 편집으로 줄만 이동(본문 불변)
