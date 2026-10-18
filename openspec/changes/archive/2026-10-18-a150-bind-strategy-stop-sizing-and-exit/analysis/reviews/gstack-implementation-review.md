# a150 gstack implementation review

Verdict: PASS for the reviewed implementation and focused evidence below. First full repository gate failed on census, fixture/inventory and structural-read checks; Gate 2 also failed on a no-op tamper fixture; Gate 3 remains pending; this is not release or live-runtime approval.

Reviewer: independent Codex-host reviewer applying gstack review's critical/informational passes and security/QA checklist. No Claude or cross-model execution is claimed. Reviewer ran no tests or gates, made no production/test edits, and performed no live/configuration operations. Base: `28a800f5ed5eed7c0db0aed4f435d0778b034191`.

## Findings resolved

- **P1 prefill generation drift:** final authority formerly accepted a missing expected position or a later CLOSED generation. Current code compares expected generation/version with the latest CLOSED row; only expected 0/0 permits absence. RED/GREEN receipts and negative source fixtures inspected.
- **P1 admitted stop weakened by existing valid exit:** retaining an existing initial stop 90 when admitted stop was 95 left price 94 unprotected. Same-generation authenticated admission now raises evaluation baseline to max(existing baseline, admitted stop), preserving initial stop/R, entry, policy and pending provenance. Tighter protection remains tighter. Latch failure does not prevent evaluation.
- **P1 stale floor recovery could quarantine valid protection:** crossed stale/new snapshots reproduced selector quarantine. Floor candidates now carry an observed-input CAS, checked transactionally before selection: effective snapshot ID, baseline, high-water, level/rung, quantity, taken ratio, pending action/level/intent, with existing lifecycle/generation guards retained. Selector refusal for these candidates returns conflict without quarantine or mutation. Engine records explicit floor-reconciliation arm suppression before any order-clear/submit call. Next actual observation uses the durable floor and normal protective path. This intentionally delays a newly raised floor's order until the next observation; no fabricated observation, same-observation re-arm, selector bypass or network lock.
- Foreign generation reuse identified by the separate reviewer was also checked: sealed projection generation must equal expected+1 and actual position generation. A different-generation admission cannot seed protection.

- **P1 seed-window race (implementer follow-up):** snapshot-ID-only CAS missed legacy scalar changes when both snapshot IDs were empty. RED receipt reproduced acceptance of a stale seed. Revised floor-only CAS compares the observed evaluator inputs listed above inside the same write transaction; mismatch returns before selector, mutation or transport. Legacy watermark, simulated fill and pending-change negatives were inspected. This supersedes the earlier snapshot-ID-only approval.

No unresolved actionable code/security finding remains in reviewed scope.

## Review coverage and conclusions

Read complete tracked source diff, new source/test files, historical fixtures, approved design/spec deltas, and relevant existing callers. Checked database atomicity, rollback, canonical bytes and hashes, missing evidence, scope, restart/replay, concurrency, trust-boundary numeric validation, and test assertions rather than test names alone.

