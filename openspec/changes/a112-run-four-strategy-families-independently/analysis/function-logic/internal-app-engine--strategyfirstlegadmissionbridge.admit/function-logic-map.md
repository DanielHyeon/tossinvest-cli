# Function Logic Map: `admit`

- Source: `internal/app/engine/strategy_first_leg_admission.go`
- Current-base source SHA-256: `a08618229629b30fd7f4f45b19b3773cb9b1e84f9dc3eebf6654e44ea4e72894`
- Signature: `strategyFirstLegAdmissionBridge.admit(params=2, results=1)`
- Source range: `69:1`–`93:2`
- AST evidence: `ast.json`, generated from frozen base `016da6245feb60e13971388be386c2c2041469a8`.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- Inputs/results are the exact AST signature above; this L0 map does not infer undocumented state.
- Any later edit must preserve OFF defaults, the owner key without family/horizon, and zero exposure-raising dispatch while a prerequisite is missing.

## Branches and early returns

- Exact AST return nodes: `72:3, 75:3, 79:3, 82:3, 86:3, 90:3, 92:2`.

| Branch | AST kind | Source location | Required test disposition |
|---|---|---|---|
| B1 | if | 71:2 | planned targeted RED before any edit; not run by L0 |
| B2 | if | 74:2 | planned targeted RED before any edit; not run by L0 |
| B3 | if | 78:2 | planned targeted RED before any edit; not run by L0 |
| B4 | if | 81:2 | planned targeted RED before any edit; not run by L0 |
| B5 | if | 85:2 | planned targeted RED before any edit; not run by L0 |
| B6 | if | 89:2 | planned targeted RED before any edit; not run by L0 |

## Calls and live bindings

| Callee expression | Source location | Current-base evidence/requirement |
|---|---|---|
| validateStrategyFirstLegResult | 70:23 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| strategyFirstLegRefusal | 75:10 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| b.loader.collectStrategyFirstLegAuthority | 77:20 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| strategyFirstLegRefusal | 79:10 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| err.Error | 79:94 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| validateStrategyFirstLegAuthority | 81:12 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| strategyFirstLegRefusal | 82:10 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| err.Error | 82:86 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| b.guardian.PrecheckQFinalCampaignFirstLeg | 84:19 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| strategyFirstLegRefusal | 86:10 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| err.Error | 86:86 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| b.guardian.IssuePrecheckedQFinalCampaignFirstLeg | 88:18 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| strategyFirstLegRefusal | 90:10 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| err.Error | 90:90 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| string | 92:81 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |

## State mutations and fallbacks

- The AST is the exhaustive current-base record of assignments, calls, branches, defers and returns. Before a function body edit, the owning lot must update this map with changed condition semantics and concrete RED/GREEN test evidence.

## Safety conclusion

- L0 status: pre-edit evidence only; no production function was edited and no branch test is claimed as run by L0.
- A named targeted RED or explicit evidence-backed not-applicable rationale is required for every edited branch before GREEN.
