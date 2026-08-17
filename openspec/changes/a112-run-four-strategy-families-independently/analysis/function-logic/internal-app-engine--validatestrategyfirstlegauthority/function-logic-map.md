# Function Logic Map: `validateStrategyFirstLegAuthority`

- Source: `internal/app/engine/strategy_first_leg_admission.go`
- Current-base source SHA-256: `a08618229629b30fd7f4f45b19b3773cb9b1e84f9dc3eebf6654e44ea4e72894`
- Signature: `validateStrategyFirstLegAuthority(params=2, results=1)`
- Source range: `138:1`–`173:2`
- AST evidence: `ast.json`, generated from frozen base `016da6245feb60e13971388be386c2c2041469a8`.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- Inputs/results are the exact AST signature above; this L0 map does not infer undocumented state.
- Any later edit must preserve OFF defaults, the owner key without family/horizon, and zero exposure-raising dispatch while a prerequisite is missing.

## Branches and early returns

- Exact AST return nodes: `144:3, 151:3, 160:3, 166:3, 170:3, 172:2`.

| Branch | AST kind | Source location | Required test disposition |
|---|---|---|---|
| B1 | if | 140:2 | planned targeted RED before any edit; not run by L0 |
| B2 | if | 150:2 | planned targeted RED before any edit; not run by L0 |
| B3 | if | 154:2 | planned targeted RED before any edit; not run by L0 |
| B4 | if | 162:2 | planned targeted RED before any edit; not run by L0 |
| B5 | if | 168:2 | planned targeted RED before any edit; not run by L0 |

## Calls and live bindings

| Callee expression | Source location | Current-base evidence/requirement |
|---|---|---|
| validateStrategyFirstLegResult | 139:32 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| request.Result.ExecutionTerms.Identity | 142:3 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| accepted.result.ExecutionTerms.Identity | 142:47 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 144:10 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| MajorDecimal | 147:25 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| terms.Entry | 147:25 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| MajorDecimal | 148:23 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| terms.EffectiveStop | 148:23 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| MajorDecimal | 149:27 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| terms.Target | 149:27 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 151:10 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| string | 154:21 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| string | 158:3 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| string | 158:31 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 160:10 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| uint64 | 164:33 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 166:10 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| strings.TrimSpace | 168:53 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| strings.TrimSpace | 169:81 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 170:10 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |

## State mutations and fallbacks

- The AST is the exhaustive current-base record of assignments, calls, branches, defers and returns. Before a function body edit, the owning lot must update this map with changed condition semantics and concrete RED/GREEN test evidence.

## Safety conclusion

- L0 status: pre-edit evidence only; no production function was edited and no branch test is claimed as run by L0.
- A named targeted RED or explicit evidence-backed not-applicable rationale is required for every edited branch before GREEN.
