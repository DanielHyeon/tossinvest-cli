# a159 gstack implementation review

2026-10-19. **PASS — scoped test implementation/code/security/QA review.** Full aggregate gate, final freshness and lifecycle remain pending. Codex-host installed gstack review/checklist workflow with direct testing/security assessment; no Claude or cross-model run claimed. Reviewer read source and implementer receipts, did not execute tests or modify production/test code. Baseline `1dbff7ced3653e3e9c1aaa892d18d76b1a50d511`.

## Scope and actual composition

Reviewed complete new tagged test file, canonical proposal/design/spec/tasks, research/source paths and implementation evidence. Canonical admission is now **8 ordered / 3 filled / 5 residual**, superseding historical design-review 10/3/7. Existing admitted fixture authority is reused; fake acknowledgement executes actual journal Prepare/DispatchStrategyVerified/RecordFill and installed ProjectPosition, ApplyPositionCampaignFill and ApplyExitFill hooks. No directly inserted position pretends to be a fill. This is fixture authority and fake transport, not full production gateway/broker execution.

Reopen restores the same persisted lineage, then duplicate fill and protection reconcile leave quantity, generation, campaign version, initial stop/policy and reconciliation event count unchanged. Current source explicitly checks reopened watermark 3/5. New sequence 2/current-version/fresh-key request must return exact ErrExposureBlocked, with no new leg or command rows. Separate unblocked control admits sequence 2; same-DB sibling-symbol control has a real admitted first leg and still admits its next leg while the filled sibling is latched. It does not claim the sibling was filled or prove every account/generation permutation.

The exit composition itself reopens the filled, latched journal, replays fill/reconcile, rejects a new leg, verifies actual working BUY and invokes existing workingSet/judge. Unsettled lookup delegates to the real gateway; failed fake cancel does not hide residual. Non-breach submits zero; breach continues evaluated protection, submits zero and produces the delayed alert after 31 fake seconds. Explicit terminal BUY observation is then applied through RecordFill; fresh observation yields a fake risk-reducing SELL of 2 against confirmed floor 2. Decision/class linkage, pending state and live SELL identity are asserted, with no second submit on subsequent observation. This does not create a universal terminal-before-cancel contract, test production quote collection, or exercise actual broker transport. Existing evaluation hooks are used directly, not the whole ObserveOnce price-fetch loop.

## False-green and evidence assessment

Final focused and race receipts exit 0; race package run reports 26.677 seconds. Four named tests and their controls cover the composition. Source SHA matches the receipt's `55dfcd2f0680f91c1dfacc23cade20a110538630d80898a560e490b3a2decd19`. Earlier fixture/scope failures are preserved and explained, not counted as behavioral mutation proof.

Two independent production-condition overlays use unchanged tests and fail behaviorally (test exit 1), not compilation:

- Disabling the explicit entry-block branch changes refusal to the remaining RECONCILE transition error. Exact ErrExposureBlocked assertion catches this. It **does not prove exposure or a new order became allowed**; remaining defense still rejects.
- Disabling identical-fill no-op handling makes replay report Changed=true with delta 0. Replay assertion catches this. It **does not prove doubled position quantity**. Persisted quantity/watermark assertions pass on original source; the mutant demonstrates no-op signaling detection.

These limited mutation claims are accurately disclosed by evidence. No live/new-qFinal/deadline/policy promotion claim follows. Current diff changes new test and a159/PM documents, with no production source/schema/API/dependency/operating changes or a156/a157 edits. SQL used for extra event/row assertions opens the disposable journal read-only. No secrets, external services or existing evidence mutations found.

No unresolved P1/P2 findings in this snapshot. Root-requested import cleanup and reopened watermark check are present and final receipts cover that source. Final metadata may be refreshed separately; scoped PASS is not aggregate gate completion.

