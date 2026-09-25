# CodeGraphContext context — a066 Wave 2A (2026-09-25)

- `make sdd-sync` at HEAD `d72bc401`: `codegraphcontext update . --quiet` **timed out after 300 s**; the sync ended
  `incomplete: codegraphcontext` (rc 2). GBrain advisory was busy (`gbrain serve` held the TossOS data home) and kept
  its previous freshness.
- Direct queries against the existing (stale) falkordb graph, each `timeout 90 codegraphcontext analyze callers <sym>`:
  `RecordFill`, `runApplyHooks`, `ApplyFill`, `checkReservation`, `CommitRiskBucketAdmission` → all
  "No callers found".
- Conclusion: CodeGraphContext supplied **no** supporting context for this lot. Nothing in this lot relies on it;
  every structural claim is backed by CodeGraph 1.6.0 plus the current HEAD source/AST (see
  `evidence-reconciliation.md`). Advisory status: `not-applicable` — tool unavailable, not skipped.
