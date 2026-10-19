# Independent a160 implementation review

Final verdict: PASS — scoped source and supplied focused/race evidence review, 2026-10-19. Initial HOLD history is preserved below. No reviewer source edits or test execution. Full gate not verified.

## P2 — watermark scope is not validated

Initial readCampaignContinuationSnapshot SELECT joined watermarks to campaign by campaign_id and checked only leg existence. Stored watermark account_ref, market and symbol were omitted by the reusable CampaignOrderRecord scanner and never compared. A row retaining campaign_id/leg_sequence but carrying another account or market/symbol would be returned as part of the requested scoped snapshot. This violates the exact-scope read contract even though it grants no authority. Require explicit same-transaction mismatch rejection and zero result; do not filter the mismatching row away and return partial success. Manager agrees; Terra was asked for a real failing counterexample before repair.

## Verification gaps under active revision

Manager identified public-wrapper transaction binding, full protected-table logical dump, actual close/reopen, existing peer scope, missing/orphan/latest generation/closed/nil/schema cases. Initial concurrency test establishes a transaction then calls the private reader directly; it proves helper snapshot behavior, not independently the public wrapper's transaction use. A bounded source/AST check can support wrapper binding if labelled accurately; do not call it a public-wrapper concurrency behavior test. No production hook/framework should be introduced solely to instrument this.

Initial persistence test opens another ReadOnly handle while the original stays open; that is an additional handle, not a close/reopen test. Initial logical dump omitted risk reservations, leases and exit rows. Final evidence must match the expanded assertions, rather than reusing an earlier PASS log.

## Source assessment and evidence limits

Public wrapper validates nil context/handle and exact account/campaign request identity, normalizes market/symbol, and uses BeginTx(ReadOnly), helper, Commit with zero results on errors. Helper enforces current OPEN generation, authoritative projection version and nonempty equal decision binding; blocked/reconcile campaigns remain inspectable; every successful result carries FullPlanUnavailable. No numeric-risk reconstruction, qFinal or authority creation was found in the initial implementation.

Fixtures manually insert position rows and call the campaign fill helper; the US case manually changes stored scope. These are synthetic reader fixtures, not full RecordFill/US production admission or broker evidence. This is permissible for a read-only library, but must be described honestly. Prior source hashes and final receipt inventory will be recorded after writer freeze.

## Rereview of reader 289432fd (intermediate test snapshot)

Watermark-scope P2 resolved: an explicit same-transaction EXISTS query detects mismatched account/market/symbol and returns zero result; invalid rows are not filtered away. Writer supplied three actual failing pre-fix counterexamples and focused GREEN. Existing read-only method census change adds only the new SELECT-only method and rationale; rejection assertions are unchanged.

All user tables now participate in logical row comparison, and ReadOnly is actually closed before reopening. Latest generation, closed handle and nil handle cases were added. Direct source review confirms the public wrapper creates one read-only transaction, passes that same transaction into the helper, and returns no usable result on error/commit failure. The executable concurrency scenario establishes a transaction snapshot, commits a separate writer, then calls the private helper; the public wrapper is used for the subsequent fresh read. This supports helper snapshot behavior plus source-verified wrapper binding, not a forced-interleaving public-wrapper execution.

Remaining fixture detail sent to writer: the test named USAndPlannedLegWithoutWatermark currently deletes the watermark of a PARTIAL leg, rather than creating a valid PLANNED leg. It must not be reported as evidence of that precise planned-leg scenario until amended. Existing peer data and schema-error cases also await writer coverage disposition. No additional production correctness defect found in this rereview. Final test/source receipt remains pending.

## Final disposition and evidence

All initial actionable findings resolved. Source 289432fd validates foreign watermark scope explicitly inside the same transaction. Final tests create an actual second PLANNED leg through PlanCampaignLeg and preserve leg-1 watermark, add a seeded same-DB peer with independent quantity/result, missing table, and nil/closed handle checks. Version assertions now require both campaign and projection version to increase in the fresh read. ReadOnly close/reopen and all user-table logical-row comparisons are present. No production caller or authority-producing API was added.

Final focused-final2 log and race-final3 log were read: both EXIT_CODE=0, with race package execution 92.346s. The different-decision overlay fails behaviorally with EXIT_CODE=1; it was run before only the unrelated concurrency version assertion changed, as implementation evidence explicitly states. Foreign account/market/symbol regression RED is retained. Method census AST/FLM/BTM checker receipt reports evidence complete and EXIT_CODE=0. These are writer-executed receipts independently inspected, not reviewer test execution.

