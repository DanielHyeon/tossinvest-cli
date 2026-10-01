VERDICT: FAIL

1. **P0 — D5 does not establish §0.3 safety for the complete added path.**  
   `openspec/changes/a091-a-stop-that-sold-nothing-is-critical/design.md:78–84` calls the added cost “outbox 기록 트랜잭션 하나” and argues that existing broker reads are much larger. On recording failure, however, `internal/obs/record_only.go:136–157` additionally logs, latches the gate, unlocks `n.mu`, and **synchronously calls `escalate` in the exit goroutine**. That calls another journal transaction (`internal/obs/notifier.go:425–430`; `internal/journal/operating_mode.go:398`). Unlocking does not move execution to another goroutine.

   Contention is also workload-dependent: `Notifier.Acknowledge` holds `n.mu` while listing and acknowledging an arbitrary backlog (`internal/obs/notifier.go:957–981`). Production `claimAndDeliver` releases it before transmission (`:293–353`), but local work still has no deadline. Later positions are processed sequentially and can lose quote freshness while waiting (`internal/app/engine/exitloop.go:477–495`).

   **Minimal fix:** budget and validate logging, lock contention, recording, and recording-failure escalation separately. Include a later protective position in the test. Supply the named allocation required by `openspec/specs/exit-policy/spec.md:68–72`; unchanged return values and a comparison with broker latency are insufficient.

2. **P0 — Newly reached recording-failure paths expose raw account numbers.**  
   D3 preserves B2 `logErr` (`design.md:53–55`), and D5 routes protective zeros through recording. `ExitObserver.logErr` writes `o.opts.AccountRef` directly (`internal/app/engine/exitloop.go:1834–1838`). Recording failure also reaches `Notifier.escalate`, whose failure and success logs both write raw `n.AccountRef` (`internal/obs/notifier.go:432–444`). The production account reference is the broker account number, not an opaque identifier (`internal/app/engine/interlock.go:680–684`). The logger does not redact these attributes (`internal/obs/log.go:113–117,190–199`).

   Thus a tail zero, previously normal, gains a failure path that emits the account number. `withoutFields` does not protect these separate logs.

   **Minimal fix:** remove or mask account attributes and account-bearing errors on every reachable reporting path; add sentinel-account tests for B2, failed recording, and failed/successful escalation.

3. **P1 — Alerts-OFF behavior is unresolved and changes upstream entry behavior.**  
   The delta requires protective zeros to be critical unconditionally (`specs/engine-safety/spec.md:25–30`). Disabled notifications resolve to a nil publisher (`internal/app/engine/notifications.go:83–87`), but the notifier is still constructed (`internal/app/engine/gateway.go:327`) and injected into the exit observer (`internal/app/engine/exitwiring.go:333–352`).

   The delivery executor counts a missing publisher as a failed attempt (`internal/app/engine/alertdelivery.go:316–338`). At three failures it latches entries and attempts `ENTRY_BLOCKED` escalation (`:98,379–382,446–471`; `internal/obs/notifier.go:45`). Recording failure triggers enforcement immediately instead.

   a095 explicitly guards its critical classification with loaded `NotificationsEnabled` (`internal/app/engine/adoption.go:449–453`); its canonical rationale is at `openspec/specs/engine-safety/spec.md:1846–1851`. That requirement specifically covers unmanaged reports, so it does **not automatically exempt protective zeros**. Nevertheless, a091’s claimed analogy (`design.md:128–131`) omits this consequential difference.

   **Minimal fix:** explicitly resolve OFF behavior in the proposal, delta, wiring, and RED matrix. Distinguish **disabled notifications** from **enabled notifications lacking transport**. The current unconditional design cannot claim OFF-equivalence.

4. **P1 — RED tasks do not fully enforce the new reporting contract.**  
   Tasks 3.1–3.7 and 4.1 (`tasks.md:47–63`) leave these holes:

   - Tail-zero cause detail is not explicitly asserted; only B2 gets that assertion.
   - Task 4.1 forbids one literal phrase, rather than asserting a truthful zero-submission result. Another false partial-sale sentence would pass.
   - Generic “protective” and “take-profit” cases do not require coverage of every orderable action or the actual `submit` classification.
   - Return-value assertions in 3.5 do not test timing, delivery, or enforcement.

   The current split itself is correct: two protective actions and three profit actions (`internal/exitpolicy/ratchet.go:94–118`; `internal/app/engine/exitloop.go:1375–1377`). Recovery introduces no sixth action: snapshots validate the action set and derivation (`internal/exitpolicy/recovery.go:223–236`; `internal/journal/exit_snapshot.go:170–175`); selecting the saved snapshot suppresses arming (`internal/journal/exit_state.go:531–546`).

   Also, the cited existing harness **only appends events** (`internal/app/engine/exitloop_test.go:122–126,236`); it cannot demonstrate durability.

   **Minimal fix:** specify all five actions through `submit`, relevant recovery cases, both causes, truthful zero wording, and unchanged submitted quantities. Add an integration fixture using real `RecordOnly`, journal, logger, and gate. Require these assertions to reject targeted wrong implementations.

