**REJECT: I3 remains partially unresolved at P1.** Cancellation handling is fixed, but several newly added adversarial tests still pass without establishing their claimed schedule or checking both required consequences.

Read-only inspection only. No tests, network calls, engine execution, or file modifications. Mutation results below are supplied evidence, not independently reproduced.

References: `J` = `internal/app/engine/a124_the_deliverer_judges_internal_test.go`; `L` = `internal/app/engine/a124_judging_does_not_delay_protection_test.go`; `G` = `internal/execgw/a124_clear_epoch_internal_test.go`; `S` = `internal/app/engine/a124_the_executor_holds_no_stop_lock_testseams_test.go`; `C` = `internal/journal/a124_settle_reads_the_committed_attempts_test.go`. Document paths are relative to the change directory.

| Finding | Status | Reason / evidence |
|---|---|---|
| I1 | **RESOLVED** | `alertdelivery.go:370–377` always judges delivery-record errors; cancelled `BeginTx` fails at `operating_mode.go:391–394`, propagating to unconditional `Block` at `alertdelivery.go:423–424`, even when the conditional block was rejected. Normal assembly’s nonempty account/non-nil gate remain prerequisites. |
| I2 | **PARTIAL** | Threshold-minus-one seeding and timestamp assertions establish outbox/escalation overlap for the selected sample (`L:218,325–332`); calibration, sample isolation, and selection-overlap weaknesses remain below. The reduced injection and operational limit are explicitly disclosed. |
| I3 | **PARTIAL — P1 remains** | Added schedules improve coverage, but ㉪ can silently degrade to ㉩; failed acknowledgement can disappear entirely without failing its test; required mode/counter assertions remain missing (`J:974–1005`; `tasks.md:50–68`). |
| I4 | **RESOLVED** | `G:148–149` now rejects the precise stale-relatch state. M29 records a behavioral failure. This remains probabilistic stress, not the requested deterministic barrier. |
| I5 | **RESOLVED** | Successful attempts read and deferred-FK setup arm the fault (`C:209–224`); `C:254–259` distinguishes commit failure from read/setup failure, retaining state/result checks. |
| Eng F1 | **RESOLVED as disclosure** | `design.md:330–334,523–524` and `review.md:619–622` record 10 ms versus 15 ms, approximately 16 ms transaction premise, and unmeasured production fsync. This is not an operational bound proved by the test. |
| Eng F2 | **RESOLVED for selected sample** | `L:331–332` requires the escalation INSERT timestamp inside the measured exit window. |
| Eng F3 | **RESOLVED** | Obsolete cancellation exemption explicitly withdrawn at `review.md:580–582`. |
| Eng F4 | **PARTIAL** | Implementation logs once per epoch (`alertdelivery.go:464–469`), but D9 still requires an actual change/first latch (`design.md:421`). `review.md:634` reinterprets rather than reconciles that contract. |
| Eng F5 | **RESOLVED for direct reassignment** | `C:342–372` rejects production assignment statements targeting the seam. It is a syntactic guard, not protection against mutation through an alias. |
| Eng F6 | **RESOLVED** | Parse errors now fail both scans: `G:179–181,224–225`; `a124_the_enforcement_boundary_internal_test.go:54–57`. |
| Eng F7 | **PARTIAL** | Median calculation is correct (`L:313–318`), but previous executors remain running until parent-test cleanup (`L:229–239`), contaminating subsequent measurements. |

**I1 fallback:** no cancellation-specific escape remains in the production path. However, its regression test checks only `latched` (`J:924–927`). The initial conditional block already satisfies that assertion. A regression skipping escalation only when cancelled would survive this test. To prove fallback specifically, clear after epoch capture or after conditional insertion, assert an escalation attempt, and require the fallback latch with no committed mode transition.

**I3 schedule audit:**

