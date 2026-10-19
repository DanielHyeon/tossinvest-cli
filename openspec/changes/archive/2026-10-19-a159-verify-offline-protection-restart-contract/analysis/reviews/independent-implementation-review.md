# Independent a159 implementation review

Verdict: PASS — scoped implementation/source and supplied focused evidence review, 2026-10-19. Final source and receipt hashes verified below. Full gate not yet verified. Reviewer did not edit production/tests or execute tests.

## Findings history and disposition

1. Earlier review compared fixture 8/3/5 with initial design 10/3/7. Manager subsequently froze 8/3/5 to reuse existing admitted authority; current design/spec and research-first implementation update agree. This finding is superseded, not an instruction to alter admission values. Historical 10/3/7 source research remains legitimate history.
2. Separate-database unblocked control did not establish within-database isolation. Added TestA159OtherCampaignRemainsIndependent uses two KR symbols in the same real journal, refreshes risk authority through existing wave helper, latches the filled scope and admits a fresh sequence-2 plan for the unaffected scope. Resolved. The sibling's first-leg admission is real; its spy acknowledgement is not evidence of a filled broker order, and no such assertion is needed for this control.
3. Fake SELL journal record omitted the issued reduction decision binding. Adapter now records DecisionID, RiskReducing safety class and derived client order ID; test compares stored attempt decision/class and SELL side with submitted request. Resolved. This still does not claim real Gateway final transport validation.

## Independent source assessment

The primary path uses genuine v4 admission and public Prepare/DispatchStrategyVerified/RecordFill with production ProjectPosition, ApplyPositionCampaignFill and ApplyExitFill hooks. It does not seed a position by direct SQL. Explicit scoped reconcile, close/reopen, duplicate fill/reconcile and fresh next-leg rejection remain on the same persisted lineage. Separate comparisons cover quantity, generation, initial stop/policy identity, campaign version and reconcile event count. Fresh command/latest version and valid sequence have an unblocked positive control, avoiding stale-CAS/idempotent-command false greens.

The exit scenario reads actual working orders and delegates unsettled lookup to a real Gateway over that journal. Fake cancel explicitly errors; price 94 yields no new sell and delayed alert after 31 fake seconds while state remains evaluated. Explicit terminal BUY observation is a sufficient fixture cleanup step, not a new universal BUY cancellation rule. A fresh observation submits quantity 2 against held quantity 3 and the existing sellable floor 2. The pending sell remains the same live order and a subsequent observation adds no second submission. Price 100 is a non-breach control. These are synthetic KR assertions only.

Current reviewed production diff is empty; repository tracked changes outside the new test are PM metadata. No operational account/settings/orders, a156 or a157 were touched by this reviewer.

## Mutation evidence limits

Entry-block mutation fails because the expected ErrExposureBlocked becomes a RECONCILE-to-LEG_PLANNED transition error. It detects refusal reason/guard ordering; another guard still prevents exposure. It does not demonstrate that disabling this check alone allows a new leg.

Replay mutation fails when a zero-delta repeated observation returns Changed=true. It detects the replay result contract; it does not demonstrate duplicate quantity or duplicate economic exposure. Both failures are behavioral, not compile failures. Final commands/source hashes must accompany the unchanged-source GREEN receipts.

## Completion boundary

No live safety certification, full Gateway transport proof, US result, a156 completion, rollout approval, archive or full-gate success follows from this review. Final frozen-source receipt verification is recorded below when available.

## Final frozen snapshot and receipts

Writer supplied final snapshot, independently reread reopen watermark assertion and final source. Focused and race logs each contain Go package success and EXIT_CODE=0. Both current mutation logs show behavioral failure and EXIT_CODE=1 at the documented assertions; exact overlay patches were read. implementation-evidence.md supplies the tagged commands and truthfully limits mutation claims. Earlier final/current receipts are superseded by snapshot receipts. No remaining actionable implementation finding.

The same-DB sibling first leg is admitted before the primary fill because a later admission can legitimately encounter shared UNKNOWN_ACTUAL_RISK. Its subsequent PlanCampaignLeg success proves scoped campaign-latch isolation only; it is not proof a new broker admission could bypass shared risk constraints.

- `internal/app/engine/a159_protection_restart_test.go`: `55dfcd2f0680f91c1dfacc23cade20a110538630d80898a560e490b3a2decd19`
- `openspec/changes/a159-verify-offline-protection-restart-contract/design.md`: `3a9c59447a6bc43abfe396ffc8f70337248a0281de62debb924bb81a5c7d4dd1`
- `openspec/changes/a159-verify-offline-protection-restart-contract/specs/strategy-runtime/spec.md`: `20d18af597bba12a593005ead65f665ab080c9d449a7a1e11773146a4912a0b0`
- `openspec/changes/a159-verify-offline-protection-restart-contract/analysis/implementation-evidence.md`: `499adde332b9ec2fe79c1520ac9f721512545361c447ffb4aeb9ed103b1fc104`
- `openspec/changes/a159-verify-offline-protection-restart-contract/analysis/verification/tossos-a159-focused-snapshot.log`: `e655f51406a8a3752d47ce798d613bf79ae8100f2b885f577cac613f078875bd`
- `openspec/changes/a159-verify-offline-protection-restart-contract/analysis/verification/tossos-a159-race-snapshot.log`: `ba2cd982ac487a83fbb6459e4512e9b38816b04be6edb8c76ccd5a0bb851f204`
- `openspec/changes/a159-verify-offline-protection-restart-contract/analysis/verification/tossos-a159-mutation-block-snapshot.log`: `cc3e4dca410c1e4b2a970197d94343f2020793f5603f0389697084af70309182`
- `openspec/changes/a159-verify-offline-protection-restart-contract/analysis/verification/tossos-a159-mutation-fill-snapshot.log`: `db645031d8c07011ed99299041f4a4ea538f277a5015dcfa52eb263ca6f12ffe`
- `openspec/changes/a159-verify-offline-protection-restart-contract/analysis/verification/mutation-block.patch`: `781c82a382cfa75b7ed720e7874f7370b094631ae024bc07dcd1758968b13514`
- `openspec/changes/a159-verify-offline-protection-restart-contract/analysis/verification/mutation-fill.patch`: `27ae04db36e8db0b2f20521efd5241e4d58fadd4f10cbecadbace67e37c04006`
