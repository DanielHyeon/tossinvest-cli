# CodeGraph baseline

Worktree `/tmp/tossos-a148-impl-20261018`, base 28a800f5ed5eed7c0db0aed4f435d0778b034191. `sdd_sync.py --no-gbrain` exited 0; CodeGraph and CodeGraphContext updates both completed. GBrain intentionally skipped as advisory. Per-symbol `*.txt` captures definition, callers, callees and impact after sync. Two new test files created during sync were pending in CodeGraph status; all indexed production symbols remain at base and source/AST confirms them. Re-index after edits before final gate.

`openState` affects `workingSet` and `ObserveOnce`. First-leg admission uses atomic journal transaction; Guardian builds lineage; final transport check currently validates owner and lease only. Immutable projection and shared parser boundaries must be preserved. `ParkerRRFromDecision` and `ParkerRealizedRR` are pure diagnostics, never sealed decisions or runtime authority. `discovery-plan.md` contains producer/consumer matrix and configuration binding.
