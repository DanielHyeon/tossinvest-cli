# Branch Test Map: `quarantineExitSnapshotTx`

AST의 모든 분기를 1행씩 덮는다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | (509) `if` — if active, ok, err := activeExitSnapshotQuarantineTx(ctx, tx, id, generation);… | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B2 | (511) `else` — } else if ok | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B3 | (511) `if` — } else if ok | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B4 | (512) `if` — if !active.NeedsReJudgement() | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B5 | (520) `if` — if err := releaseExitSnapshotQuarantineTx(ctx, tx, active, QuarantineReleaseSe… | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B6 | (528) `if` — if err := tx.QueryRowContext(ctx, `SELECT coalesce(max(quarantine_version),0)+… | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |

추가 근거: `TestANewQuarantineRecordsTheSelectorThatJudgedIt`,
`TestAQuarantineWrittenBeforeTheColumnReadsAsUnknown`,
`TestAQuarantineFromTheCurrentSelectorIsNotReJudged` (journal 계층).
