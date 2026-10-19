# a150 independent implementation review

Date: 2026-10-18. Base `28a800f5ed5eed7c0db0aed4f435d0778b034191`. Independent reviewer, separate from Terra sole implementation/test writer. Scope: approved design/spec/tasks, tracked diff, new production/test files, historical fixture evidence, relevant existing Gateway propagation and AST/FLM context. Reviewer made no source/test edits, ran no tests/gates, made no live calls or memory writes; only this report is written.

## Verdict

**PASS for implementation review of the refreshed hashes below.** This rereview supersedes the earlier intermediate PASS and its subsequent withdrawal. Admitted-generation, stale-floor selector and empty-snapshot seed-window findings are resolved in the inspected source; current focused receipts are reviewed. Full aggregate gate, final SDD checks, Manager acceptance and commit/push/archive remain separate, not completed by this review. No live operational or profitability claim.

## Finding resolved: P1 admitted generation not checked against position

Initial StrategyProtectionForPosition compared supplied/stored position and mutable campaign actual generation, without checking the generation sealed inside v4 strategyflow. A stored position with generation2 could retain generation1's entry decision and receive usable policy despite mismatch; workingSet could then initialize from it. Corrupting campaign actual generation to2 also bypassed that comparison. Terra reproduced the counterexample in generation-red receipt.

Revised read validates the v4 projection generation against campaign expected+1, returns no usable policy for foreign stored position generation, and carries admitted generation into final transport checks. New journal admission independently enforces expected+1. Tests cover campaign actual generation1 and2 with a foreign position2, plus atomic refusal of mismatched new admission. Existing valid exit preservation remains separate from permission to seed a new exit. Source and current journal/engine receipts confirm resolution.

## Follow-up findings resolved: floor races and order side effects

The previous review was withdrawn after `/tmp/tossos-a150-stale-floor-red.log` demonstrated stale floor95/high100 versus newer saved90/high101 could quarantine valid protection. A snapshot-ID-only repair was then held after Terra reproduced a legacy scalar-write race while both snapshot IDs remained empty (`seed-cas-red`).

Final floor-only CAS compares the exact observed snapshot ID (including empty) plus baseline, high water, ratchet level, rung, position quantity, cumulative taken ratio and pending action/level/intent. Existing lifecycle and position-generation guards remain. Comparison occurs inside the journal transaction before recovery selection or state/event writes. A stale candidate returns observation conflict without quarantining the valid saved exit; selector incompatibility in this explicitly fenced floor path also returns conflict rather than quarantine.

Engine captures pre-floor state, raises only the evaluation baseline, and passes the observed tuple to journal. On the first orderable floor reconciliation it supplies the distinct truthful admitted-floor suppression reason before clearTheSymbol, so no outstanding-order cancel, read-for-clear or place occurs on this attempt. Journal returns a distinct floor-suppressed outcome; engine submits only an actually Armed result. A subsequent normal observation using the stored floor can follow existing order handling at the same market price. No fake observation identity, new re-arm privilege or network lock added.

Tests inspected cover newer full snapshot, empty-snapshot legacy watermark update, simulated quantity/taken-ratio change, pending-state change, pre-existing pending sale preservation, restart immediately after suppressed breach, next same-price arming, and pending-sale deduplication. The final tuple-CAS receipt `/tmp/tossos-a150-seed-cas-green.log` reports engine PASS3.348s; preceding floor-final receipt PASS2.373s. This explicitly introduces one normal observation-cycle delay for the initial floor recovery, as approved in design; it does not promise immediate execution or cancel existing working protection. Persistent conflicts/errors remain visible rather than fabricating success.

## Other reviewed boundaries

- New first-leg journal admission requires v4; only historical exact replay remains legacy. No public fixture bypass added. Independent decision marker0/4 is written with admission and immutable by trigger. Payload loss cannot make marker4 legacy. Projection version4 includes immutable exit policy kind/ID/version/digest; old RiskIntent and v2/v3 serialization preserved.
- Writer supplies pre-change v2/v3 golden bytes and genuine v35 first-leg database. Migration/read/replay tests use these rather than creating historical rows through the new writer. Marker downgrade/mismatch, lost binding and injected post-decision reservation failure have explicit negative tests.
- Production loader copies selected configured policy and includes it in binding identity. Guardian precheck requires snapshot. Current options do not replace a valid admitted snapshot at first exit initialization. Existing legacy/adopted paths retain their compatibility behavior.
- Reconciliation transaction scopes account, campaign, decision and actual position generation, appends idempotent event and blocks entry without reviving CLOSED campaign. Restart after latch retains block and can seed the admitted exit. Failures do not claim fill refusal or delete reservations.
- Existing valid exit continues when admission evidence is unreadable or latch persistence fails. Admitted stop can raise the evaluation baseline as a floor while historical entry/initial-risk/policy remain unchanged; it never lowers baseline. Tests cover persistent floor after restart, breach, pending-sale deduplication, and authoritative sellable cap2 for held quantity3. Alerts describe admitted-floor enforcement as unconfirmed until compatible judgement commits.
- Final transport authority checks marker/payload, campaign block, position generation and exit binding in one DB transaction. Independent recheck refuses reproducible durable mismatch even after latch write failure. No new EntryGate or network-duration lock introduced.
- a148 numeric follow-up is limited to diagnostic caller inputs,64-character decimal grammar before shared parser. Exponent/radix/fraction/oversize rejected; signed loss remains valid and missing remains unknown. Frozen lane/parser/IDs unchanged.

