# Branch Test Map: `Context.ReconcileDriver`

AST의 모든 분기를 1행씩 덮는다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | (342) `if` — if c == nil | `TestTheUnmanagedAlertNamesTheStockInKorean`, `TestAnUnnamedStockRendersAsItsCodeInTheAlert`, `TestNamingAStockCostsNoExtraRequest`, `TestTheCycleLearnsTheNameFromTheSnapshotItAlreadyTook` | yes | yes |
| B2 | (345) `if` — if !c.Automation.Verified | `TestTheUnmanagedAlertNamesTheStockInKorean`, `TestAnUnnamedStockRendersAsItsCodeInTheAlert`, `TestNamingAStockCostsNoExtraRequest`, `TestTheCycleLearnsTheNameFromTheSnapshotItAlreadyTook` | yes | yes |
| B3 | (357) `if` — if opts.Prices == nil | `TestTheUnmanagedAlertNamesTheStockInKorean`, `TestAnUnnamedStockRendersAsItsCodeInTheAlert`, `TestNamingAStockCostsNoExtraRequest`, `TestTheCycleLearnsTheNameFromTheSnapshotItAlreadyTook` | yes | yes |
| B4 | (360) `if` — if opts.Alerts == nil && c.Notifier != nil | `TestTheUnmanagedAlertNamesTheStockInKorean`, `TestAnUnnamedStockRendersAsItsCodeInTheAlert`, `TestNamingAStockCostsNoExtraRequest`, `TestTheCycleLearnsTheNameFromTheSnapshotItAlreadyTook` | yes | yes |
| B5 | (363) `if` — if opts.Log == nil | `TestTheUnmanagedAlertNamesTheStockInKorean`, `TestAnUnnamedStockRendersAsItsCodeInTheAlert`, `TestNamingAStockCostsNoExtraRequest`, `TestTheCycleLearnsTheNameFromTheSnapshotItAlreadyTook` | yes | yes |

추가 근거: `TestALabelNamesTheStockAndKeepsTheCode`, `TestAnUnknownSymbolRendersAsItsCode`,
`TestANameSurvivesThePositionItNamed`, `TestAnEmptyReadDoesNotEraseAKnownName`,
`TestTheZeroValueIsUsable` (registry 계약).
