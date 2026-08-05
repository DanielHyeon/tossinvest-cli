# Branch Test Map: `NewContext`

AST의 모든 분기를 1행씩 덮는다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | (420) `if` — if err != nil | `TestTheUnmanagedAlertNamesTheStockInKorean`, `TestAnUnnamedStockRendersAsItsCodeInTheAlert`, `TestNamingAStockCostsNoExtraRequest`, `TestTheCycleLearnsTheNameFromTheSnapshotItAlreadyTook` | yes | yes |
| B2 | (427) `if` — if err != nil | `TestTheUnmanagedAlertNamesTheStockInKorean`, `TestAnUnnamedStockRendersAsItsCodeInTheAlert`, `TestNamingAStockCostsNoExtraRequest`, `TestTheCycleLearnsTheNameFromTheSnapshotItAlreadyTook` | yes | yes |
| B3 | (431) `if` — if clk == nil | `TestTheUnmanagedAlertNamesTheStockInKorean`, `TestAnUnnamedStockRendersAsItsCodeInTheAlert`, `TestNamingAStockCostsNoExtraRequest`, `TestTheCycleLearnsTheNameFromTheSnapshotItAlreadyTook` | yes | yes |
| B4 | (445) `if` — if opts.Publisher != nil | `TestTheUnmanagedAlertNamesTheStockInKorean`, `TestAnUnnamedStockRendersAsItsCodeInTheAlert`, `TestNamingAStockCostsNoExtraRequest`, `TestTheCycleLearnsTheNameFromTheSnapshotItAlreadyTook` | yes | yes |
| B5 | (453) `if` — if err := recordGateSettings(auditLog, gate, cfg.Engine.Adoption, notification | `TestTheUnmanagedAlertNamesTheStockInKorean`, `TestAnUnnamedStockRendersAsItsCodeInTheAlert`, `TestNamingAStockCostsNoExtraRequest`, `TestTheCycleLearnsTheNameFromTheSnapshotItAlreadyTook` | yes | yes |
| B6 | (468) `if` — if err != nil | `TestTheUnmanagedAlertNamesTheStockInKorean`, `TestAnUnnamedStockRendersAsItsCodeInTheAlert`, `TestNamingAStockCostsNoExtraRequest`, `TestTheCycleLearnsTheNameFromTheSnapshotItAlreadyTook` | yes | yes |
| B7 | (475) `if` — if err != nil | `TestTheUnmanagedAlertNamesTheStockInKorean`, `TestAnUnnamedStockRendersAsItsCodeInTheAlert`, `TestNamingAStockCostsNoExtraRequest`, `TestTheCycleLearnsTheNameFromTheSnapshotItAlreadyTook` | yes | yes |
| B8 | (484) `if` — if err := bindApplyHooks(jrn); err != nil | `TestTheUnmanagedAlertNamesTheStockInKorean`, `TestAnUnnamedStockRendersAsItsCodeInTheAlert`, `TestNamingAStockCostsNoExtraRequest`, `TestTheCycleLearnsTheNameFromTheSnapshotItAlreadyTook` | yes | yes |
| B9 | (503) `if` — if err != nil | `TestTheUnmanagedAlertNamesTheStockInKorean`, `TestAnUnnamedStockRendersAsItsCodeInTheAlert`, `TestNamingAStockCostsNoExtraRequest`, `TestTheCycleLearnsTheNameFromTheSnapshotItAlreadyTook` | yes | yes |
| B10 | (518) `if` — if gate.Enabled && guardian == nil && !opts.disableProductionGuardian | `TestTheUnmanagedAlertNamesTheStockInKorean`, `TestAnUnnamedStockRendersAsItsCodeInTheAlert`, `TestNamingAStockCostsNoExtraRequest`, `TestTheCycleLearnsTheNameFromTheSnapshotItAlreadyTook` | yes | yes |
| B11 | (520) `if` — if factory == nil | `TestTheUnmanagedAlertNamesTheStockInKorean`, `TestAnUnnamedStockRendersAsItsCodeInTheAlert`, `TestNamingAStockCostsNoExtraRequest`, `TestTheCycleLearnsTheNameFromTheSnapshotItAlreadyTook` | yes | yes |
| B12 | (526) `if` — if err != nil | `TestTheUnmanagedAlertNamesTheStockInKorean`, `TestAnUnnamedStockRendersAsItsCodeInTheAlert`, `TestNamingAStockCostsNoExtraRequest`, `TestTheCycleLearnsTheNameFromTheSnapshotItAlreadyTook` | yes | yes |
| B13 | (546) `if` — if err != nil | `TestTheUnmanagedAlertNamesTheStockInKorean`, `TestAnUnnamedStockRendersAsItsCodeInTheAlert`, `TestNamingAStockCostsNoExtraRequest`, `TestTheCycleLearnsTheNameFromTheSnapshotItAlreadyTook` | yes | yes |
| B14 | (550) `if` — if !automation.Verified | `TestTheUnmanagedAlertNamesTheStockInKorean`, `TestAnUnnamedStockRendersAsItsCodeInTheAlert`, `TestNamingAStockCostsNoExtraRequest`, `TestTheCycleLearnsTheNameFromTheSnapshotItAlreadyTook` | yes | yes |
| B15 | (555) `if` — if err != nil | `TestTheUnmanagedAlertNamesTheStockInKorean`, `TestAnUnnamedStockRendersAsItsCodeInTheAlert`, `TestNamingAStockCostsNoExtraRequest`, `TestTheCycleLearnsTheNameFromTheSnapshotItAlreadyTook` | yes | yes |

추가 근거: `TestALabelNamesTheStockAndKeepsTheCode`, `TestAnUnknownSymbolRendersAsItsCode`,
`TestANameSurvivesThePositionItNamed`, `TestAnEmptyReadDoesNotEraseAKnownName`,
`TestTheZeroValueIsUsable` (registry 계약).
