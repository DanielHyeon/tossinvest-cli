# Branch Test Map: `Journal.QuarantineExitSnapshot`

AST의 모든 분기를 1행씩 덮는다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | (735) `if` — if id == "" || generation < 0 || why == "" || proof == "" | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B2 | (739) `if` — if err != nil | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B3 | (744) `if` — if err := tx.QueryRowContext(ctx, `SELECT instance_seq FROM positions WHERE id… | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B4 | (747) `if` — if actual != generation | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B5 | (750) `if` — if active, ok, err := activeExitSnapshotQuarantineTx(ctx, tx, id, generation);… | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B6 | (752) `else` — } else if ok | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B7 | (752) `if` — } else if ok | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B8 | (753) `if` — if !active.NeedsReJudgement() | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B9 | (759) `if` — if err := releaseExitSnapshotQuarantineTx(ctx, tx, active, QuarantineReleaseSe… | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B10 | (767) `if` — if err := tx.QueryRowContext(ctx, `SELECT coalesce(max(quarantine_version),0)+… | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B11 | (773) `if` — if _, err := tx.ExecContext(ctx, `INSERT INTO exit_snapshot_quarantines | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B12 | (779) `if` — if err := tx.Commit(); err != nil | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |

추가 근거: `TestANewQuarantineRecordsTheSelectorThatJudgedIt`,
`TestAQuarantineWrittenBeforeTheColumnReadsAsUnknown`,
`TestAQuarantineFromTheCurrentSelectorIsNotReJudged` (journal 계층).
