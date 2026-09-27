1. **Yes—equivalent on every reachable production path.** In `internal/app/engine/alertdelivery.go`:

   | Path | Evidence |
   |---|---|
   | First error | Missing map entry / initial listing state yields count 0; helper returns `(1, epoch)` at :544–548. |
   | After judgement | Row entry deleted at :516; listing zeroed at :531, both before `judge`. Next error starts at 1. |
   | Below-limit epoch change | Returns `(1, current epoch)` at :553–554; remains an active run. |
   | Applied settlement | Failed-attempt settlement clears at :362; delivery settlement clears at :392; deletion implementation at :562. |
   | Claim finds settled | Clears at :302. |
   | Complete-listing prune | Guard at :249; absent entries deleted at :574–576. |
   | Successful listing | Listing counter zeroed at :248. |

   **No path stores a zero-count row entry.** The only row assignment (:513) stores an unjudged result, whose count is positive (:548, :554, :556). **The listing field does store zero**, intentionally representing absence after success/judgement—exactly where the old code also set `listSeen=false`.

   No required reset leaves stale non-zero state. Retention across truncated listings, non-Applied settlements, release, cancellation, or unobserved delivery/rearm remains intentional under `design.md:355–368`. An old epoch can remain until the next counted error; AA2 then determines judgement versus restart. Counts are bounded below the limit, so overflow cannot invalidate the invariant.

2. **AA2 and Z1 remain intact.** Threshold check at `alertdelivery.go:550` precedes epoch comparison at :553. Threshold return uses **`run.epoch`**, not newly read `epoch` (:551). Both callers pass that returned `blockEpoch` unchanged to `judge` (:517, :532), which uses it in `BlockUnlessClearedSince` (:429). Escalation remains requested for both counters, with unconditional blocking on escalation failure (:437–438).

3. **New findings: none.** No P0–P3 issue introduced within the scoped counter change.

   Supporting test source: `internal/app/engine/a124_the_deliverer_judges_internal_test.go:613` (epoch restart), :629 and :1083 (AA2/Z1), :688 (listing success), :842 (claim-settled), :1324 (Applied), :1345 (prune), :1382 (both post-judgement resets). Requested mutations are recorded as caught in `analysis/harness/mutation-ledger.tsv:21–25`, :27–29, :40–43, relative to the change directory.

   Static review only; no tests, engine, network calls, or writes performed. Fresh AST generation and execution gates were excluded by the read-only scope; ledger results were inspected, not rerun.

VERDICT: PASS — Removing `seen` preserves reachable counter states, required resets, AA2 ordering, and Z1’s previous-increment epoch.