## Concurrency evidence and limits

TestA150FinalTransportRefusesCommittedProtectionLatch uses explicit goroutine barriers and a second journal connection. It proves the journal check after committed latch refuses a simulated callback; a callback whose authority check already passed can proceed. SUBMITTING and RESERVED remain intact. This is not a new full Gateway race test and does not prove bytes were never sent after an earlier authorization.

Existing TestStrategyGatewayFinalGateWaitCannotUseExpiredReplacedOwnerKRUS injects final transport proof rejection into real Gateway with fake broker, asserts StateNotDispatched and broker places0. Unchanged gateway.go calls that proof before plan.call. Combined evidence supports final-check error propagation, not stronger retroactive cancellation semantics.

## Receipts inspected

Current post-generation/floor focused receipts: `/tmp/tossos-a150-journal-current.log` PASS6.109s; engine-current PASS5.909s; guardian-current PASS10.577s; historical-current PASS5.146s. Also inspected v4-atomic, historical-golden/replay, diagnostic and focused Gateway receipts and corresponding test source. Reviewer did not rerun them. `git diff --check` succeeded.

Earlier `/tmp/tossos-a150-current-packages.log` has a journal default600-second timeout; other packages passed. It is not current aggregate success. Full gate has not run at report creation and must establish repository-wide result after source freeze. AST/FLM evidence supports review structure; it is not independent branch-execution proof.

Final post-tuple-CAS focused sweeps also inspected: `/tmp/tossos-a150-journal-final.log` PASS4.080s; engine-final PASS5.867s; scope-guardian PASS1.497s; scope-partial PASS1.056s; scope-engine PASS1.386s. These are focused receipts, not aggregate gate evidence.

## Gate-discovered census follow-up

First aggregate attempt failed at engine tests because the a090 exact workingSet exit inventory still named ten exits and omitted the newly introduced post-mark protection refusal. Manager approved the narrowly scoped test update. Rereview confirms one exact AST path added, comment updated to eleven, and all original mutation checks retained. Final manager-approved naming repair restores the historical FrozenTen test identifier with an explicit comment identifying the current eleven-exit census; no assertion or behavior changed. Added ObserveOnce behavioral test creates an unrecoverable foreign-generation position and requires Unobserved1, Judged0, error and queued reconciliation alert. It does not claim external alert delivery.

`/tmp/tossos-a150-census-green.log` reports engine PASS2.877s for the focused census/mutation/behavior cases. Production source hashes still match the preceding review inventory; only two test files are changed/added to that inventory. The failed aggregate attempt remains failed evidence. Successful full rerun remains required; this focused fix is not aggregate completion.

## First aggregate repair rereview

The first aggregate finished FAIL; success is not inferred from repair tests. Reviewed the subsequent floor-reader production change and fixture/inventory fixes. Floor CAS now reads fill-owned progress through existing scanExitState/exitStateSelect in the same transaction, plus quantity from positions. All compared tuple fields and selector-before-write ordering remain unchanged. Existing apply-hook ownership guard was not relaxed. Current recordExitJudgementTx AST source SHA equals exit_state.go SHA; refreshed branch map includes the exact tuple comparison and FLM records this repair.

Schema census adds only strategy_binding_version; historical-v25 fixture copying excludes that documented additive v36 column rather than tolerating arbitrary schema differences. Eligibility-speller inventory adds only the two exact lineage/scope reader files with reasons, not wildcard exclusions or weakened predicates.

Weekly KR/US tests now load genuine test-seam WEEKLY projections, rebind account/symbol/campaign and canonical lineage/terms/v4 identities, then run production VerifyAcceptedProjection and risk-binding validation. They no longer disguise a Parker DecisionRecord by changing lane text. Both market fixtures retain explicit weekly execution policy and units. Existing negative assertions still require ErrWeeklyReservationConflict, zero partial admission rows for altered reservation digest, and rejection of non-weekly reservation smuggling; success/replay tests check actual stored reservation/week binding. No public legacy bypass introduced. These are synthetic test-seam policy fixtures, not operational lane profitability evidence.

