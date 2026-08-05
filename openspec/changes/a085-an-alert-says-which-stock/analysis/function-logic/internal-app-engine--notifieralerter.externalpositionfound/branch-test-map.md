# Branch Test Map: `notifierAlerter.ExternalPositionFound`

AST의 모든 분기를 1행씩 덮는다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | (100) `if` — if a.notifier == nil | `TestTheUnmanagedAlertNamesTheStockInKorean`, `TestAnUnnamedStockRendersAsItsCodeInTheAlert`, `TestNamingAStockCostsNoExtraRequest`, `TestTheCycleLearnsTheNameFromTheSnapshotItAlreadyTook` | yes | yes |

추가 근거: `TestALabelNamesTheStockAndKeepsTheCode`, `TestAnUnknownSymbolRendersAsItsCode`,
`TestANameSurvivesThePositionItNamed`, `TestAnEmptyReadDoesNotEraseAKnownName`,
`TestTheZeroValueIsUsable` (registry 계약).
