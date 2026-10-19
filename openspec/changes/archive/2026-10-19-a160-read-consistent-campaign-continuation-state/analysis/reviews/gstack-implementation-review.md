# a160 gstack implementation review

2026-10-19. **PASS — scoped code/security/QA review.** Aggregate gate and lifecycle pending. Codex-host installed gstack review/checklist with direct testing/security review; no Claude or cross-model execution claimed. Reviewer inspected full source/tests and implementer receipts, did not execute tests or change code. Baseline `5f3506599fd9cbd40b888aa954ae880244bedf9b`.

## Source assessment

Public ReadOnly method validates request/handle/context, normalizes market/symbol while retaining exact account/campaign identity, opens one read-only transaction, calls the helper with that exact transaction, commits only after complete success, and returns a zero snapshot on every failure. All helper queries and scanners use the same transaction. No alternate DB lookup, writer, repair, schema mutation, external service or new execution consumer is introduced.

Campaign scope/current actual generation, latest scoped OPEN position, nonempty exact entry-decision binding and positive authoritative projection version agree before success. PLANNED/prospective/closed/missing/mismatched facts fail. Entry-blocked/RECONCILE facts remain inspectable without clearing them. Legs and watermarks are ordered, watermark scope is checked, orphan leg references fail, no required legs fails; a genuine planned second leg without its own watermark remains observable. Stored values are facts rather than a complete semantic replay audit.

Every success contains FullPlanUnavailable. No Ready, qFinal, reconstructed risk, new reservation/decision/lease or full-plan inference is returned. Existing canonical allocation policies remain untouched. The only existing-function change registers exactly the new read-only method in the explicit reflection allowlist; forbidden-method assertions remain unchanged. Its current AST/FLM/BTM and checker receipt are present, and acknowledge the unchanged negative branch was not separately mutated.

## Finding history and tests

Independent reviewer found that foreign account/market/symbol watermarks could be returned under an otherwise valid campaign. Three preserved behavioral failures appear in foreign-red.log. Current reader rejects any such row before returning the result; focused tests verify empty failure for all three fields. This is resolved, not a clean first-pass claim.

Final tests cover KR/US stored facts, same-DB peer isolation, requested/stored scope, generation replacement, missing/foreign decision, missing projection/legs/table, orphan and foreign watermarks, blocked/closed/prospective states, planned-leg absence of watermark, invalid/cancelled/nil/closed handles, reopen equality and logical table rows unchanged. These use disposable synthetic database setup, including direct SQL position updates. They do not claim actual RecordFill, real admission or real broker evidence for the fixture.

The deterministic concurrency test establishes an old snapshot with a private tx reader, commits synthetic position update plus real campaign fill hook on a separate connection, then observes old position/watermark via the old tx and newer values plus increased campaign/projection versions via a subsequent public read. **Public wrapper transaction ownership is statically reviewed; concurrent execution itself exercises the private helper, not an instrumented public invocation.** No sleep-based claim or new framework is required for this bounded proof. It does not guarantee return-time freshness or execution authority.

Final focused receipt passes. Final race receipt `tossos-a160-race-final3.log` reports journal PASS, 92.346s, EXIT_CODE=0 on source `289432fd3c6d98e9c3801be8b504af3976c030f205a08b323ac82ef5ab173d89` and test `d69c0544d8291b4fce663d1814ce43b881961ed2a11de743d2b42fa0e22afe58`. Decision-binding overlay produced behavioral failure on foreign decision; it proves that guard's detection, not mutation coverage for all scope/transaction branches. Earlier focused/race receipts are retained history, not substitutes for final source evidence.

No unresolved actionable P1/P2 finding in current snapshot. No live readiness, a156 completion, a157 work, actual-risk restoration or later-leg authority claimed. Final metadata may be refreshed without repeating code review; actual full gate remains separate.

## Gate 1 compatibility addendum — PASS

