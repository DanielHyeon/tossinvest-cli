# Function Logic Map: `collectStrategyFirstLegAuthority`

Base 28a800f5; source `internal/app/engine/strategy_account_first_leg_authority.go`; authoritative branch and call extraction: `ast.json`; syntax cross-check: `ast-grep.txt`; scan: `risk-pattern-report.md`.

## Inputs and invariants

Inputs and durable identities are read from the function signature and callers. The approved design retains legacy v2/v3, authoritative risk and exit protection, atomic reservation, fail-closed v4, and no new network lock. Refer to `analysis/discovery-plan.md` and design decisions for exact binding.

## Branches and early returns

| Branch | Source line | Existing condition / branch | Required regression |
|---|---|---|---|
| B1 | 289 | `if loader == nil \|\| ctx == nil \|\| loader.clk == nil \|\| loader.journal == nil \|\| loader.guardian == nil {` | Preserve existing branch; new contract refusal and failure paths require added tests. |
| B2 | 295 | `if !riskAuthority.snapshot.Ready \|\| !fx.snapshot.Ready \|\| !account.snapshot.Ready \|\|` | Preserve existing branch; new contract refusal and failure paths require added tests. |
| B3 | 302 | `if !scoped {` | Preserve existing branch; new contract refusal and failure paths require added tests. |
| B4 | 306 | `if result.Lineage.Identity != accepted.result.Lineage.Identity \|\| result.ExecutionTerms.Identity() != accepted.result.ExecutionTerms.Identity() {` | Preserve existing branch; new contract refusal and failure paths require added tests. |
| B5 | 311 | `if !proposal.familyActivation().Verified() && len(proposal.entries) != 1 {` | Preserve existing branch; new contract refusal and failure paths require added tests. |
| B6 | 318 | `if !keyed { // 봉인이 정규화를 보장하므로 도달 불가 — 바뀌면 범위 거절이 아니라 결함(리뷰 A #4)` | Preserve existing branch; new contract refusal and failure paths require added tests. |
| B7 | 324 | `if !riskScoped {` | Preserve existing branch; new contract refusal and failure paths require added tests. |
| B8 | 326 | `if !scopeLocal {` | Preserve existing branch; new contract refusal and failure paths require added tests. |
| B9 | 332 | `if !accountScoped {` | Preserve existing branch; new contract refusal and failure paths require added tests. |
| B10 | 339 | `if !currencyKnown {` | Preserve existing branch; new contract refusal and failure paths require added tests. |
| B11 | 343 | `if err := riskBundle.Validate(scope); err != nil \|\| scope.AccountID != result.Lineage.AccountRef \|\|` | Preserve existing branch; new contract refusal and failure paths require added tests. |
| B12 | 348 | `if err != nil \|\| cas.Claimed \|\| cas.State != "FLAT" && cas.State != "CLOSED" \|\| cas.Generation < 0 \|\|` | Preserve existing branch; new contract refusal and failure paths require added tests. |
| B13 | 355 | `for _, entry := range entries {` | Preserve existing branch; new contract refusal and failure paths require added tests. |
| B14 | 366 | `if !entryOK \|\| !stopOK \|\| !targetOK {` | Preserve existing branch; new contract refusal and failure paths require added tests. |
| B15 | 375 | `if readCtx == nil \|\| readCtx.Err() != nil \|\| now.IsZero() \|\| now.After(accountAuthority.FreshUntil()) {` | Preserve existing branch; new contract refusal and failure paths require added tests. |
| B16 | 379 | `if versionErr != nil {` | Preserve existing branch; new contract refusal and failure paths require added tests. |

Return lines: 290, 297, 303, 307, 312, 319, 327, 329, 334, 340, 345, 350, 367, 376, 380, 382, 386. Full expressions and effects are in source and `ast-grep.txt`; line excerpts do not replace source review.

## Calls and live bindings

AST calls: errors.New, StrategyMarket, loader.proposals.forMarket, loader.risk.forMarket, loader.fx.forMarket, loader.accounts.forMarket, loader.schedule.forMarket, proposal.authorityForOwnerScope, proposalAuthority.Proposal, result.ExecutionTerms.Identity, accepted.result.ExecutionTerms.Identity, Verified, proposal.familyActivation, len, strategyOwnerKeyOf, riskAuthority.forScope, riskAuthority.riskScopeCause, fmt.Errorf, account.forScope, account.accountScopeCause, riskBundle.Scope, riskBundle.Validate, string, scope.AsOf.Equal, loader.journal.CurrentPositionCampaignCAS, uint64, riskBundle.Entries, make, append, MajorDecimal, result.ExecutionTerms.Entry, result.ExecutionTerms.EffectiveStop, result.ExecutionTerms.Target, strategyFirstLegBindingDigest, riskBundle.Digest, strings.TrimPrefix, UTC, loader.clk.Now, readCtx.Err, now.IsZero, now.After, accountAuthority.FreshUntil, loader.journal.ReservationVersion, accountAuthority.ObservedAt, accountAuthority.OpenExposure, accountAuthority.AccountState, riskBundle.Policy, otherStrategyMarket, loader.guardian.PolicyVersion, loader.guardian.LimitsDigest, proposalAuthority.WeeklyBinding. Journal transaction failures return errors with rollback; projection and digest functions do not perform transport. Existing configuration binding and writer/replay scope are documented in discovery.

## State mutations and fallbacks

Inspect source and AST assignments for current mutations. No legacy fallback is permitted for marker 4. Existing protective exits and in-flight reservations remain authoritative; see approved design for transactional ordering and restrictions.

## Safety conclusion

High-risk impact: yes. Pre-edit inventory only; branch test outcomes are pending and cannot be inferred from these maps.

Source/function binding: `internal/app/engine/strategy_account_first_leg_authority.go` / `productionStrategyFirstLegAuthorityLoader.collectStrategyFirstLegAuthority`.


Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.




Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.
