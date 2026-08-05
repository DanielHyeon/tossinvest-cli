# Branch Test Map: `buildGateway`

AST의 모든 분기를 1행씩 덮는다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | (204) `if` — if err := checkProjectionWired(in.journal); err != nil | `TestTheUnmanagedAlertNamesTheStockInKorean`, `TestAnUnnamedStockRendersAsItsCodeInTheAlert`, `TestNamingAStockCostsNoExtraRequest`, `TestTheCycleLearnsTheNameFromTheSnapshotItAlreadyTook` | yes | yes |
| B2 | (226) `if` — if err := tracker.Restore(ctx); err != nil | `TestTheUnmanagedAlertNamesTheStockInKorean`, `TestAnUnnamedStockRendersAsItsCodeInTheAlert`, `TestNamingAStockCostsNoExtraRequest`, `TestTheCycleLearnsTheNameFromTheSnapshotItAlreadyTook` | yes | yes |
| B3 | (250) `if` — if err != nil | `TestTheUnmanagedAlertNamesTheStockInKorean`, `TestAnUnnamedStockRendersAsItsCodeInTheAlert`, `TestNamingAStockCostsNoExtraRequest`, `TestTheCycleLearnsTheNameFromTheSnapshotItAlreadyTook` | yes | yes |
| B4 | (274) `if` — if err != nil | `TestTheUnmanagedAlertNamesTheStockInKorean`, `TestAnUnnamedStockRendersAsItsCodeInTheAlert`, `TestNamingAStockCostsNoExtraRequest`, `TestTheCycleLearnsTheNameFromTheSnapshotItAlreadyTook` | yes | yes |

추가 근거: `TestALabelNamesTheStockAndKeepsTheCode`, `TestAnUnknownSymbolRendersAsItsCode`,
`TestANameSurvivesThePositionItNamed`, `TestAnEmptyReadDoesNotEraseAKnownName`,
`TestTheZeroValueIsUsable` (registry 계약).