- Schema 36 adds independent immutable decision binding marker (legacy 0/new 4). New decision marker, risk binding, campaign, reservation and admission persist atomically. Missing marked binding fails closed. Historical v2/v3 canonical identities and RiskIntent remain unchanged. Reader rollback requires a schema-36/v4-aware binary; older schema readers intentionally reject the migrated database.
- Admission copies and validates the pinned policy snapshot. Current runtime configuration cannot replace its exit seed. Legacy/adopted behavior and unsupported lanes remain bounded by existing paths.
- Reconciliation validates account, campaign, decision, symbol and generation, never resurrects CLOSED positions, and durably latches entry blocking with idempotent replay. Inspected subsequent write/replay paths preserve the latch. Valid existing exits continue; valid admission can seed a missing exit after latch/restart.
- Final transport authority checks durable scope, marker/binding, campaign latch, projection/position generation, and exit in one read snapshot. Read errors fail closed. Durable binding conflicts still refuse when latch write fails. Volatile-only mismatches cannot be guaranteed durable if persistence fails; alarms/retry remain necessary.
- Authorization linearizes at final authority's consistent read. Already authorized work can reach the broker after a later latch and retains SUBMITTING/reservation uncertainty. No physical-byte cutoff or retroactive cancellation is claimed.
- Journal barrier test uses simulated transport callbacks and a second connection. Separately, existing Gateway call chain invokes final authority before send tracking/plan.call; KR/US fake-broker test verifies final-proof error yields zero broker places. These are complementary evidence, not one end-to-end race test.
- Floor recovery preserves prior state on conflict, persisted baseline survives restart, repeated real observations retry protective execution, pending order is not duplicated, and sellable quantity cap remains applied. Existing selector, journal validation and policy identity checks remain active.
- RR diagnostic leaf grammar is bounded before rational parsing, rejecting p/P, exponent/hex/underscore/fraction and excessive length without changing the shared lane parser. Missing evidence remains distinct from zero or computed values. No new untrusted dynamic SQL, secret handling or external transport surface was introduced.

## Evidence inspected (executed by implementer, not reviewer)

- `/tmp/tossos-a150-journal-current.log`: tagged journal `^TestA150`, PASS.
- `/tmp/tossos-a150-engine-current.log`: tagged engine `^TestA150`, PASS before final floor-CAS extension; final floor extension separately covered below.
- `/tmp/tossos-a150-guardian-current.log`: tagged QFinal campaign/account/FX and Gateway final-authority KR/US tests, PASS.
- `/tmp/tossos-a150-historical-current.log`: historical migration replay and first-leg atomic admission, PASS.
- `/tmp/tossos-a150-focused-leaf.log`: RR diagnostic and protection-reconcile tests, both packages PASS.
- `/tmp/tossos-a150-generation-green3.log`: generation/projection/prefill negatives, PASS.
- `/tmp/tossos-a150-transport-barrier.log`: journal final-authority latch barrier, PASS; no race-detector execution inferred.
- `/tmp/tossos-a150-floor-recovery.log`: initial floor/restart/pending fixtures, PASS.
- `/tmp/tossos-a150-stale-floor-red.log`: expected reproduced quarantine failure; `/tmp/tossos-a150-stale-floor-green.log`: fix PASS.
- `/tmp/tossos-a150-floor-final.log`: final floor tests include an existing pending sell, restart immediately after suppressed breach, stale CAS, same-price real-observation retry, and sellable cap; PASS. Distinct floor-reconciliation outcome inspected; engine authorizes only `ExitArmArmed`.
- `/tmp/tossos-a150-two-stage-floor.log`: explicit floor suppression, no precommit clearing, next actual observation and stale recovery tests, PASS.

- `/tmp/tossos-a150-seed-cas-red.log`: expected stale-seed acceptance failure. `/tmp/tossos-a150-seed-cas-green.log`: revised seed-input CAS plus floor/restart/stale tests PASS. Fill/pending changes are simulated database state transitions, not end-to-end fill transport tests.

- Final current-source `^TestA150` journal and engine sweeps: `/tmp/tossos-a150-journal-final.log` and `/tmp/tossos-a150-engine-final.log`, both PASS. Additional existing scope receipts `/tmp/tossos-a150-scope-guardian.log`, `/tmp/tossos-a150-scope-partial.log`, `/tmp/tossos-a150-scope-engine.log` also show PASS; no broader coverage inferred.

Earlier broad package timeout and superseded Guardian fixture failure are not passing evidence. AST/branch inventories support review navigation; they do not establish exhaustive branch coverage. Full-gate completion, final SDD synchronization and live execution remain outside this review receipt. The first gate failure remains recorded; focused census repair does not establish a successful full gate.

## Gate-discovered test-only correction

