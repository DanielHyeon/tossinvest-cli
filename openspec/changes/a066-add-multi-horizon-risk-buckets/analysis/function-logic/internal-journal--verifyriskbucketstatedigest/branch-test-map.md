# Branch Test Map: `verifyRiskBucketStateDigest`

Post-edit (6.5 fix lot, 2026-09-28). Pre-edit AST and map: `analysis/pre-edit/6.5-fixes/internal-journal--verifyriskbucketstatedigest/`. Mutation column refers to the mutation-6.5 ledger `analysis/mutation-6.5/fixes-ledger.tsv`.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | reconstruction error (1095) | fill-layer tests | — | covered (suite rc 0 at the fix tree) |
| B2 | seal query error (1099) | storage exit — structural test | n/a | covered (suite rc 0 at the fix tree) |
| B3 | seal missing — `sql.ErrNoRows` → `ErrRiskBucketReplayMismatch` "state seal missing" (1102) — **new 6.5** | `TestA066FillWithAMissingStateSealLatchesAndKeepsTheFill` (RED at 4fb0b26e) | RED at `4fb0b26e` (`pre-edit/6.5-fixes/red-at-4fb0b26e.log`) | covered (suite rc 0 at the fix tree) |
| B4 | digest mismatch (1107) | state digest drift tests | — | covered (suite rc 0 at the fix tree) |
