**APPROVE-WITH-FIXES** — production logic matches decisions **1-1 / 2-1**. Two test-strength findings; no P0/P1 production defect found.

Reviewed `7b1748481..10ab9f87a`.

1. **P2 — AST test proves lexical placement, not execution order or count.**  
   [a157_first_leg_structure_test.go:36](/mnt/D/Axipient/workspace/TossOS/internal/execgw/a157_first_leg_structure_test.go:36) counts selector calls and string literals; lines 69–76 compare source positions.

   Counterexample: turn the binding IIFE into `checkBinding := func() error { … }`, invoke it before sizing, then invoke it again afterward. Every watched literal still occurs once, before sizing. The structural test accepts this; deterministic behavioral outcomes remain unchanged. Deferring its sole invocation until after sizing also evades the lexical-order check while retaining the same refusal precedence.

   **Fix:** assert the binding body belongs to an immediately invoked function in a top-level assignment preceding the sizing statement, executed outside loops/deferred calls. Add mutations for delayed invocation and repeated closure invocation. The recorded M15 tests literal B3 duplication only, so its detection does not close this gap. These counterexamples were assessed statically, not executed.

2. **P3 — success characterization does not fully pin D3’s sealed-value equality.**  
   [a157_first_leg_binding_testseam_test.go:320](/mnt/D/Axipient/workspace/TossOS/internal/execgw/a157_first_leg_binding_testseam_test.go:320) checks quantities, successful issuance, decision ID and persistence. It does not directly compare the complete strategy/campaign/weekly seal or `issuedAt` against baseline expectations.

   Existing eight-lane and weekly-binding tests provide useful additional coverage, including policy binding and defensive copying, but do not make this test a complete seal-equivalence assertion.

   **Fix:** extend the fixed-clock characterization with exact expected persisted strategy/campaign/weekly values and issuance timestamp, or compare the complete internal precheck through package-local tests.

The requested production checks otherwise hold:

- **Accepted set / initialization:** [riskguardian_first_leg.go:66](/mnt/D/Axipient/workspace/TossOS/internal/execgw/riskguardian_first_leg.go:66) preserves B2–B8 predicates and relative order. Success requires both `bindErr == nil` and entry success. `policyVersion` and `weekly` are assigned before the binding closure succeeds; non-weekly `weekly == nil` is intentional. No newly introduced panic path found. Equivalence assumes identical clock observations, as D2 specifies.
- **D2(a)–(f):** [lines 111–118](/mnt/D/Axipient/workspace/TossOS/internal/execgw/riskguardian_first_leg.go:111) implement returned daily-loss chain refusal → binding defect → other entry refusal. Existing earlier entry checks still determine whether a daily-loss refusal is reached.
- **Classifier / admission:** [DailyLossChainRefusalOf:48](/mnt/D/Axipient/workspace/TossOS/internal/execgw/a157_daily_loss_kind.go:48) transcribes the previous conditions and order exactly. [admit:111](/mnt/D/Axipient/workspace/TossOS/internal/app/engine/strategy_first_leg_admission.go:111) preserves causes, escalation and subsequent branches. Existing a137 tests cover escalation failure and accumulated scope errors.
- **Assembly:** [strategy_entry_supervisor.go:396](/mnt/D/Axipient/workspace/TossOS/internal/app/engine/strategy_entry_supervisor.go:396) guards typed nil correctly. Construction only stores callbacks/gate; current projection reads neither, so no current logging or blocking behavior changes.
- `riskguardian_qfinal.go` is unchanged; the first-leg issuance function is byte-for-byte unchanged.

**Verification limit:** targeted tests could not start: `mkdir /tmp/go-build…: read-only file system`. Recorded mutation results were inspected, not independently rerun. Project memory lookup found no relevant evidence. No files modified; UI review is N/A for this backend-only change.