Reviewed a090_workingset_exits_test.go diff: adds exactly the a150 post-markHeld protection refusal path, updates ten to eleven and the test name; AST mutation checks remain intact. Added ObserveOnce behavior fixture confirms foreign-generation unrecoverable position stays Unobserved=1, Judged=0, returns an error and queues one reconciliation alert. This asserts alert queuing, not external delivery. `/tmp/tossos-a150-census-green.log` shows focused PASS. No production source changed relative to prior manifest. First full gate failed on this census; full rerun success is not claimed.

## First-gate remaining repairs rereview

PASS for these repairs. Production delta since prior review is limited to floor CAS reading through existing `scanExitState`/`exitStateSelect` inside the same write transaction, plus quantity read. All previously compared values, mismatch rejection before selector, lifecycle/generation checks and rollback semantics remain. Guarded-column test itself is unchanged; no write-authority exception was added.

Schema census explicitly adds the v36 marker. Historical v25 scratch migration allows only that known additive column alongside previous v34 additions. Eligibility inventory adds two exact journal lineage readers with reasons; source review confirms these verify entry lineage/reconcile scope rather than implement a competing exit-eligibility predicate.

Weekly fixtures now use complete KR/US WEEKLY accepted projections and rebuild canonical v4 binding (including policy, payload digest, lineage and production binding verification). Read both new JSON fixtures and helper. Existing positive paired admission/replay and negatives for mismatched weekly reservation and non-weekly smuggling remain, including exact error and zero persisted rows assertions. These are test-seam fixtures, not live strategy evidence; fixture origin is implementer-reported and the canonical reader verification is visible in test source.

Read passing receipts `/tmp/tossos-a150-gate1-journal-repair.log`, `/tmp/tossos-a150-gate1-position-repair.log`, `/tmp/tossos-a150-gate1-floor-repair.log`, plus census receipt. First aggregate failure remains recorded; no successful rerun inferred.

## Gate 2 test-only repair

PASS for the reviewed repair. The v4 first-leg fixture now mutates its actual v4 schema to v9. Every lineage mutation asserts it changed input before asserting rejection, avoiding a false-green/no-op fixture. Separate canonical v3 mutations retain v3 and now assert changed payload. All twelve zero-row table checks remain, including decisions/reservations/campaigns/bindings/dispatch leases. Read `/tmp/tossos-a150-gate2-tamper-repair.log`: seven tagged tests in the file reported PASS. Only strategyflow_projection_test.go changed relative to prior manifest; all production hashes unchanged. Gate 2 failure remains recorded; Gate 3 success is not claimed.

## Reviewed snapshot

Gate 2 test-only repair supersedes `7aecdb80545420a5e5ad3da75bb817a46583b43f59371275e8929be9fe696cff`. Manifest covers source, tests, fixtures and a150 design/spec/task documents; analysis/report excluded. Later edits require refresh.

Manifest SHA-256: `7cfce5d7ff66b106743874f9e91d78781249c64abf2381f0b49fb4f5dfbc1892`