Limits retained: transaction interleaving is exercised through the internal helper, while public-wrapper transaction binding is directly source-reviewed. Synthetic SQL position setup/update and manually changed US scope are reader fixtures, not full RecordFill or real US admission/broker proof. The method reports stored facts with FullPlanUnavailable; no inference of complete plan, risk, sellable, qFinal, next-leg permission, a156 completion or live readiness is supported. Aggregate gate/lifecycle remain pending.

## Final reviewed hashes

- `internal/journal/campaign_continuation_snapshot.go`: `289432fd3c6d98e9c3801be8b504af3976c030f205a08b323ac82ef5ab173d89`
- `internal/journal/campaign_continuation_snapshot_test.go`: `d69c0544d8291b4fce663d1814ce43b881961ed2a11de743d2b42fa0e22afe58`
- `internal/journal/readonly_test.go`: `f4b38427349a3a51c99342b031940d54d3adfc00b961b64a9dde06c41bb7e108`
- `openspec/changes/a160-read-consistent-campaign-continuation-state/analysis/implementation-evidence.md`: `6f111fa94e1822ba4f1072f7bdbc07d59a2877d6d9e0d71ed8f35aca14e9fc74`
- `openspec/changes/a160-read-consistent-campaign-continuation-state/analysis/verification/tossos-a160-focused-final2.log`: `0cef644357fc63ecdf54963f9fc339029dfe0bd7a5aa98d60f7665f3004ed867`
- `openspec/changes/a160-read-consistent-campaign-continuation-state/analysis/verification/tossos-a160-race-final3.log`: `f10f103f7a4dbede59976b9efa6d6320187eb21bd581f94a8227a974f92f0fa1`
- `openspec/changes/a160-read-consistent-campaign-continuation-state/analysis/verification/tossos-a160-foreign-red.log`: `e6af60937c18d7e177460ddf3f55e885e65b676289c98c2fc2228ccfd75a5a07`
- `openspec/changes/a160-read-consistent-campaign-continuation-state/analysis/verification/tossos-a160-mutation-final2.log`: `2056d906a1443d5c67731fafdaece97505194ec05061d388508520e00276dd7e`
- `openspec/changes/a160-read-consistent-campaign-continuation-state/analysis/verification/tossos-a160-logic-check-final.log`: `c7cd74aa242a5aa044b572dbf454fb6dd0b96324b776945ce509a74205b11ced`

## Gate-attempt-1 narrow repair review

Scoped verdict: PASS for the test-only repair. Original aggregate gate failed make test (EXIT_CODE=2), because TestExitEligibilityHasOneSpelling did not yet list the new reader's EntryDecisionID use. Failure remains preserved; aggregate retry success is not yet established.

Reviewed exact diff: eligibilitySpellers adds only internal/journal/campaign_continuation_snapshot.go with reason that entry decision identity binds campaign to current position, not general exit eligibility. Scanner function, token list, truth-table assertions and production ExitEligible predicate are unchanged. Reader and new test hashes remain 289432fd / d69c0544. This exact path exception is consistent with Manager's amended design and does not authorize arbitrary future spellings.

Read repair-position and repair-journal logs (EXIT_CODE=0), repair-logic (evidence complete, EXIT_CODE=0), and repair-mutation (EXIT_CODE=1). The mutation removes the existing provenance path from the allowlist and causes the real scanner to report its entry_decision_id spelling. It demonstrates the existing unlisted-path failure still runs; it does not prove every conceivable eligibility drift is detected. No reviewer test execution or source edits. Final sync/check and actual complete aggregate retry remain required.

Repair hashes:

- `/tmp/tossos-a148-impl-20261018/internal/position/eligibility_drift_test.go`: `b15021c0a6b0c432f6ddd58ca1c7120645eac3ef65f6c5baca9cde13808cfa28`
- `/tmp/tossos-a148-impl-20261018/openspec/changes/a160-read-consistent-campaign-continuation-state/design.md`: `c5483b420b116fefbf7bd8029192586b726d723dcb85a8a352bca79a2bb37127`
- `/tmp/tossos-a148-impl-20261018/openspec/changes/a160-read-consistent-campaign-continuation-state/analysis/verification/gate-attempt1.log`: `746f36c82e4a36b2044232860a4104eb3c6ed1899b02e8d928817b956942cbe5`
- `/tmp/tossos-a160-repair-position.log`: `75d2042a84b4dfa224c4333de941a30669df0e01cfd8446cbe9ce8b888ede854`
- `/tmp/tossos-a160-repair-journal.log`: `9b7acba31085f7635d1b272740aa32356784033a94aeb0e0bd3a0750e3c225c8`
- `/tmp/tossos-a160-repair-logic.log`: `c7cd74aa242a5aa044b572dbf454fb6dd0b96324b776945ce509a74205b11ced`
- `/tmp/tossos-a160-repair-mutation.log`: `96a75179b0e440562901b3d218e6240b32de587a7735e7de2534e698997c8cca`