Aggregate gate attempt 1 failed stage 7 on TestExitEligibilityHasOneSpelling (exit 2), preserved in gate-attempt1.log. Reviewed minimal repair: exactly one path/reason in eligibilitySpellers for this reader, whose EntryDecisionID use checks exact campaign lineage rather than the general exit-eligibility predicate. Existing walker, tokens and assertions have no diff. Reader/test hashes remain 289432fd… / d69c0544…; no production behavior changed. Position focused test (0.038s), journal focused test (10.558s) and logic-map check pass. Removing an existing provenance exception in a disposable overlay still triggers the original guard (behavioral exit 1); this verifies its rejection remains active, not unrestricted AST completeness. Design/evidence record the exception and failed gate. Scoped repair PASS; fresh full gate is required. Prior metadata freeze updates are included as preparation, not gate completion.

## Reviewed manifest

Manifest SHA-256: `3e8a1a40444b0eb6f112807d591961a7259f86e275ff595a5e52d71391829ff7`.

```text
289432fd3c6d98e9c3801be8b504af3976c030f205a08b323ac82ef5ab173d89  internal/journal/campaign_continuation_snapshot.go
d69c0544d8291b4fce663d1814ce43b881961ed2a11de743d2b42fa0e22afe58  internal/journal/campaign_continuation_snapshot_test.go
f4b38427349a3a51c99342b031940d54d3adfc00b961b64a9dde06c41bb7e108  internal/journal/readonly_test.go
b15021c0a6b0c432f6ddd58ca1c7120645eac3ef65f6c5baca9cde13808cfa28  internal/position/eligibility_drift_test.go
0fa72d5bc35a06cc425f0ce1e96e272ae291df98ad44421ef5f34f5352a4101c  openspec/changes/a160-read-consistent-campaign-continuation-state/analysis/code-context/eligibility/post-ast.json
c5adee8ee29bbd768191afebce32f32b7081a0b0f394671859387839d42ffbfd  openspec/changes/a160-read-consistent-campaign-continuation-state/analysis/code-context/eligibility/pre-ast.json
4868c7c3339d057ca34858dc1daf456bc574e85ea94d562764c0726f883c49b8  openspec/changes/a160-read-consistent-campaign-continuation-state/analysis/code-context/eligibility/pre-edit.md
6e5459f49f552ced0d88298269fd8dbab873083665ca0b039b3fae73d6d7975c  openspec/changes/a160-read-consistent-campaign-continuation-state/analysis/function-logic/internal-journal--testthereadonlyhandlehasnowritemethods/ast.json
e4ed51ae85a8f3a78dbbe25eab574f251fc7de70b731661b792bc662e90117c1  openspec/changes/a160-read-consistent-campaign-continuation-state/analysis/function-logic/internal-journal--testthereadonlyhandlehasnowritemethods/branch-test-map.md
6f9e5737ed33cc2e07dc1c24ac9205d154b25329d735ebb882c12f5248c83606  openspec/changes/a160-read-consistent-campaign-continuation-state/analysis/function-logic/internal-journal--testthereadonlyhandlehasnowritemethods/function-logic-map.md
7c28f182e9038df7a2df29db876ce874062bbc4168334489962a955620e6afe9  openspec/changes/a160-read-consistent-campaign-continuation-state/analysis/function-logic/internal-journal--testthereadonlyhandlehasnowritemethods/risk-pattern-report.md
367f4a67f6e4c3e9cfa1954022d0842de33e8a70d6e21e8fbfcb44961fd7ef7b  openspec/changes/a160-read-consistent-campaign-continuation-state/analysis/implementation-evidence.md
8ecc0ced95f8470a34910f2e71b0c49896cd6fd4ebbc705805c36af67332004b  openspec/changes/a160-read-consistent-campaign-continuation-state/analysis/pre-edit.md
cb695516b539040baa4b1d97f13cb9772736c7dcb9d1eb023141499f05f42af2  openspec/changes/a160-read-consistent-campaign-continuation-state/analysis/research-first.md
746f36c82e4a36b2044232860a4104eb3c6ed1899b02e8d928817b956942cbe5  openspec/changes/a160-read-consistent-campaign-continuation-state/analysis/verification/gate-attempt1.log
22e0ac1001e76b85380380afd0c9544ffa7f4451d3c8795eb3075a78a923b814  openspec/changes/a160-read-consistent-campaign-continuation-state/analysis/verification/tossos-a160-focused-final.log
0cef644357fc63ecdf54963f9fc339029dfe0bd7a5aa98d60f7665f3004ed867  openspec/changes/a160-read-consistent-campaign-continuation-state/analysis/verification/tossos-a160-focused-final2.log
29f3fecb474d5e382e1b450f438d180ca42b7e2aa15b13f11d7fe57606e6017a  openspec/changes/a160-read-consistent-campaign-continuation-state/analysis/verification/tossos-a160-focused6.log
e6af60937c18d7e177460ddf3f55e885e65b676289c98c2fc2228ccfd75a5a07  openspec/changes/a160-read-consistent-campaign-continuation-state/analysis/verification/tossos-a160-foreign-red.log
c7cd74aa242a5aa044b572dbf454fb6dd0b96324b776945ce509a74205b11ced  openspec/changes/a160-read-consistent-campaign-continuation-state/analysis/verification/tossos-a160-logic-check-final.log
07f3768e5a84d8610a5564815b3585760d86064a42ea660c2bbc4c25a7398593  openspec/changes/a160-read-consistent-campaign-continuation-state/analysis/verification/tossos-a160-mutation-final.log
2056d906a1443d5c67731fafdaece97505194ec05061d388508520e00276dd7e  openspec/changes/a160-read-consistent-campaign-continuation-state/analysis/verification/tossos-a160-mutation-final2.log
ef07bd4a36712edc8b999e85eb9221d17ddacfe95d10ac4da6d159b999d09869  openspec/changes/a160-read-consistent-campaign-continuation-state/analysis/verification/tossos-a160-pre-sync.log
49ce2dcf77d574b1ef8a4a823941324d4b7b2585d42f95db92f3a907d31963ac  openspec/changes/a160-read-consistent-campaign-continuation-state/analysis/verification/tossos-a160-race-final2.log
f10f103f7a4dbede59976b9efa6d6320187eb21bd581f94a8227a974f92f0fa1  openspec/changes/a160-read-consistent-campaign-continuation-state/analysis/verification/tossos-a160-race-final3.log
9b7acba31085f7635d1b272740aa32356784033a94aeb0e0bd3a0750e3c225c8  openspec/changes/a160-read-consistent-campaign-continuation-state/analysis/verification/tossos-a160-repair-journal.log
c7cd74aa242a5aa044b572dbf454fb6dd0b96324b776945ce509a74205b11ced  openspec/changes/a160-read-consistent-campaign-continuation-state/analysis/verification/tossos-a160-repair-logic.log
96a75179b0e440562901b3d218e6240b32de587a7735e7de2534e698997c8cca  openspec/changes/a160-read-consistent-campaign-continuation-state/analysis/verification/tossos-a160-repair-mutation.log
75d2042a84b4dfa224c4333de941a30669df0e01cfd8446cbe9ce8b888ede854  openspec/changes/a160-read-consistent-campaign-continuation-state/analysis/verification/tossos-a160-repair-position.log
fc721aaed4376830c513627aa0ea102bf4fc2a427942ed6cd69bd73a4e3fcb9f  openspec/changes/a160-read-consistent-campaign-continuation-state/base-commit.txt
c5483b420b116fefbf7bd8029192586b726d723dcb85a8a352bca79a2bb37127  openspec/changes/a160-read-consistent-campaign-continuation-state/design.md
f185c414dce9e040bc62856c83c9791303bc55800b37be6a8fb0712dc232b071  openspec/changes/a160-read-consistent-campaign-continuation-state/proposal.md
ce1eeaa55d609fe9f501df0d3aae4de0b4f358d597b73f887554bfa3eabe3d8f  openspec/changes/a160-read-consistent-campaign-continuation-state/review.md
e1b3335b97fde646249b9882502d46ed489aba0fbb19a66b7b0f9d881a3e333b  openspec/changes/a160-read-consistent-campaign-continuation-state/specs/position-campaigns/spec.md
2cec6de20c2fcd120efdb5b115b5281deb899d64091200963cd55c1cb0271b99  openspec/changes/a160-read-consistent-campaign-continuation-state/tasks.md
```
