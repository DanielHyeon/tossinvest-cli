# Function Logic Map: `validateStrategyFirstLegAuthority`

- Source: `internal/app/engine/strategy_first_leg_admission.go`
- Current-base source SHA-256: `3c82793300b39454ccac2ff41fe97c0b76ed2190534c5a76c0f9f7abd59652e5`
- Signature: `validateStrategyFirstLegAuthority(params=2, results=1)`
- Source range: `143:1`–`178:2`
- AST evidence: `ast.json`, generated from frozen base `016da6245feb60e13971388be386c2c2041469a8`.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- Inputs/results are the exact AST signature above; this L0 map does not infer undocumented state.
- Any later edit must preserve OFF defaults, the owner key without family/horizon, and zero exposure-raising dispatch while a prerequisite is missing.

## Branches and early returns

- Exact AST return nodes: `149:3, 156:3, 165:3, 171:3, 175:3, 177:2`.

| Branch | AST kind | Source location | Required test disposition |
|---|---|---|---|
| B1 | if | 145:2 | planned targeted RED before any edit; not run by L0 |
| B2 | if | 155:2 | planned targeted RED before any edit; not run by L0 |
| B3 | if | 159:2 | planned targeted RED before any edit; not run by L0 |
| B4 | if | 167:2 | planned targeted RED before any edit; not run by L0 |
| B5 | if | 173:2 | planned targeted RED before any edit; not run by L0 |

## Calls and live bindings

| Callee expression | Source location | Current-base evidence/requirement |
|---|---|---|
| validateStrategyFirstLegResult | 144:32 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| request.Result.ExecutionTerms.Identity | 147:3 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| accepted.result.ExecutionTerms.Identity | 147:47 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 149:10 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| MajorDecimal | 152:25 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| terms.Entry | 152:25 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| MajorDecimal | 153:23 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| terms.EffectiveStop | 153:23 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| MajorDecimal | 154:27 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| terms.Target | 154:27 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 156:10 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| string | 159:21 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| string | 163:3 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| string | 163:31 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 165:10 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| uint64 | 169:33 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 171:10 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| strings.TrimSpace | 173:53 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| strings.TrimSpace | 174:81 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 175:10 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |

## State mutations and fallbacks

- The AST is the exhaustive current-base record of assignments, calls, branches, defers and returns. Before a function body edit, the owning lot must update this map with changed condition semantics and concrete RED/GREEN test evidence.

## Safety conclusion

- L0 status: pre-edit evidence only; no production function was edited and no branch test is claimed as run by L0.
- A named targeted RED or explicit evidence-backed not-applicable rationale is required for every edited branch before GREEN.

a112 5.2.2.2: 같은 파일의 다른 함수 편집으로 줄만 이동(본문 불변)

a112 5.2.2.2 리뷰 수리: 같은 파일의 다른 함수 편집으로 줄만 이동(본문 불변)