```text
391c96677d82b7007a6e803ebe6af63c122a2cdc3df9e7c5270cbe6b44c0e44c  internal/app/engine/a090_workingset_exits_test.go
fabdc28424d658cc2f9e6292882ef8fdbb2d620cebed12cfbdc314c034aa5a33  internal/app/engine/a112_owner_scope_trading_test.go
7c036e065f741716be210e710c433d1ddb776659983f5031b809e0310a187b57  internal/app/engine/a112_qfinal_family_test.go
a938c785556fc89f2ae7238c9a6a2fc53760a9cebd3726acc38e1fba0158716e  internal/app/engine/a150_exit_policy_test.go
5988ac1d55f7bc2010ad5d4f0822e68776371dad1ea2b7d910a89471c1e02ea7  internal/app/engine/a150_exit_protection_test.go
008c69b6d1396bf35250d7a752dd9a1faa38e1b3f568bc0452a02db65dc6d3c6  internal/app/engine/a150_policy_fixture_test.go
ba30515476584fead62907157bb5ba985271057b2da882c6fda79e2b29951598  internal/app/engine/exitloop.go
d883d171c910408540843948b54de8b51372cec3b9b51d059ab15a0f28623632  internal/app/engine/strategy_account_first_leg_authority.go
d796d422ef4d4c658b3019bc5b006f015feefba849cd9305e629df49288496c4  internal/app/engine/strategy_account_first_leg_authority_test.go
52e4ecf3e6fc26e0152f604f3fb78e6752347eef5ac696a66c0bf5f00c063996  internal/app/engine/strategy_dispatch_cycle_test.go
7a11ed1ff513843a3d050235935a42dfc16a3b0918d285f0c23b5e9d0c58fd4a  internal/app/engine/strategy_entry_supervisor.go
0eb6a371653573f8112661d705c920569f88fbdaf127b297fe15a246698d7e47  internal/app/engine/strategy_first_leg_identity_backstop_test.go
0ce061c8635dcafb3543ec5e68540270f9cf64f7c1f0bd602150c76b6ad7c064  internal/app/engine/strategy_protection.go
46e12750c731680ea07bf7bd79ca71abf17f239e6a39e3d60e5c616cb4925355  internal/execgw/a150_protection_testseam_test.go
39cf0defa13475c3c16c232e515242a432a0ca5c2b6425c805e6f5351de05f24  internal/execgw/riskguardian_account_base_testseam_test.go
d3a5752d8bfdd437ca32eabbf0ad34d594de9f4a9afcccbff36a4254a6fc36ec  internal/execgw/riskguardian_first_leg.go
a3b2f1e6434736a20a6f2bea5f90fa4487f5859578bfaddbd4b6afb8fb8c5924  internal/journal/a150_first_leg_test.go
ea0d2da37afe8a1e00cc0835a774c67826431c890c0894f3cb71a204f7f88aba  internal/journal/a150_fixture_test.go
220e19e4b2c30d51a588b3a9f09b8a6133c812650c5aa7dd69308b2795716529  internal/journal/a150_marker_migration_test.go
e65dfcdf89b8a249b588ad9a96459e53ba4ca3b3e934da7cfae38b648aefc9a9  internal/journal/a150_protection_read_test.go
d31ede78b2d67f2dd8b9e0c2d0696a613706bfbfe1888fd680bfe7664e74279c  internal/journal/a150_strategy_projection_test.go
19ebe0eeb9879bcbe16664a39d2ed1785c8cadb55dc97bc3b3eb1455980ff5ce  internal/journal/exit_snapshot.go
ea9f939d35d3bfa5d6dacac241496dc89e67f62037f89c2b95a299e4b3d47a34  internal/journal/exit_state.go
602b8ec0587ab6f01f4edea222657107b6685a63c9b56db94d67162a059b6bde  internal/journal/schema.go
c02f8d21546ec7c18d002f27605ee1c506cb07519d7355ea518ccbef901a5e64  internal/journal/schema_test.go
2b82ef970b73ae626bc54a1da851280cbb0a23a809e729efc47ae0cc0d4bfe52  internal/journal/strategy_dispatch_runtime.go
b26f55de7f355147cc3f3b5df6fa3035b99e444109f0a1c0f5c6fae968f62e79  internal/journal/strategy_dispatch_runtime_test.go
8516689f07ccf901deecc6cd90c3b9bb8bb4ca7a18684f9d5294d1a73087eea2  internal/journal/strategy_first_leg_atomic.go
604dd43d6685ab68db5a88a4db777893a63482471591019ea61d841536ca59b0  internal/journal/strategy_first_leg_atomic_test.go
5e1cbd1d9aa0bf9b899981bca883fadd90b0a07c06373c8a9ffb5eacd38df116  internal/journal/strategy_lineage.go
13c0cc8c982be8718dc2ab15c21670c11e40935eacddde72ca24aea67d62ea29  internal/journal/strategy_protection_policy.go
b2a3b9ab6c217ecae0b1341da472b01845f5311db2876de519654fe573029e7c  internal/journal/strategy_protection_read.go
57a172e322e569c691ac6354910311af19250e646f4d74f0bae9c30911c36be4  internal/journal/strategy_protection_reconcile.go
cc37883c98ca1c80e7ba7618c3ed0b22a51139efabad1c54f61f7f0dea636956  internal/journal/strategy_protection_v36.go
2fa64960bf80aae576e9b61162028cca8816edd90a30bc9ee970f2cd8399124c  internal/journal/strategy_weekly_first_leg_test.go
8748f7be2b9f6f260f382b27b5e567d269e42395d1491cb919a5205fc3ca03e3  internal/journal/strategyflow_projection.go
85127245be19a62ddd1a5f85358b22e9af0214b840ba93c5064c62a35169d78e  internal/journal/strategyflow_projection_test.go
f73fdfa9b3abf68e2cf275bc90af495d12f36639361b61b29c3f6a20f018b664  internal/journal/testdata/a150-baseline-v2.json
4856d8b98518c41b7680118d0ca68185801e68f0e12c9a87b14aa6618c3966a2  internal/journal/testdata/a150-baseline-v3.json
190d0ed5a89a62110ff70b52232b3a3ae490d3678cbfb87332f7b7c63ac04776  internal/journal/testdata/a150-first-leg-v35.db.gz
a351f21d0b8701b439f6d8df3ac2a9bc410e30e41e49fd52087d86e0c459e485  internal/journal/testdata/a150-projection-KR.json
0041bd22437e27f82f2e272d8b2c2aff019fe0c28ddce6f56093f77e05cfe9e4  internal/journal/testdata/a150-projection-US.json
df669b62adcfddfcfa75976d1a14e112e0bfbfc32977a91f42012d5e32d5b7ac  internal/journal/testdata/a150-weekly-projection-KR.json
6ef18422109f54a1792ba579a806f6e8246933a3fb113d08d69cc2aef92d925d  internal/journal/testdata/a150-weekly-projection-US.json
a3bc229f60cec0cd7267a40b3e560871e51bdd55c641dd3a4f6a5e27cc489447  internal/position/eligibility_drift_test.go
31b542cb06ba994ee26ef013d35935770ff66f3650dec419bd776c9b77ffd604  internal/positioncampaign/a150_replay_test.go
b8fdc7bed454fe5c84286648019cbc50b0ac7d6f115c34022c58584aff4434e4  internal/positioncampaign/command.go
7f71c0d45024aacff66358acd9ae389f32f4742466005e4b815ccaa3490381db  internal/positioncampaign/replay.go
c591dc3f376f339b778c6911ed809592ac75e091f16d1ba32df6e1d133631546  internal/strategyengine/rr_diagnostic.go
7b441f7a73aae11a77a3932673b56657fe1ea97db8bf62396d9fc7aaad72b8f1  internal/strategyengine/rr_diagnostic_bounds_test.go
1b21106f0105a2cb47bc0ce2cb464933c6866c1a1bc7105a83a45611f1371d27  openspec/changes/a150-bind-strategy-stop-sizing-and-exit/base-commit.txt
6fa126acb1a86c9b6bc7cdae80a6b7cccb134196eeca0370f9b98d11e115ddd0  openspec/changes/a150-bind-strategy-stop-sizing-and-exit/design.md
8780966ff247d18be63f9d6055e7d6b5a6790b46178118ef68052e367b8743c7  openspec/changes/a150-bind-strategy-stop-sizing-and-exit/proposal.md
5f1abfab4eace0ac0aaf0144bb09f8ea31af1ab2a542af7929a094cbc25f282d  openspec/changes/a150-bind-strategy-stop-sizing-and-exit/specs/risk-management/spec.md
3cf4165501578e22744c7333bb2cf5da5748ac032b7b76f7d6b76dbf8c92bcc2  openspec/changes/a150-bind-strategy-stop-sizing-and-exit/specs/strategy-engine/spec.md
07335f4b4cbcdc114d01d6a3dd9bc3ddc42d39922bb93c343a524d5993ff7309  openspec/changes/a150-bind-strategy-stop-sizing-and-exit/tasks.md
```
