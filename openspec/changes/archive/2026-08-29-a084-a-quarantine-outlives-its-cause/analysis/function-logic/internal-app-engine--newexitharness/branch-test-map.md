# Branch Test Map: `newExitHarness`

AST의 모든 분기를 1행씩 덮는다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | (186) `if` — if err != nil | `newExitHarness`를 쓰는 a084 테스트 | yes | yes |
| B2 | (190) `if` — if err := j.SetApplyHooks(journal.ApplyHooks | `newExitHarness`를 쓰는 a084 테스트 | yes | yes |
| B3 | (205) `if` — if err != nil | `newExitHarness`를 쓰는 a084 테스트 | yes | yes |
| B4 | (235) `if` — if mutate != nil | `newExitHarness`를 쓰는 a084 테스트 | yes | yes |
| B5 | (239) `if` — if err != nil | `newExitHarness`를 쓰는 a084 테스트 | yes | yes |

추가 근거: `TestANewQuarantineRecordsTheSelectorThatJudgedIt`,
`TestAQuarantineWrittenBeforeTheColumnReadsAsUnknown`,
`TestAQuarantineFromTheCurrentSelectorIsNotReJudged` (journal 계층).
