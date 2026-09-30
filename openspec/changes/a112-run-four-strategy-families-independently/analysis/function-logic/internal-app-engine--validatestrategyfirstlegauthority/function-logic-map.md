# Function Logic Map: `validateStrategyFirstLegAuthority`

- Source: `internal/app/engine/strategy_first_leg_admission.go`
- Current-base source SHA-256: `c31f12fd07855ab32d29c815c8b7b21e14c83add0cd3e1c46503bf01c15eda22`
- Signature: `validateStrategyFirstLegAuthority(params=2, results=1)`
- Source range: `146:1`–`181:2`
- AST evidence: `ast.json`, generated from frozen base `016da6245feb60e13971388be386c2c2041469a8`.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- Inputs/results are the exact AST signature above; this L0 map does not infer undocumented state.
- Any later edit must preserve OFF defaults, the owner key without family/horizon, and zero exposure-raising dispatch while a prerequisite is missing.

## Branches and early returns

- Exact AST return nodes: `152:3, 159:3, 168:3, 174:3, 178:3, 180:2`.

| Branch | AST kind | Source location | Required test disposition |
|---|---|---|---|
| B1 | if | 148:2 | planned targeted RED before any edit; not run by L0 |
| B2 | if | 158:2 | planned targeted RED before any edit; not run by L0 |
| B3 | if | 162:2 | planned targeted RED before any edit; not run by L0 |
| B4 | if | 170:2 | planned targeted RED before any edit; not run by L0 |
| B5 | if | 176:2 | planned targeted RED before any edit; not run by L0 |

## Calls and live bindings

| Callee expression | Source location | Current-base evidence/requirement |
|---|---|---|
| validateStrategyFirstLegResult | 147:32 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| request.Result.ExecutionTerms.Identity | 150:3 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| accepted.result.ExecutionTerms.Identity | 150:47 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 152:10 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| MajorDecimal | 155:25 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| terms.Entry | 155:25 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| MajorDecimal | 156:23 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| terms.EffectiveStop | 156:23 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| MajorDecimal | 157:27 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| terms.Target | 157:27 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 159:10 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| string | 162:21 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| string | 166:3 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| string | 166:31 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 168:10 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| uint64 | 172:33 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 174:10 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| strings.TrimSpace | 176:53 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| strings.TrimSpace | 177:81 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 178:10 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |

## State mutations and fallbacks

- The AST is the exhaustive current-base record of assignments, calls, branches, defers and returns. Before a function body edit, the owning lot must update this map with changed condition semantics and concrete RED/GREEN test evidence.

## Safety conclusion

- L0 status: pre-edit evidence only; no production function was edited and no branch test is claimed as run by L0.
- A named targeted RED or explicit evidence-backed not-applicable rationale is required for every edited branch before GREEN.

a112 5.2.2.2: 같은 파일의 다른 함수 편집으로 줄만 이동(본문 불변)
