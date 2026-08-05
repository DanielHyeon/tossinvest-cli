# Branch Test Map: `Client.HoldingsRaw`

AST의 모든 분기를 1행씩 덮는다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | (119) `if` — if symbol != "" | `TestTheUnmanagedAlertNamesTheStockInKorean`, `TestAnUnnamedStockRendersAsItsCodeInTheAlert`, `TestNamingAStockCostsNoExtraRequest`, `TestTheCycleLearnsTheNameFromTheSnapshotItAlreadyTook` | yes | yes |
| B2 | (123) `if` — if err := c.getAcct(ctx, "/api/v1/holdings", q, &raw); err != nil | `TestTheUnmanagedAlertNamesTheStockInKorean`, `TestAnUnnamedStockRendersAsItsCodeInTheAlert`, `TestNamingAStockCostsNoExtraRequest`, `TestTheCycleLearnsTheNameFromTheSnapshotItAlreadyTook` | yes | yes |
| B3 | (127) `range` — for _, item := range raw.Items | `TestTheUnmanagedAlertNamesTheStockInKorean`, `TestAnUnnamedStockRendersAsItsCodeInTheAlert`, `TestNamingAStockCostsNoExtraRequest`, `TestTheCycleLearnsTheNameFromTheSnapshotItAlreadyTook` | yes | yes |

추가 근거: `TestALabelNamesTheStockAndKeepsTheCode`, `TestAnUnknownSymbolRendersAsItsCode`,
`TestANameSurvivesThePositionItNamed`, `TestAnEmptyReadDoesNotEraseAKnownName`,
`TestTheZeroValueIsUsable` (registry 계약).
