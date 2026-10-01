DECLARATION: I did not read or write anything under ~/.codex and made no file changes.

Reviewed `18109568` statically. `a127/` denotes `openspec/changes/a127-strategy-authorities-read-the-current-ledger/`. No P0 established.

1. **P1 — D6’s exhaustive chain and “only 11–12 are automatic” conclusion are false.**  
   `a127/design.md:117–134`.

   “Promotion is not an order gate” is correct, but refresh workers still require `accepting`, an unlatched worker and a cycle (`internal/app/engine/strategy_entry_supervisor.go:1041`). These are **automatic runtime conditions**, preceding dispatch.

   Additional necessary conditions omitted from the chain:
   - **Automatic:** campaign unclaimed and FLAT/CLOSED (`internal/app/engine/strategy_market_handoff_delivery.go:42`); dispatch-owner acquisition and successful lease issuance/claim/fencing (`strategy_dispatch_cycle.go:147,190,196`; `internal/execgw/strategy_gateway.go:93`).
   - **Automatic:** no conflicting outstanding attempt, unspent decision, supported order shape and sufficient readable buying power (`internal/execgw/gateway.go:525,535,594`; `failclosed.go:40,129`). Protection readiness also depends on runtime evidence, not merely human deployment.
   - **Human/operator configuration:** trading place/sell/cancel and LIVE permission (`internal/app/engine/interlock.go:571`), plus usable official credentials (`engine.go:379`). LIVE permission remains enforced by `internal/trading/service.go:244`; official order capability by `:188`. Gateway generates execute/confirmation options (`internal/execgw/gateway.go:461`), so these are not fresh per-order human approval.

   Conditions 11–12 cannot account for all these gates as currently described. Narrow the completeness claim or enumerate them.

2. **P2 — D7’s error precedence holds inside journal loading, not across the risk loader.**  
   `a127/design.md:149–155`.

   Preparing immediately after risk’s version check (`internal/riskbucket/production_snapshot_authority.go:384`) and before latch query `:389` is achievable. Route’s corresponding insertion is after `internal/strategyrouter/production.go:616`, before returning the transaction at `:621`.

   However, `LoadProductionRiskSnapshotAuthority` calls `bindProductionRiskInputs` at `:169`, before journal loading at `:173`. Missing symbol-sector mapping returns `ErrProductionRiskScopeRefused` at `:311`. Consequently, that refusal can precede file/version/schema defects. Post-version prepares cannot establish the stated global precedence. Qualify it as applying **after successful policy/input binding**, specifically before journal-derived scope refusals.

3. **P2 — The SHALL overclaims preparation ordering and scope-refusal precedence.**  
   `a127/specs/strategy-runtime/spec.md:9–11`.

   The same-read-only-transaction requirement is achievable and testable for both loaders; S13 appropriately strengthens risk receiver/lifetime checks. But “all read queries before the first read” literally includes `PRAGMA user_version`, which D7 executes before preparing data queries. Specify “after version verification, before the first ledger-data query.”

   The unconditional precedence over scope refusal also contradicts finding 2. S14 tests latch refusal only; it does not establish precedence over policy-binding refusal. Add that boundary explicitly.

4. **P3 — Several evidence coordinates are stale.**  
   `a127/design.md:143–151`.

   At this commit: risk version/latch/usage are `:384/:389/:408`, not `:375/:380/:399`; route BeginTx is `:610`, not `:603`; no-active-owner return is `:659–660`, not `:652–653`. D6’s `engine.go:226/:682`, proposal/account/handoff citations and supervisor promotion explanation remain substantively correct; the refresh return is `:1045–1046`.

VERDICT: FAIL