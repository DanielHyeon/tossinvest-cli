# Function Logic Map: `collectStrategyFirstLegAuthority`

- Source: `internal/app/engine/strategy_account_first_leg_authority.go`
- Current-base source SHA-256: `1d710710098c03669719779609db137d0660a361a025d070128e8993772ed063`
- Signature: `productionStrategyFirstLegAuthorityLoader.collectStrategyFirstLegAuthority(params=2, results=2)`
- Source range: `210:1`–`281:2`
- AST evidence: `ast.json`, generated from frozen base `016da6245feb60e13971388be386c2c2041469a8`.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- Inputs/results are the exact AST signature above; this L0 map does not infer undocumented state.
- Any later edit must preserve OFF defaults, the owner key without family/horizon, and zero exposure-raising dispatch while a prerequisite is missing.

## Branches and early returns

- Exact AST return nodes: `212:3, 219:3, 224:3, 229:3, 234:3, 251:3, 260:4, 264:4, 266:3, 270:2`.

| Branch | AST kind | Source location | Required test disposition |
|---|---|---|---|
| B1 | if | 211:2 | planned targeted RED before any edit; not run by L0 |
| B2 | if | 217:2 | planned targeted RED before any edit; not run by L0 |
| B3 | if | 223:2 | planned targeted RED before any edit; not run by L0 |
| B4 | if | 227:2 | planned targeted RED before any edit; not run by L0 |
| B5 | if | 232:2 | planned targeted RED before any edit; not run by L0 |
| B6 | range | 239:2 | planned targeted RED before any edit; not run by L0 |
| B7 | if | 250:2 | planned targeted RED before any edit; not run by L0 |
| B8 | if | 259:3 | planned targeted RED before any edit; not run by L0 |
| B9 | if | 263:3 | planned targeted RED before any edit; not run by L0 |

## Calls and live bindings

| Callee expression | Source location | Current-base evidence/requirement |
|---|---|---|
| errors.New | 212:51 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| StrategyMarket | 214:12 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| loader.proposals.forMarket | 215:52 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| loader.risk.forMarket | 215:88 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| loader.fx.forMarket | 216:3 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| loader.accounts.forMarket | 216:32 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| loader.schedule.forMarket | 216:67 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| len | 217:5 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| errors.New | 219:51 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| proposalAuthority.Proposal | 222:12 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| result.ExecutionTerms.Identity | 223:68 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| accepted.result.ExecutionTerms.Identity | 223:104 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| errors.New | 224:51 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| riskAuthority.bundle.Scope | 226:11 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| riskAuthority.bundle.Validate | 227:12 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| string | 228:3 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| string | 228:27 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| scope.AsOf.Equal | 228:102 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| errors.New | 229:51 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| loader.journal.CurrentPositionCampaignCAS | 231:14 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| string | 231:88 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| uint64 | 233:40 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| errors.New | 234:51 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| riskAuthority.bundle.Entries | 236:13 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| make | 237:13 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| len | 237:50 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| make | 238:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| len | 238:63 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| append | 240:13 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| append | 242:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| MajorDecimal | 247:25 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| result.ExecutionTerms.Entry | 247:25 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| MajorDecimal | 248:23 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| result.ExecutionTerms.EffectiveStop | 248:23 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| MajorDecimal | 249:27 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| result.ExecutionTerms.Target | 249:27 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| errors.New | 251:51 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| strategyFirstLegBindingDigest | 253:19 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| riskAuthority.bundle.Digest | 253:76 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| strings.TrimPrefix | 254:38 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| strings.TrimPrefix | 255:37 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| UTC | 258:10 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| loader.clk.Now | 258:10 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| readCtx.Err | 259:24 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| now.IsZero | 259:48 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| now.After | 259:64 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| account.authority.FreshUntil | 259:74 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| errors.New | 260:38 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| loader.journal.ReservationVersion | 262:26 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| account.authority.ObservedAt | 266:40 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| account.authority.OpenExposure | 266:104 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| string | 270:89 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| account.authority.AccountState | 272:12 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| riskAuthority.bundle.Policy | 273:80 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| loader.guardian.PolicyVersion | 275:26 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| loader.guardian.LimitsDigest | 275:81 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| strings.TrimPrefix | 278:81 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| strings.TrimPrefix | 279:42 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| strings.TrimPrefix | 280:39 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| proposalAuthority.WeeklyBinding | 280:99 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |

## State mutations and fallbacks

- The AST is the exhaustive current-base record of assignments, calls, branches, defers and returns. Before a function body edit, the owning lot must update this map with changed condition semantics and concrete RED/GREEN test evidence.

## Safety conclusion

- L0 status: pre-edit evidence only; no production function was edited and no branch test is claimed as run by L0.
- A named targeted RED or explicit evidence-backed not-applicable rationale is required for every edited branch before GREEN.
