# Function Logic Map: `dispatch`

- Source: `internal/app/engine/strategy_dispatch_cycle.go`
- Current-base source SHA-256: `0ce70d7b683d586d4224440b2fe66df7e018caacdb20b7c5ae1f46e7ad98d7b1`
- Signature: `strategyDispatchCycle.dispatch(params=2, results=2)`
- Source range: `63:1`–`149:2`
- AST evidence: `ast.json`, generated from frozen base `016da6245feb60e13971388be386c2c2041469a8`.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- Inputs/results are the exact AST signature above; this L0 map does not infer undocumented state.
- Any later edit must preserve OFF defaults, the owner key without family/horizon, and zero exposure-raising dispatch while a prerequisite is missing.

## Branches and early returns

- Exact AST return nodes: `67:3, 70:3, 76:3, 82:3, 86:3, 90:3, 94:3, 98:3, 102:3, 109:3, 115:3, 128:3, 134:3, 138:3, 140:2, 147:4`.

| Branch | AST kind | Source location | Required test disposition |
|---|---|---|---|
| B1 | if | 66:2 | planned targeted RED before any edit; not run by L0 |
| B2 | if | 69:2 | planned targeted RED before any edit; not run by L0 |
| B3 | if | 74:2 | planned targeted RED before any edit; not run by L0 |
| B4 | if | 81:2 | planned targeted RED before any edit; not run by L0 |
| B5 | if | 85:2 | planned targeted RED before any edit; not run by L0 |
| B6 | if | 89:2 | planned targeted RED before any edit; not run by L0 |
| B7 | if | 93:2 | planned targeted RED before any edit; not run by L0 |
| B8 | if | 97:2 | planned targeted RED before any edit; not run by L0 |
| B9 | if | 101:2 | planned targeted RED before any edit; not run by L0 |
| B10 | if | 108:2 | planned targeted RED before any edit; not run by L0 |
| B11 | if | 114:2 | planned targeted RED before any edit; not run by L0 |
| B12 | if | 127:2 | planned targeted RED before any edit; not run by L0 |
| B13 | if | 133:2 | planned targeted RED before any edit; not run by L0 |
| B14 | if | 137:2 | planned targeted RED before any edit; not run by L0 |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `delivered.Result` | 64:12 |
| `validateStrategyFirstLegResult` | 65:23 |
| `errors.New` | 67:28 |
| `errors.New` | 70:28 |
| `StrategyMarket` | 72:12 |
| `cycle.schedule.forMarket` | 73:18 |
| `cycle.fx.forMarket` | 73:52 |
| `errors.New` | 76:28 |
| `strategyFirstLegPlaceIntent` | 81:15 |
| `cycle.gateway.ObserveStrategyProtection` | 84:21 |
| `strings.ToLower` | 84:66 |
| `string` | 84:82 |
| `cycle.gateway.ObserveStrategyEntryGate` | 88:25 |
| `strings.ToLower` | 88:69 |
| `string` | 88:85 |
| `cycle.dispatchOwner` | 92:16 |
| `cycle.firstLeg.admit` | 96:14 |
| `fmt.Errorf` | 98:28 |
| `cycle.journal.LookupDecision` | 100:19 |
| `errors.New` | 102:28 |
| `uint64` | 106:24 |
| `bundle.Generation` | 107:20 |
| `cycle.risk.forMarket` | 107:20 |
| `errors.New` | 109:28 |
| `schedule.restore.Activation.Generation` | 111:26 |
| `schedule.restore.Activation.ExpiresAt` | 112:25 |
| `activationExpiresAt.IsZero` | 114:34 |
| `now.IsZero` | 114:66 |
| `now.Before` | 114:83 |
| `errors.New` | 115:28 |
| `journal.StrategyDispatchMarket` | 117:63 |
| `protection.Generation` | 120:25 |
| `strconv.FormatUint` | 120:68 |
| `protection.Generation` | 120:87 |
| `protection.Digest` | 120:135 |
| `reconciliation.Generation` | 121:29 |
| `reconciliation.Digest` | 121:80 |
| `strategyRuntimeBuildDigest` | 122:94 |
| `min` | 123:9 |
| `activationExpiresAt.Sub` | 123:29 |
| `cycle.journal.IssueVerifiedFirstLegStrategyDispatchLease` | 124:16 |
| `cycle.journal.ClaimStrategyDispatchLease` | 130:18 |
| `strategyFirstLegPlaceIntent` | 136:17 |
| `cycle.gateway.PlaceClaimedStrategy` | 140:9 |
| `cycle.revalidateSchedule` | 147:11 |

## State mutations and fallbacks

- The AST is the exhaustive current-base record of assignments, calls, branches, defers and returns. Before a function body edit, the owning lot must update this map with changed condition semantics and concrete RED/GREEN test evidence.

## Safety conclusion

- L0 status: pre-edit evidence only; no production function was edited and no branch test is claimed as run by L0.
- A named targeted RED or explicit evidence-backed not-applicable rationale is required for every edited branch before GREEN.
