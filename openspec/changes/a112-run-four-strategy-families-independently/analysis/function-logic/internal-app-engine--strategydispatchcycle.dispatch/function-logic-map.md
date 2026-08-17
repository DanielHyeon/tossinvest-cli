# Function Logic Map: `dispatch`

- Source: `internal/app/engine/strategy_dispatch_cycle.go`
- Current-base source SHA-256: `0ce70d7b683d586d4224440b2fe66df7e018caacdb20b7c5ae1f46e7ad98d7b1`
- Signature: `strategyDispatchCycle.dispatch(params=2, results=2)`
- Source range: `53:1`–`138:2`
- AST evidence: `ast.json`, generated from frozen base `016da6245feb60e13971388be386c2c2041469a8`.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- Inputs/results are the exact AST signature above; this L0 map does not infer undocumented state.
- Any later edit must preserve OFF defaults, the owner key without family/horizon, and zero exposure-raising dispatch while a prerequisite is missing.

## Branches and early returns

- Exact AST return nodes: `56:3, 59:3, 65:3, 71:3, 75:3, 79:3, 83:3, 87:3, 91:3, 98:3, 104:3, 117:3, 123:3, 127:3, 129:2, 136:4`.

| Branch | AST kind | Source location | Required test disposition |
|---|---|---|---|
| B1 | if | 55:2 | planned targeted RED before any edit; not run by L0 |
| B2 | if | 58:2 | planned targeted RED before any edit; not run by L0 |
| B3 | if | 63:2 | planned targeted RED before any edit; not run by L0 |
| B4 | if | 70:2 | planned targeted RED before any edit; not run by L0 |
| B5 | if | 74:2 | planned targeted RED before any edit; not run by L0 |
| B6 | if | 78:2 | planned targeted RED before any edit; not run by L0 |
| B7 | if | 82:2 | planned targeted RED before any edit; not run by L0 |
| B8 | if | 86:2 | planned targeted RED before any edit; not run by L0 |
| B9 | if | 90:2 | planned targeted RED before any edit; not run by L0 |
| B10 | if | 97:2 | planned targeted RED before any edit; not run by L0 |
| B11 | if | 103:2 | planned targeted RED before any edit; not run by L0 |
| B12 | if | 116:2 | planned targeted RED before any edit; not run by L0 |
| B13 | if | 122:2 | planned targeted RED before any edit; not run by L0 |
| B14 | if | 126:2 | planned targeted RED before any edit; not run by L0 |

## Calls and live bindings

| Callee expression | Source location | Current-base evidence/requirement |
|---|---|---|
| validateStrategyFirstLegResult | 54:23 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| errors.New | 56:28 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| errors.New | 59:28 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| StrategyMarket | 61:12 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| cycle.schedule.forMarket | 62:18 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| cycle.fx.forMarket | 62:52 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| errors.New | 65:28 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| strategyFirstLegPlaceIntent | 70:15 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| cycle.gateway.ObserveStrategyProtection | 73:21 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| strings.ToLower | 73:66 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| string | 73:82 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| cycle.gateway.ObserveStrategyEntryGate | 77:25 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| strings.ToLower | 77:69 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| string | 77:85 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| cycle.dispatchOwner | 81:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| cycle.firstLeg.admit | 85:14 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 87:28 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| cycle.journal.LookupDecision | 89:19 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| errors.New | 91:28 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| uint64 | 95:24 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| bundle.Generation | 96:20 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| cycle.risk.forMarket | 96:20 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| errors.New | 98:28 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| schedule.restore.Activation.Generation | 100:26 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| schedule.restore.Activation.ExpiresAt | 101:25 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| activationExpiresAt.IsZero | 103:34 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| now.IsZero | 103:66 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| now.Before | 103:83 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| errors.New | 104:28 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| journal.StrategyDispatchMarket | 106:63 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| protection.Generation | 109:25 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| strconv.FormatUint | 109:68 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| protection.Generation | 109:87 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| protection.Digest | 109:135 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| reconciliation.Generation | 110:29 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| reconciliation.Digest | 110:80 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| strategyRuntimeBuildDigest | 111:94 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| min | 112:9 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| activationExpiresAt.Sub | 112:29 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| cycle.journal.IssueVerifiedFirstLegStrategyDispatchLease | 113:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| cycle.journal.ClaimStrategyDispatchLease | 119:18 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| strategyFirstLegPlaceIntent | 125:17 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| cycle.gateway.PlaceClaimedStrategy | 129:9 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| cycle.revalidateSchedule | 136:11 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |

## State mutations and fallbacks

- The AST is the exhaustive current-base record of assignments, calls, branches, defers and returns. Before a function body edit, the owning lot must update this map with changed condition semantics and concrete RED/GREEN test evidence.

## Safety conclusion

- L0 status: pre-edit evidence only; no production function was edited and no branch test is claimed as run by L0.
- A named targeted RED or explicit evidence-backed not-applicable rationale is required for every edited branch before GREEN.
