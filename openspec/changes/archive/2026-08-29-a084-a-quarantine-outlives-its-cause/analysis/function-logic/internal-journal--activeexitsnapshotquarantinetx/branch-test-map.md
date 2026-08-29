# Branch Test Map: `activeExitSnapshotQuarantineTx`

AST의 모든 분기를 1행씩 덮는다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | (793) `if` — if errors.Is(err, sql.ErrNoRows) | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |

추가 근거: `TestANewQuarantineRecordsTheSelectorThatJudgedIt`,
`TestAQuarantineWrittenBeforeTheColumnReadsAsUnknown`,
`TestAQuarantineFromTheCurrentSelectorIsNotReJudged` (journal 계층).