| Schedule | Assessment |
|---|---|
| Delivered → empty acknowledgement | Present; DELIVERED state proves delivery occurred, and gate/mode assertions exist (`J:933–954`). The exact delivered schedule lacks its own on-time comparison; the earlier comparison uses acknowledgement of a pending row (`J:434–451`). |
| ㉩ / ㉪ | Clear execution is observable through the expected absent latch. But ㉪ ignores claim/delivery failure and never asserts DELIVERED (`J:975–985`): it can execute only ㉩ and pass. |
| Failed acknowledgement | **Not established adequately.** Hook merely acknowledges another row, ignores error, and never proves execution (`J:993–1005`). Removing the hook preserves the expected latch. No mode assertion, no injected acknowledgement failure, and no record-failure run to test counter preservation. |
| Listing threshold + clear, both orders | Present; missing clear would leave a latch and fail. Both gate and mode checked (`J:1009–1025`). |
| Same-ID rearm carryover | Present with claim/rearm checks and latch assertion (`J:1033–1061`); mode assertion missing. |
| Lease expiry while gate blocked | Hook arrival and real lease theft checked (`S:83–97`); final mode and actual republication not checked (`S:99–107`). |
| Other previously missing coverage | Actual close/reopen added (`J:1083–1118`); real lease theft added (`J:1128–1145`), though not rearm-specific; timeout-return stub added (`J:1068–1070`); actual token checked (`J:1150–1163`). Positive complete-list pruning and remote-publish-stall/emergency-Notify variants remain absent from these fixes. |

The **structural lock pin is an acceptable narrow substitute** for inserting an escalation call into the current epoch methods: `G:256–264` rejects such calls, and M28 exercises that rejection. It does **not** prove the broader “nothing waits under the lock” claim: it permits every selector named `Lock`/`Unlock`, regardless of receiver, and does not inspect fallback `EntryGate.Block`. Current production bodies are safe by inspection (`retry.go:532–538,559–580`); describe M28 as a structural kill, not a behavioral escalation-contention test.

The latency acceptance is no longer vacuous in the original zero-judgement sense: the chosen sample must contain an escalation write and outbox writes during the exit window. Nevertheless, only that sample is validated; actual injected duration is discarded, and the scaling test does not check overlap. Median-of-three supports a typical-case measurement, not a maximum-delay guarantee.

| ID | Severity | Finding introduced by / remaining in fixes | Evidence | Suggested fix |
|---|---|---|---|---|
| R1 | **P1** | Newly added adversarial tests can pass without their distinguishing event; I3 closure is unsupported. | `J:975–1005`; required dual assertions at `tasks.md:50–60`. | Check every arrangement result and hook execution; assert DELIVERED for ㉪; inject failed acknowledgement; verify gate, mode, and counter preservation. |
| R2 | **P2** | Cancellation regression does not prove unconditional fallback or escalation execution. | `J:918–927`; initial block at `alertdelivery.go:416`. | Force an intervening clear and assert escalation attempt, failed persistence, and fallback latch. |
| R3 | **P2** | Measurement samples retain background executors; calibrated delay is logged but unenforced; scaling acceptance can have no overlap. | `L:110–112,213,229–239,315,388–394`. | Stop/join each executor before returning; validate measured injection tolerance and overlap for every sample used in acceptance. |
| R4 | **P2** | Deployment recipe does not ensure measurement uses the operational disk or measure individual transaction durations. Changing working directory does not relocate `t.TempDir()`. | `tasks.md:123–126`; `exitloop_test.go:180`; `L:260–290`. | Specify `TMPDIR` on the target filesystem and verify the actual journal path; instrument transaction durations if enforcing the ≈16 ms premise. |

**Mutation evidence:** M27–M29 are real named-test failures in the supplied ledger:

- **M27:** `TestADeliveryRecordErrorFollowedByCancellationStillLatches`.
- **M28:** `TestTheEpochMethodsCallNothingUnderTheLock`.
- **M29:** `TestTheEpochComparisonAndTheLatchAreOneStep` and the structural pin.

Their transformations are present at `analysis/harness/mutate_a124.py:152–162`. The runner rejects build/setup failures (`:176–178`) and requires green controls (`:199–206`). M29 inserts a sleep, so its structural failure alone would be insufficient; the separately recorded atomicity-test failure supplies the behavioral evidence.

Coverage is still incomplete: M27 catches skipping the entire cancelled judgement, not cancelled escalation alone; M22 checks fallback generically, not that cancellation-specific combination. No listed mutant validates the latency-harness fixes, successful ㉪ delivery, failed-acknowledgement counter preservation, or the strengthened commit-fault arming check specifically. M25/M26 cover attempts/result placement, not all I5 regressions. The mutation runner also excludes the latency tests and `TestOnlyTestsReassignTheAttemptsRead` from its selected suites (`mutate_a124.py:28–39`).

VERDICT: REJECT — I3’s P1 adversarial-evidence gap remains: several required schedules still pass without proving the event or both gate and durable-mode outcomes.