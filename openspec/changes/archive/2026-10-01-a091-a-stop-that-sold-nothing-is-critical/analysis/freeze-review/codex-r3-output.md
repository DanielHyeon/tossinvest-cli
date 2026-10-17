ACCESS-VIOLATION: The comparison command used shell process substitution; diff read /dev/fd/63 and /dev/fd/62 outside the review tree.
VERDICT: FAIL

Document references below are relative to `openspec/changes/a091-a-stop-that-sold-nothing-is-critical/`. CLOSED means addressed in the freeze contract, not implemented or tested.

| Round-2 finding | Status | Evidence |
|---|---|---|
| 1. Complete §0.3 cost | **PARTIAL** | `design.md:113–125` now names failure escalation and contention; `tasks.md:93–99` still lacks an acceptance budget and explicit failed-record/later-position verification. |
| 2. Raw account exposure | **PARTIAL** | B2 masking planned, but escalation deliberately deferred (`design.md:163–170`); returning recording errors also reach raw-account `logErr` (`internal/app/engine/exitloop.go:1818–1838`). |
| 3. Alerts-OFF consequences | **CLOSED** | Loaded-flag-only gate and enabled-with-missing-transport distinction explicit (`design.md:44–62`; `tasks.md:54–55,84–89`). Separate OFF reporting inconsistency below. |
| 4. RED coverage | **PARTIAL** | Real recorder harness, five-action matrix, truthful wording added (`tasks.md:46–78`); recovery cases and recording-failure/privacy coverage remain incomplete. |
| 5. Episode cause contract | **CLOSED** | `design.md:147–155` and delta `specs/engine-safety/spec.md:47–50` are deliverable: PENDING preserves content (`internal/journal/outbox.go:294–345,379–381`), while each Notify logs Body despite field removal (`internal/obs/record_only.go:51,106–108`; `internal/obs/notifier.go:168–169`). |
| 6. Replay premises | **CLOSED** | Four delivery/configuration arms and settled-row reminder tests replace the unconditional PENDING premise (`tasks.md:84–92`). |
| 7. HTTP “worst-case” numbers | **CLOSED** | Realizable request sequences and explicit “not a wall-clock bound” replace 98s/196s (`analysis/harness/write_bundles.py:27–34`); unsupported comparison removed (`design.md:129`). |
| 8. Stale task premises | **CLOSED** | Broker-request claim and §0.3 verification rewritten; B2 kind assertion qualified (`tasks.md:68,98–101`). |

The exit-policy block **is faithfully copied**, apart from the stated exception sentence and appended scenario: canonical `openspec/specs/exit-policy/spec.md:61–114` corresponds to delta `specs/exit-policy/spec.md:10–63`; additions are at `:11,65–67`. Its grade exception agrees with engine-safety’s enabled/protective/zero rule and exclusions. The defective predicates below concern implementing those exclusions.

1. **P1 — Holdings-zero equivalence is false without additional preconditions.**  
   Claim: `design.md:85–89`; test: `tasks.md:57–58`.

   Counterexample: valid evaluation time, fresh holdings `"0"`, local sells `"0"`, **nil sellable snapshot**. `ConfirmedFloorQuantity` returns quantity `"0"` with `FloorBoundNoSnapshot`, not Holdings (`internal/riskcalc/confirmed_floor.go:146–158,236–243`). Stale sellable similarly returns StaleSnapshot; malformed sellable or invalid local sells returns an error (`:141–158`). Thus **fresh holdings zero ⇒ Holdings-bound zero is false**. The reverse implication holds for a successful Holdings-bound result through the arithmetic path (`:164–190`).

   **Minimal fix:** state the implication accurately; qualify the converse with valid local quantity and both snapshots fresh/valid. Add missing/stale/malformed inputs to the matrix. If Q2 requires recognizing known holdings-zero through those other failures, preserve that evidence explicitly rather than infer it from `Bound`.

2. **P1 — `ctx.Err()` at B2 cannot establish cancellation caused the failure.**  
   `design.md:83,126–128` suppresses reporting whenever the caller context has ended. A genuine floor error can return, then cancellation occur before B2 checks `ctx.Err()` (`internal/app/engine/exitloop.go:1621–1622`). That real failure becomes “shutdown cancellation.” Conversely, cancellation after the check but before recording still causes the supposedly excluded latch/escalation (`internal/obs/record_only.go:137–157`). `ClassCanceled` alone is also insufficient: it includes any wrapped deadline error (`internal/execgw/retry.go:69–74`).

   **Minimal fix:** specify cancellation provenance and the race policy; suppress only an identified cancellation outcome, preserving independent failure evidence. Add tests for real-error-then-cancel and cancellation between classification and recording. The current two cases in `tasks.md:59–60` miss both races.

3. **P0 — Naming the privacy residual does not satisfy the invariant.**  
   `design.md:168–170` explicitly leaves raw-account escalation logs reachable. They remain at `internal/obs/notifier.go:433–441`. A second leak is omitted from that residual: failed recording returns through `ExitObserver.alert` to `logErr`, which adds the raw account (`internal/app/engine/exitloop.go:1818–1838`).

   **Minimal fix:** include both failure-reporting paths in the masking change and sentinel tests, or make their repair a prerequisite. Testing only B2 and successful recording (`tasks.md:69–70`) cannot close this blocker.

4. **P0 — §0.3 acceptance criteria remain incomplete.**  
   Naming two unbounded costs (`design.md:113–125`) is progress, but `tasks.md:93–94` only requests distributions; it specifies neither an acceptable allocation nor a failure disposition. The canonical rule requires a positive allocation below the observation period (`openspec/specs/exit-policy/spec.md:68–72`). The design itself admits later positions can miss their freshness window (`design.md:117–120`).

   **Minimal fix:** require an explicit allocation and acceptance gate covering successful recording, failed recording/escalation, contention, and a later protective position. Freeze need not claim measurements already exist, but must define what implementation must demonstrate.

5. **P1 — OFF behavior still contradicts its own preservation claim.**  
   `design.md:57–60` says OFF changes only wording and leaves a normal alert; `tasks.md:54` exercises both zero causes. Upstream **B2 emits no alert at all**, only a log (`internal/app/engine/exitloop.go:1622–1628`). Adding normal Notify there introduces relay/publish activity despite the claimed equivalence.

   **Minimal fix:** explicitly preserve OFF B2’s zero Notify calls; assert that count. Keep OFF tail reporting normal. This implements Q1 without changing its decision.

6. **P2 — The ledger reread still overstates what the cited evidence proves.**  
   `design.md:21–24` calls manual orders locking sellable quantity a measured conclusion and excludes local sells because they are subtracted separately. Counterexample: holdings `5`, sellable `0`, local sells `5` still produces **Sellable-bound zero**: subtraction leaves the floor equal to base, so the bound is unchanged (`internal/riskcalc/confirmed_floor.go:169–185`). Bound therefore does not identify the reserving order or actor. The listed `23:17:16 → 23:27:41` sequence also spans **10m25s**, not four minutes (`design.md:16–22`).

   **Minimal fix:** retain the documented observations, label manual-order causation as inference unless order-level evidence is supplied, and correct the duration. Q2’s decision remains unchanged.