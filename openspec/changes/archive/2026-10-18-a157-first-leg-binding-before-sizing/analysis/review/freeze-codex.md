**REJECT as written.** The input substitutions are sound, but “order-only” overstates the change: it suppresses a durable account-wide safety transition. D2 also omits outcomes, and the verification plan does not cover those changes.

Read-only review. All eight supplied AST artifacts match current source hashes. No files modified; tests and mutations were not executed.

1. **P1 — D2-c weakens durable account protection; rejecting this request is not equivalent protection.**

   [design.md:63](/mnt/D/Axipient/workspace/TossOS/openspec/changes/a157-first-leg-binding-before-sizing/design.md:63) recommends accepting suppression of daily-loss escalation. Currently, [`admit():118`](/mnt/D/Axipient/workspace/TossOS/internal/app/engine/strategy_first_leg_admission.go:118) invokes escalation; [`EscalateDailyLossLimit():232`](/mnt/D/Axipient/workspace/TossOS/internal/execgw/a137_daily_loss.go:232) persists `ENTRY_BLOCKED`.

   After reordering, a binding defect prevents that transition. A market fault is weaker in scope and persistence: the production refresh worker records the error and continues ([supervisor:974](/mnt/D/Axipient/workspace/TossOS/internal/app/engine/strategy_entry_supervisor.go:974)); it does not establish the missing account-wide block. Other paths evaluating daily loss later does not prove equivalent protection now.

   **Fix:** preserve account-wide escalation through an explicitly designed safety path, or obtain explicit approval for this safety-semantic change and revise the “order-only” claim. Test persisted operating mode, subsequent cycles, and the other market—not merely zero issuance for the defective request. No newly permitted order was demonstrated, but the protective state transition is demonstrably lost.

2. **P2 — D2’s outcome enumeration is incomplete.**

   [D2(c)](/mnt/D/Axipient/workspace/TossOS/openspec/changes/a157-first-leg-binding-before-sizing/design.md:57) describes the old outcome as escalation plus an error-free cycle. That holds only when escalation succeeds and no earlier scope errors were collected.

   Escalation failure returns `errors.Join(refusal, err)` ([daily-loss:235](/mnt/D/Axipient/workspace/TossOS/internal/execgw/a137_daily_loss.go:235)), producing a fault rather than the `DailyLossEntryBlock` quiet-stop outcome. Even successful escalation preserves accumulated scope errors ([delivery:83](/mnt/D/Axipient/workspace/TossOS/internal/app/engine/strategy_market_handoff_delivery.go:83)). D2(d) also loses the specific unmeasurable-loss recording path when replaced by a binding error ([delivery:46](/mnt/D/Axipient/workspace/TossOS/internal/app/engine/strategy_market_handoff_delivery.go:46)).

   **Fix:** split escalation success/failure, distinguish per-request classification from whole-cycle result, and document changed daily-loss counters/notices. Describe production reachability: `validateStrategyFirstLegAuthority` already rejects many binding defects before precheck ([admission:206](/mnt/D/Axipient/workspace/TossOS/internal/app/engine/strategy_first_leg_admission.go:206)); malformed weekly binding is a useful reachable co-occurrence fixture.

3. **P2 — The accepted-set proof incorrectly excludes time-dependent evaluation.**

   [design.md:50](/mnt/D/Axipient/workspace/TossOS/openspec/changes/a157-first-leg-binding-before-sizing/design.md:50) says the clock is read after the predicates. `PrecheckQFinalEntry` actually reads it at lines **118, 127, 132, and 165** ([source:118](/mnt/D/Axipient/workspace/TossOS/internal/execgw/riskguardian_qfinal.go:118)). Bucket admission uses that evaluation time to reject expired evidence ([admission.go:151](/mnt/D/Axipient/workspace/TossOS/internal/riskbucket/admission.go:151)).

   Moving work before these reads can cross a freshness boundary. Thus identical acceptance and identical `issuedAt` cannot be promised unconditionally under wall-clock execution.

   **Fix:** state equivalence for identical inputs and identical clock observations. Add freshness-boundary tests and correct the transitive-call argument. Distinguish unchanged predicates from unchanged execution timing.

4. **P2 — Tests cover the motivating skip case, but not the safety-changing co-occurrences.**

   [tasks.md:16](/mnt/D/Axipient/workspace/TossOS/openspec/changes/a157-first-leg-binding-before-sizing/tasks.md:16) and [D6:96](/mnt/D/Axipient/workspace/TossOS/openspec/changes/a157-first-leg-binding-before-sizing/design.md:96) primarily exercise cap-zero plus B2/B5/B8 and isolated failures.

   **Fix:** require:
   - Binding defect plus bucket exhaustion, unmeasurable loss, and daily-loss limit—including escalation success/failure baselines.
   - Real Guardian → `admit` → delivery/supervisor coverage for reachable weekly-binding defects; assert mode, counters, notices, later-scope evaluation, and issuance.
   - Differential canonicalization cases, caller-supplied nested `Admission.QCandidate` disagreement, nil Guardian, zero execution terms, and nil weekly binding.
   - Actual constructor-to-log A2 tests with repeated events, day rollover, and nil logger.

   M13’s intentionally uncompilable move is a build rejection, not behavioral mutation evidence; report it separately.

The specific dependency and panic questions otherwise check out:

- **Market/Symbol:** entry canonicalization changes a local value copy. Existing B5 already compares against the original `request.Entry`, so moving it does not switch from normalized to raw inputs ([first-leg:71](/mnt/D/Axipient/workspace/TossOS/internal/execgw/riskguardian_first_leg.go:71), [entry:174](/mnt/D/Axipient/workspace/TossOS/internal/execgw/riskguardian_qfinal.go:174)).
- **Owner:** B3/B5 read prospective generation, lane, and campaign; they do not read the trimmed `AccountID`.
- **QCandidate:** entry overwrites nested admission quantity from `Entry.QCandidate`; `CalculateAdmission` copies it unchanged ([entry:114](/mnt/D/Axipient/workspace/TossOS/internal/execgw/riskguardian_qfinal.go:114), [calculator:66](/mnt/D/Axipient/workspace/TossOS/internal/riskbucket/admission.go:66)).
- **Panic:** the proposed leading nil-Guardian guard is necessary. With it, no additional input-induced panic was identified: price conversion bounds its scale, generation arithmetic is guarded, and weekly binding checks nil before dereferencing.
- **A2:** the two callback assignments match the production constructor. Both callback factories safely return nil without a logger. No independent wiring defect found.