Focused receipts inspected: gate1-journal-repair PASS5.083s, gate1-position-repair PASS0.028s, gate1-floor-repair engine PASS7.743s. Full aggregate rerun remains required. Reviewer ran no tests and changed only this report.

## Second aggregate tamper-test repair rereview

Second aggregate remains FAIL: manager reports ordinary suite and tagged engine passed, but tagged journal outer-schema tamper used the legacy v3 replacement on a v4 first-leg payload, leaving input unchanged. The reviewed repair targets v4 to v9 and requires every mutation to change the lineage before invoking admission. KR/US cases still require rejection and zero rows in all twelve listed admission/reservation/lineage/campaign/dispatch tables. Separate historical v3 canonical cases retain v3 mutations and now explicitly reject no-op test setup. No production validation or assertion was weakened; no production source hash changed against the preceding inventory.

Inspected focused receipt /tmp/tossos-a150-gate2-tamper-repair.log: journal PASS6.682s. This receipt is not full gate success. Gate 3 pending; reviewer ran no tests and modified only this report.

## Reviewed file hashes (SHA-256)

Refreshed after second aggregate tamper-test repair; previous inventories superseded.

- `internal/app/engine/a090_workingset_exits_test.go`: `391c96677d82b7007a6e803ebe6af63c122a2cdc3df9e7c5270cbe6b44c0e44c`
- `internal/app/engine/a112_owner_scope_trading_test.go`: `fabdc28424d658cc2f9e6292882ef8fdbb2d620cebed12cfbdc314c034aa5a33`
- `internal/app/engine/a112_qfinal_family_test.go`: `7c036e065f741716be210e710c433d1ddb776659983f5031b809e0310a187b57`
- `internal/app/engine/a150_exit_policy_test.go`: `a938c785556fc89f2ae7238c9a6a2fc53760a9cebd3726acc38e1fba0158716e`
- `internal/app/engine/a150_exit_protection_test.go`: `5988ac1d55f7bc2010ad5d4f0822e68776371dad1ea2b7d910a89471c1e02ea7`
- `internal/app/engine/a150_policy_fixture_test.go`: `008c69b6d1396bf35250d7a752dd9a1faa38e1b3f568bc0452a02db65dc6d3c6`
- `internal/app/engine/exitloop.go`: `ba30515476584fead62907157bb5ba985271057b2da882c6fda79e2b29951598`
- `internal/app/engine/strategy_account_first_leg_authority.go`: `d883d171c910408540843948b54de8b51372cec3b9b51d059ab15a0f28623632`
- `internal/app/engine/strategy_account_first_leg_authority_test.go`: `d796d422ef4d4c658b3019bc5b006f015feefba849cd9305e629df49288496c4`
- `internal/app/engine/strategy_dispatch_cycle_test.go`: `52e4ecf3e6fc26e0152f604f3fb78e6752347eef5ac696a66c0bf5f00c063996`
- `internal/app/engine/strategy_entry_supervisor.go`: `7a11ed1ff513843a3d050235935a42dfc16a3b0918d285f0c23b5e9d0c58fd4a`
- `internal/app/engine/strategy_first_leg_identity_backstop_test.go`: `0eb6a371653573f8112661d705c920569f88fbdaf127b297fe15a246698d7e47`
- `internal/app/engine/strategy_protection.go`: `0ce061c8635dcafb3543ec5e68540270f9cf64f7c1f0bd602150c76b6ad7c064`
- `internal/execgw/a150_protection_testseam_test.go`: `46e12750c731680ea07bf7bd79ca71abf17f239e6a39e3d60e5c616cb4925355`
- `internal/execgw/riskguardian_account_base_testseam_test.go`: `39cf0defa13475c3c16c232e515242a432a0ca5c2b6425c805e6f5351de05f24`
- `internal/execgw/riskguardian_first_leg.go`: `d3a5752d8bfdd437ca32eabbf0ad34d594de9f4a9afcccbff36a4254a6fc36ec`
- `internal/journal/a150_first_leg_test.go`: `a3b2f1e6434736a20a6f2bea5f90fa4487f5859578bfaddbd4b6afb8fb8c5924`
- `internal/journal/a150_fixture_test.go`: `ea0d2da37afe8a1e00cc0835a774c67826431c890c0894f3cb71a204f7f88aba`
- `internal/journal/a150_marker_migration_test.go`: `220e19e4b2c30d51a588b3a9f09b8a6133c812650c5aa7dd69308b2795716529`
- `internal/journal/a150_protection_read_test.go`: `e65dfcdf89b8a249b588ad9a96459e53ba4ca3b3e934da7cfae38b648aefc9a9`
- `internal/journal/a150_strategy_projection_test.go`: `d31ede78b2d67f2dd8b9e0c2d0696a613706bfbfe1888fd680bfe7664e74279c`
- `internal/journal/exit_snapshot.go`: `19ebe0eeb9879bcbe16664a39d2ed1785c8cadb55dc97bc3b3eb1455980ff5ce`
- `internal/journal/exit_state.go`: `ea9f939d35d3bfa5d6dacac241496dc89e67f62037f89c2b95a299e4b3d47a34`
- `internal/journal/schema.go`: `602b8ec0587ab6f01f4edea222657107b6685a63c9b56db94d67162a059b6bde`
- `internal/journal/schema_test.go`: `c02f8d21546ec7c18d002f27605ee1c506cb07519d7355ea518ccbef901a5e64`
- `internal/journal/strategy_dispatch_runtime.go`: `2b82ef970b73ae626bc54a1da851280cbb0a23a809e729efc47ae0cc0d4bfe52`
- `internal/journal/strategy_dispatch_runtime_test.go`: `b26f55de7f355147cc3f3b5df6fa3035b99e444109f0a1c0f5c6fae968f62e79`
- `internal/journal/strategy_first_leg_atomic.go`: `8516689f07ccf901deecc6cd90c3b9bb8bb4ca7a18684f9d5294d1a73087eea2`
- `internal/journal/strategy_first_leg_atomic_test.go`: `604dd43d6685ab68db5a88a4db777893a63482471591019ea61d841536ca59b0`
- `internal/journal/strategy_lineage.go`: `5e1cbd1d9aa0bf9b899981bca883fadd90b0a07c06373c8a9ffb5eacd38df116`
- `internal/journal/strategy_protection_policy.go`: `13c0cc8c982be8718dc2ab15c21670c11e40935eacddde72ca24aea67d62ea29`
- `internal/journal/strategy_protection_read.go`: `b2a3b9ab6c217ecae0b1341da472b01845f5311db2876de519654fe573029e7c`
- `internal/journal/strategy_protection_reconcile.go`: `57a172e322e569c691ac6354910311af19250e646f4d74f0bae9c30911c36be4`
- `internal/journal/strategy_protection_v36.go`: `cc37883c98ca1c80e7ba7618c3ed0b22a51139efabad1c54f61f7f0dea636956`
- `internal/journal/strategy_weekly_first_leg_test.go`: `2fa64960bf80aae576e9b61162028cca8816edd90a30bc9ee970f2cd8399124c`
- `internal/journal/strategyflow_projection.go`: `8748f7be2b9f6f260f382b27b5e567d269e42395d1491cb919a5205fc3ca03e3`
- `internal/journal/strategyflow_projection_test.go`: `85127245be19a62ddd1a5f85358b22e9af0214b840ba93c5064c62a35169d78e`
- `internal/journal/testdata/a150-baseline-v2.json`: `f73fdfa9b3abf68e2cf275bc90af495d12f36639361b61b29c3f6a20f018b664`
- `internal/journal/testdata/a150-baseline-v3.json`: `4856d8b98518c41b7680118d0ca68185801e68f0e12c9a87b14aa6618c3966a2`
- `internal/journal/testdata/a150-first-leg-v35.db.gz`: `190d0ed5a89a62110ff70b52232b3a3ae490d3678cbfb87332f7b7c63ac04776`
- `internal/journal/testdata/a150-projection-KR.json`: `a351f21d0b8701b439f6d8df3ac2a9bc410e30e41e49fd52087d86e0c459e485`
- `internal/journal/testdata/a150-projection-US.json`: `0041bd22437e27f82f2e272d8b2c2aff019fe0c28ddce6f56093f77e05cfe9e4`
- `internal/journal/testdata/a150-weekly-projection-KR.json`: `df669b62adcfddfcfa75976d1a14e112e0bfbfc32977a91f42012d5e32d5b7ac`
- `internal/journal/testdata/a150-weekly-projection-US.json`: `6ef18422109f54a1792ba579a806f6e8246933a3fb113d08d69cc2aef92d925d`
- `internal/position/eligibility_drift_test.go`: `a3bc229f60cec0cd7267a40b3e560871e51bdd55c641dd3a4f6a5e27cc489447`
- `internal/positioncampaign/a150_replay_test.go`: `31b542cb06ba994ee26ef013d35935770ff66f3650dec419bd776c9b77ffd604`
- `internal/positioncampaign/command.go`: `b8fdc7bed454fe5c84286648019cbc50b0ac7d6f115c34022c58584aff4434e4`
- `internal/positioncampaign/replay.go`: `7f71c0d45024aacff66358acd9ae389f32f4742466005e4b815ccaa3490381db`
- `internal/strategyengine/rr_diagnostic.go`: `c591dc3f376f339b778c6911ed809592ac75e091f16d1ba32df6e1d133631546`
- `internal/strategyengine/rr_diagnostic_bounds_test.go`: `7b441f7a73aae11a77a3932673b56657fe1ea97db8bf62396d9fc7aaad72b8f1`
