# Function Logic Map: `Context.NewPairedStrategyEntryProductionAssembly`

Source `internal/app/engine/strategy_entry_supervisor.go`; base 28a800f5. AST `ast.json`; syntax `ast-grep.txt`; risk scan `risk-pattern-report.md`.

## Inputs and invariants

Exact authority and legacy compatibility from approved design. Preserve protection and reservations.

## Branches and early returns

| Branch | Line | Condition | Required regression |
|---|---|---|---|
| B1 | 326 | `if c == nil {` | Preserve existing branch; test new contract separately. |
| B2 | 332 | `if c.Journal != nil && strings.TrimSpace(c.Journal.Path()) != "" {` | Preserve existing branch; test new contract separately. |
| B3 | 343 | `if journalPath != "" {` | Preserve existing branch; test new contract separately. |
| B4 | 353 | `if err != nil {` | Preserve existing branch; test new contract separately. |
| B5 | 374 | `if clk != nil {` | Preserve existing branch; test new contract separately. |
| B6 | 390 | `for _, market := range []StrategyMarket{StrategyMarketKR, StrategyMarketUS} {` | Preserve existing branch; test new contract separately. |
| B7 | 395 | `if err != nil {` | Preserve existing branch; test new contract separately. |
| B8 | 402 | `if err := c.publishStrategyRuntime(assembly); err != nil {` | Preserve existing branch; test new contract separately. |

## Calls and live bindings

AST calls: errors.New, collect, newStrategyScheduleAuthorityLoader, strings.TrimSpace, c.Journal.Path, filepath.Join, filepath.Dir, newStrategyCandidateAuthorityLoader, newStrategyRouteAuthorityLoader, strings.ToUpper, newStrategyFXAuthorityLoader, c.productionStrategyLanes, withStrategyLanes, newStrategyProposalAuthorityLoader, proposalAuthority.ResultAuthority, newStrategyRiskAuthorityLoader, newStrategyAccountAuthorityLoader, newProductionStrategyFirstLegAuthorityLoader, newStrategyFirstLegAdmissionBridge, newStrategyDispatchCycle, collectMarket, strategyScheduleStillMatchesAdmission, scheduleAuthority.Snapshot, make, append, c.productionStrategyWorker, NewStrategyEntrySupervisor, candidateAuthority.Snapshot, routeAuthority.Snapshot, fxAuthority.Snapshot, proposalAuthority.Snapshot, riskAuthority.Snapshot, accountAuthority.Snapshot, c.publishStrategyRuntime. Source contains full expressions and error propagation; config binding is documented in discovery.

## State mutations and fallbacks

Journal transaction commits or rolls back atomically; do not infer missing v4 as legacy. Preserve existing protection.

## Safety conclusion

High-risk yes. Test mappings below are planned, not claims of executed coverage.


Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.




Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.
