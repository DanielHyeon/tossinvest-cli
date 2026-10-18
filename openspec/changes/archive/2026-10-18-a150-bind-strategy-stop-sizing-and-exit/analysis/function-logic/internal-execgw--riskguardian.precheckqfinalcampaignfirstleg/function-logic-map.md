# Function Logic Map: `PrecheckQFinalCampaignFirstLeg`

Base 28a800f5; source `internal/execgw/riskguardian_first_leg.go`; authoritative branch and call extraction: `ast.json`; syntax cross-check: `ast-grep.txt`; scan: `risk-pattern-report.md`.

## Inputs and invariants

Inputs and durable identities are read from the function signature and callers. The approved design retains legacy v2/v3, authoritative risk and exit protection, atomic reservation, fail-closed v4, and no new network lock. Refer to `analysis/discovery-plan.md` and design decisions for exact binding.

## Branches and early returns

| Branch | Source line | Existing condition / branch | Required regression |
|---|---|---|---|
| B1 | 51 | `if err != nil {` | Preserve existing branch; new contract refusal and failure paths require added tests. |
| B2 | 57 | `if err != nil {` | Preserve existing branch; new contract refusal and failure paths require added tests. |
| B3 | 61 | `if strings.TrimSpace(owner.Key.ProspectiveGeneration) != "" {` | Preserve existing branch; new contract refusal and failure paths require added tests. |
| B4 | 67 | `if !entryOK \|\| !stopOK \|\| !targetOK {` | Preserve existing branch; new contract refusal and failure paths require added tests. |
| B5 | 71 | `if request.Result.Quantity != entry.decision.QCandidate \|\| lineage.AccountRef != g.accountRef \|\|` | Preserve existing branch; new contract refusal and failure paths require added tests. |
| B6 | 80 | `if currentGeneration < 0 \|\| currentGeneration >= math.MaxInt64 \|\| request.Campaign.ExpectedPositionVersion < 0 \|\|` | Preserve existing branch; new contract refusal and failure paths require added tests. |
| B7 | 85 | `if strings.TrimSpace(request.AttemptID) == "" \|\| request.AttemptID != strings.TrimSpace(request.AttemptID) \|\| request.Revision < 1 \|\|` | Preserve existing branch; new contract refusal and failure paths require added tests. |
| B8 | 91 | `if err != nil {` | Preserve existing branch; new contract refusal and failure paths require added tests. |
| B9 | 105 | `if err != nil {` | Preserve existing branch; new contract refusal and failure paths require added tests. |

Return lines: 52, 58, 62, 68, 76, 82, 87, 92, 106, 111. Full expressions and effects are in source and `ast-grep.txt`; line excerpts do not replace source review.

## Calls and live bindings

AST calls: g.PrecheckQFinalEntry, journal.QFinalPolicyVersion, strings.TrimSpace, qFinalRefusal, MajorDecimal, terms.Entry, terms.EffectiveStop, terms.Target, strings.EqualFold, string, uint64, guardianWeeklyFirstLegBinding, UTC, g.clk.Now, strconv.FormatUint, journal.ProjectAcceptedStrategyflowLineage. Journal transaction failures return errors with rollback; projection and digest functions do not perform transport. Existing configuration binding and writer/replay scope are documented in discovery.

## State mutations and fallbacks

Inspect source and AST assignments for current mutations. No legacy fallback is permitted for marker 4. Existing protective exits and in-flight reservations remain authoritative; see approved design for transactional ordering and restrictions.

## Safety conclusion

High-risk impact: yes. Pre-edit inventory only; branch test outcomes are pending and cannot be inferred from these maps.

Source/function binding: `internal/execgw/riskguardian_first_leg.go` / `RiskGuardian.PrecheckQFinalCampaignFirstLeg`.


Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.




Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.
