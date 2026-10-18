# Function Logic Map: `RecordQFinalCampaignFirstLeg`

Base 28a800f5; source `internal/journal/strategy_first_leg_atomic.go`; authoritative branch and call extraction: `ast.json`; syntax cross-check: `ast-grep.txt`; scan: `risk-pattern-report.md`.

## Inputs and invariants

Inputs and durable identities are read from the function signature and callers. The approved design retains legacy v2/v3, authoritative risk and exit protection, atomic reservation, fail-closed v4, and no new network lock. Refer to `analysis/discovery-plan.md` and design decisions for exact binding.

## Branches and early returns

| Branch | Source line | Existing condition / branch | Required regression |
|---|---|---|---|
| B1 | 102 | `if j == nil \|\| j.db == nil {` | Preserve existing branch; new contract refusal and failure paths require added tests. |
| B2 | 105 | `if strings.TrimSpace(request.Issue.Admission.Owner.Key.ProspectiveGeneration) != "" {` | Preserve existing branch; new contract refusal and failure paths require added tests. |
| B3 | 108 | `if err := validateFirstLegCampaignRequest(request.Campaign); err != nil {` | Preserve existing branch; new contract refusal and failure paths require added tests. |
| B4 | 111 | `if request.RouterID != strategyrouter.RouterID \|\| request.RouterVersion != strategyrouter.RouterRelease {` | Preserve existing branch; new contract refusal and failure paths require added tests. |
| B5 | 116 | `if err != nil {` | Preserve existing branch; new contract refusal and failure paths require added tests. |
| B6 | 122 | `if err != nil {` | Preserve existing branch; new contract refusal and failure paths require added tests. |
| B7 | 125 | `if !replayCandidate {` | Preserve existing branch; new contract refusal and failure paths require added tests. |
| B8 | 127 | `if err != nil {` | Preserve existing branch; new contract refusal and failure paths require added tests. |
| B9 | 133 | `if err != nil {` | Preserve existing branch; new contract refusal and failure paths require added tests. |
| B10 | 136 | `if err := validateWeeklyFirstLegBindingTx(ctx, tx, prepared); err != nil {` | Preserve existing branch; new contract refusal and failure paths require added tests. |
| B11 | 140 | `if replayCandidate {` | Preserve existing branch; new contract refusal and failure paths require added tests. |
| B12 | 143 | `if err != nil {` | Preserve existing branch; new contract refusal and failure paths require added tests. |
| B13 | 146 | `if !replayed {` | Preserve existing branch; new contract refusal and failure paths require added tests. |
| B14 | 149 | `if err := verifyFirstLegReplayTx(ctx, tx, prepared, token); err != nil {` | Preserve existing branch; new contract refusal and failure paths require added tests. |
| B15 | 152 | `if err := tx.Commit(); err != nil {` | Preserve existing branch; new contract refusal and failure paths require added tests. |
| B16 | 158 | `if replayed, _, err := recoverQFinalIssueReplayTx(ctx, tx, request.Issue.Admission, prepared.decision,` | Preserve existing branch; new contract refusal and failure paths require added tests. |
| B17 | 161 | `} else if replayed {` | Preserve existing branch; new contract refusal and failure paths require added tests. |
| B18 | 161 | `} else if replayed {` | Preserve existing branch; new contract refusal and failure paths require added tests. |
| B19 | 164 | `if err := reservePrecheck(ctx, tx, prepared.reserve, prepared.reservePlan, j.clk.Now().UTC()); err != nil {` | Preserve existing branch; new contract refusal and failure paths require added tests. |
| B20 | 167 | `if err := insertDecisionRow(ctx, tx, prepared.decision); err != nil {` | Preserve existing branch; new contract refusal and failure paths require added tests. |
| B21 | 171 | `if err != nil {` | Preserve existing branch; new contract refusal and failure paths require added tests. |
| B22 | 176 | `if err != nil {` | Preserve existing branch; new contract refusal and failure paths require added tests. |
| B23 | 179 | `if _, err := insertExactStrategyDecision(ctx, tx, prepared.strategyPlan.Lineage); err != nil {` | Preserve existing branch; new contract refusal and failure paths require added tests. |
| B24 | 182 | `if _, err := insertExactStrategyAttempt(ctx, tx, prepared.strategyPlan, prepared.decision.ID, prepared.decision.AccountRef); err != nil {` | Preserve existing branch; new contract refusal and failure paths require added tests. |
| B25 | 185 | `if err := insertFirstLegCampaignTx(ctx, tx, prepared, token, j.nowString()); err != nil {` | Preserve existing branch; new contract refusal and failure paths require added tests. |
| B26 | 188 | `if err := insertFirstLegBindingTx(ctx, tx, prepared, token, j.nowString()); err != nil {` | Preserve existing branch; new contract refusal and failure paths require added tests. |
| B27 | 191 | `if err := tx.Commit(); err != nil {` | Preserve existing branch; new contract refusal and failure paths require added tests. |

Return lines: 103, 106, 109, 112, 117, 123, 128, 134, 137, 144, 147, 150, 153, 155, 160, 162, 165, 168, 172, 177, 180, 183, 186, 189, 192, 194. Full expressions and effects are in source and `ast-grep.txt`; line excerpts do not replace source review.

## Calls and live bindings

AST calls: errors.New, strings.TrimSpace, fmt.Errorf, validateFirstLegCampaignRequest, j.db.BeginTx, tx.Rollback, firstLegReplayTokenTx, j.mintFirstLegToken, j.prepareQFinalCampaignFirstLeg, validateWeeklyFirstLegBindingTx, recoverQFinalIssueReplayTx, verifyFirstLegReplayTx, tx.Commit, firstLegReceipt, reservePrecheck, UTC, j.clk.Now, insertDecisionRow, reserveRows, commitFreshRiskBucketAdmissionTx, insertExactStrategyDecision, insertExactStrategyAttempt, insertFirstLegCampaignTx, j.nowString, insertFirstLegBindingTx. Journal transaction failures return errors with rollback; projection and digest functions do not perform transport. Existing configuration binding and writer/replay scope are documented in discovery.

## State mutations and fallbacks

Inspect source and AST assignments for current mutations. No legacy fallback is permitted for marker 4. Existing protective exits and in-flight reservations remain authoritative; see approved design for transactional ordering and restrictions.

## Safety conclusion

High-risk impact: yes. Pre-edit inventory only; branch test outcomes are pending and cannot be inferred from these maps.

Source/function binding: `internal/journal/strategy_first_leg_atomic.go` / `Journal.RecordQFinalCampaignFirstLeg`.


Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.

## Current implementation clarification

New admissions require a canonical v4 snapshot and projection generation equal to the expected campaign generation plus one before any durable inserts. Exact historical replay stays ahead of the v4-only new admission guard. Atomic rollback and old v35 replay are directly tested.




Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.
