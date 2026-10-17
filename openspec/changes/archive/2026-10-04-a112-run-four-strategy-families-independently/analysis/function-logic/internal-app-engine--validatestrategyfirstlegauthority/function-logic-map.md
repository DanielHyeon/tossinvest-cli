# Function Logic Map: `validateStrategyFirstLegAuthority`

- Source: `internal/app/engine/strategy_first_leg_admission.go`
- Current-base source SHA-256: `254b4a6abb0d95febd036b0f437829c391e61000fa71b9b9abac1241ee14444c`
- Signature: `validateStrategyFirstLegAuthority(params=2, results=1)`
- Source range: `155:1`–`190:2`
- AST evidence: `ast.json`, generated from frozen base `016da6245feb60e13971388be386c2c2041469a8`.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- Inputs/results are the exact AST signature above; this L0 map does not infer undocumented state.
- Any later edit must preserve OFF defaults, the owner key without family/horizon, and zero exposure-raising dispatch while a prerequisite is missing.

## Branches and early returns

- Exact AST return nodes: `161:3, 168:3, 177:3, 183:3, 187:3, 189:2`.

| Branch | AST kind | Source location | Required test disposition |
|---|---|---|---|
| B1 | if | 157:2 | planned targeted RED before any edit; not run by L0 |
| B2 | if | 167:2 | planned targeted RED before any edit; not run by L0 |
| B3 | if | 171:2 | planned targeted RED before any edit; not run by L0 |
| B4 | if | 179:2 | planned targeted RED before any edit; not run by L0 |
| B5 | if | 185:2 | planned targeted RED before any edit; not run by L0 |

## Calls and live bindings

| Callee expression | Source location | Current-base evidence/requirement |
|---|---|---|
| validateStrategyFirstLegResult | 156:32 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| request.Result.ExecutionTerms.Identity | 159:3 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| accepted.result.ExecutionTerms.Identity | 159:47 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 161:10 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| MajorDecimal | 164:25 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| terms.Entry | 164:25 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| MajorDecimal | 165:23 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| terms.EffectiveStop | 165:23 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| MajorDecimal | 166:27 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| terms.Target | 166:27 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 168:10 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| string | 171:21 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| string | 175:3 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| string | 175:31 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 177:10 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| uint64 | 181:33 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 183:10 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| strings.TrimSpace | 185:53 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| strings.TrimSpace | 186:81 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 187:10 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |

## State mutations and fallbacks

- The AST is the exhaustive current-base record of assignments, calls, branches, defers and returns. Before a function body edit, the owning lot must update this map with changed condition semantics and concrete RED/GREEN test evidence.

## Safety conclusion

- L0 status: pre-edit evidence only; no production function was edited and no branch test is claimed as run by L0.
- A named targeted RED or explicit evidence-backed not-applicable rationale is required for every edited branch before GREEN.

a112 5.2.2.2: 같은 파일의 다른 함수 편집으로 줄만 이동(본문 불변)

a112 5.2.2.2 리뷰 수리: 같은 파일의 다른 함수 편집으로 줄만 이동(본문 불변)

a112 6.2: 같은 파일의 다른 함수 편집으로 줄만 이동(본문 불변)