5. **P2 — Shared episode keys need an explicit cause-retention contract.**  
   The SHALL says “원인은 세부 정보로 구분해 담아야 한다” (`specs/engine-safety/spec.md:28`), while D5 deliberately shares `exit.stop_sold_nothing|<position id>` between B2 and tail and preserves the first PENDING content (`design.md:86–91`).

   Code confirms that PENDING means `owed=true, rearm=false` (`internal/journal/outbox.go:379–381`); content updates occur only during rearming (`:294–345`). A tail-zero followed by B2 therefore delivers the original tail cause. The reverse sequence retains B2.

   **The SHALL is satisfiable as first-episode alert detail plus per-observation log detail; it is not satisfiable as every observation’s cause being retained in the shared outbox row.** The delta does not choose between those meanings. Moreover, `RecordOnly` removes fields before logging, and `logEvent` does not automatically log the episode key (`internal/obs/record_only.go:51,106–108`; `internal/obs/notifier.go:160–175`).

   **Minimal fix:** specify first-cause retention explicitly, require subsequent causes and position identity in safe log content, and test both cause orders while PENDING. Preserve a097’s settlement semantics.

6. **P2 — Replay 5.1/5.2 cannot prove the stated outcome without delivery-state controls.**  
   D5 concludes “행 1 · PENDING 1” for thirteen observations over three minutes (`design.md:89`), and task 5.1 expects that replay to exercise rearming (`tasks.md:68–73`).

   One key explains one row, **not necessarily PENDING**: successful delivery settles it, and subsequent observations inside the one-hour window leave it settled (`internal/journal/outbox.go:382–410`; `internal/obs/notifier.go:59`). Conversely, a continuously PENDING row never exercises the settled-row reminder boundary. Removing the old synchronous H3 failure does not imply no delivery errors: the current nil-publisher executor deliberately emits them (`internal/app/engine/alertdelivery.go:316–325`).

   **Minimal fix:** separate record-only deduplication, successful delivery, nil/failing transport, and delivered/acknowledged rows before and after the reminder boundary. Assert attempts, state, gate, mode, and specific error causes.

7. **P2 — “Worst ≈98s / 196s” is not a verified wall-clock bound.**  
   The calculation in `analysis/harness/write_bundles.py:28–33`, cited by `design.md:83–84`, omits unbounded mutex and filesystem work. Token acquisition and refresh hold `m.mu` (`internal/official/token.go:61–78,109–125`) across cache access and exchange; exchange also saves the cache (`:172,178–179,204–223`).

   Its HTTP arithmetic is also not an attainable six-request sequence: the second refresh happens only if the first **adopted** a cached token; an exchange on the first refresh breaks the loop (`internal/official/client.go:344–360`). The verified constants are three query attempts, an eight-second retry budget, 400/800ms default backoff with 25% jitter, and a 15-second default HTTP timeout (`internal/execgw/retry.go:129–137,344–385`; `internal/official/client.go:20,131`).

   **Minimal fix:** label any HTTP-only estimate with its assumptions, enumerate realizable request sequences, and exclude it as evidence that an unmeasured local transaction is “orders of magnitude” smaller.

8. **P2 — Several rebased task premises remain false or contradictory.**  
   `tasks.md:79` still says “`applyFloor`는 브로커에 닿지 않는다.” In production it calls `ConfirmedFloor`, which invokes Holdings and, if that succeeds, SellableQuantity (`internal/app/engine/exitloop.go:1621`; `internal/app/engine/exitwiring.go:207–240`). The harness cites these calls at `:204` and `:227`, which are not their call sites (`analysis/harness/write_bundles.py:27`). Task 3.7 also says to change B2’s kind without the protective qualifier, contradicting task 3.4a’s take-profit preservation (`tasks.md:51,56–58`).

   The principal refreshed anchors **do check out**: 19 critical kinds (`internal/obs/event.go:337–360`), the `applyFloor` coordinates and two successful zero-return paths (`internal/app/engine/exitloop.go:1617–1661`), numerical zero detection (`:1871–1877`), and whole-share projection (`internal/exitpolicy/snapshot.go:73–93,138–144,203–209`). These do not repair the stale tasks.

   **Minimal fix:** change 6.3 to “a091 adds zero broker requests,” correct call coordinates, qualify 3.7 as protective-only, and replace 6.2’s “timing unchanged by diff” premise with the complete §0.3 verification in finding 1.