Final metadata refresh (2026-10-19T06:17:40+09:00): tasks/review/Manager acceptance and PM metadata included. Task 3 denotes gate/lifecycle preparation, not execution success. Test source remains 55dfcd2f…; no repeated code review or test run. Freeze follows this refresh.

Gate attempt 1 failed stage 5 on the missing literal Function Logic Map not-applicable declaration, with zero existing modified functions. review.md now contains the required exact marker and retains failure cause/receipt. Metadata-only correction; source 55dfcd2f… remains unchanged, scoped code review PASS retained. Aggregate rerun remains pending. No checker or code change.

## Reviewed manifest

Manifest SHA-256: `b95dad9f63bb8a1e9f2f41f9233355ae73289235de4d8b32321527fe68185560`.

```text
6ad0ecc459c6bc6101f1727a75dd97848810f9896106167244ed96f35c50382f  docs/pm/generated/00-master-tracker.md
00a5f0ca882802950318afb9819dbc2d5b58d17b1dc039b3b042cf7f1de5c2d4  docs/pm/generated/01-active-change-map.md
84d4184729f93d7bf1af68fe3d8bd0ddd55344f85dfccb27a844629ea492b908  docs/pm/generated/02-release-readiness.md
ce20c66e49d7889aa233180b83f3059146a08758859b884a94fad74998ae0067  docs/pm/portfolio/_registry.yaml
8fbb9c8833c9c9faf527a5703e6f34bddee7a07ce277462578120f3277e66df5  docs/pm/portfolio/features/FEAT-TOS-013.yaml
1148870b617d4ac466fb26d8c0446f54c4a229f5d51a581a07924547acb7b0e3  docs/pm/portfolio/stories/STORY-TOS-a159.yaml
55dfcd2f0680f91c1dfacc23cade20a110538630d80898a560e490b3a2decd19  internal/app/engine/a159_protection_restart_test.go
499adde332b9ec2fe79c1520ac9f721512545361c447ffb4aeb9ed103b1fc104  openspec/changes/a159-verify-offline-protection-restart-contract/analysis/implementation-evidence.md
ae74d03abe8525400a5ff4dec730002437b995e3c81330f4de59d6a5f951399a  openspec/changes/a159-verify-offline-protection-restart-contract/analysis/reviews/gstack-design-review.md
123d9f9a7f8401604ab3d74c5dd99479b1e49fd41c6d90395dd6d0ee8eae39f5  openspec/changes/a159-verify-offline-protection-restart-contract/analysis/reviews/independent-design-review.md
73295a024cf1451aa0123def32a1c074c4d4768d830bc760d34e13fd60bf2275  openspec/changes/a159-verify-offline-protection-restart-contract/analysis/reviews/independent-implementation-review.md
2f778c62fccd01f0d847df561566bac76185a8afbe6307041f7c70bdfc38d60f  openspec/changes/a159-verify-offline-protection-restart-contract/analysis/reviews/manager-verification.md
c4dead7eaf162ba62c255224f0164ee15b9e2d7336a946b0778590ef7da70cf7  openspec/changes/a159-verify-offline-protection-restart-contract/analysis/verification/gate-attempt1.log
781c82a382cfa75b7ed720e7874f7370b094631ae024bc07dcd1758968b13514  openspec/changes/a159-verify-offline-protection-restart-contract/analysis/verification/mutation-block.patch
27ae04db36e8db0b2f20521efd5241e4d58fadd4f10cbecadbace67e37c04006  openspec/changes/a159-verify-offline-protection-restart-contract/analysis/verification/mutation-fill.patch
f0eb05d0cff7771f07fbc7a3b0f4b479c2c50c72f0bb7751efecea145ecd64f5  openspec/changes/a159-verify-offline-protection-restart-contract/analysis/verification/preedit-sync.log
1b99b8e2594eb107c3f2ec03ee0d72dde80592404ce88ef77c65369d49d28916  openspec/changes/a159-verify-offline-protection-restart-contract/analysis/verification/tossos-a159-fixture3.log
de2569c5a95376b773120d7098c9be50809a62b38e380985c7c920aa81a603e6  openspec/changes/a159-verify-offline-protection-restart-contract/analysis/verification/tossos-a159-fixture4.log
313fff89e4e5d7f6991da8069ba0e183bb0f7f44cdf86cdcbf524a1c378e3184  openspec/changes/a159-verify-offline-protection-restart-contract/analysis/verification/tossos-a159-focused-current2.log
009a6dbf12ec053e35ce26e5c5cdea5ee5b0f789b7d52f2897728cc28ce0f982  openspec/changes/a159-verify-offline-protection-restart-contract/analysis/verification/tossos-a159-focused-final.log
e655f51406a8a3752d47ce798d613bf79ae8100f2b885f577cac613f078875bd  openspec/changes/a159-verify-offline-protection-restart-contract/analysis/verification/tossos-a159-focused-snapshot.log
4825603b21cff458f8f1e01b2b57e1069ea6260de3daf44094313db88eb09e42  openspec/changes/a159-verify-offline-protection-restart-contract/analysis/verification/tossos-a159-mutation-block-final.log
cc3e4dca410c1e4b2a970197d94343f2020793f5603f0389697084af70309182  openspec/changes/a159-verify-offline-protection-restart-contract/analysis/verification/tossos-a159-mutation-block-snapshot.log
e728fbf491daff36c3dfcb71626a814f521a017bf245f710062e0b100cd76792  openspec/changes/a159-verify-offline-protection-restart-contract/analysis/verification/tossos-a159-mutation-fill-final.log
db645031d8c07011ed99299041f4a4ea538f277a5015dcfa52eb263ca6f12ffe  openspec/changes/a159-verify-offline-protection-restart-contract/analysis/verification/tossos-a159-mutation-fill-snapshot.log
affe3cbe01f51aaf69e5ee34ef3aa454d9a14a5d7cf990aa6251b51f91c40ef5  openspec/changes/a159-verify-offline-protection-restart-contract/analysis/verification/tossos-a159-race-final.log
ba2cd982ac487a83fbb6459e4512e9b38816b04be6edb8c76ccd5a0bb851f204  openspec/changes/a159-verify-offline-protection-restart-contract/analysis/verification/tossos-a159-race-snapshot.log
b0111ebfd6fa900ad365e06d224e2ddfab636924a3c64dc20d1d33ed7472b708  openspec/changes/a159-verify-offline-protection-restart-contract/analysis/verification/tossos-a159-scope-attempt2.log
f40d75432ec2b2cb174990c63d08f2598915533e32ab6fb6a9db75956b67479c  openspec/changes/a159-verify-offline-protection-restart-contract/analysis/verification/tossos-a159-scope-probe.log
bd859ff961ff210f91d0bcc8e115db4302eb72d95ce032fd5ce5f41d6ba67031  openspec/changes/a159-verify-offline-protection-restart-contract/base-commit.txt
3a9c59447a6bc43abfe396ffc8f70337248a0281de62debb924bb81a5c7d4dd1  openspec/changes/a159-verify-offline-protection-restart-contract/design.md
6cdab1b56127ac3368c151b02bdee3d16f981be8adf22141c475de6e81b0e3b9  openspec/changes/a159-verify-offline-protection-restart-contract/proposal.md
4c6cf6e00c1e5b11f919c7f2a171840b1ff9951ab159395b010b513f3dfd8ad3  openspec/changes/a159-verify-offline-protection-restart-contract/review.md
20d18af597bba12a593005ead65f665ab080c9d449a7a1e11773146a4912a0b0  openspec/changes/a159-verify-offline-protection-restart-contract/specs/strategy-runtime/spec.md
f1d7e65f3e535414d7e402b2153eb5e99a5b5139abf18620a16b4701f859a82e  openspec/changes/a159-verify-offline-protection-restart-contract/tasks.md
```
