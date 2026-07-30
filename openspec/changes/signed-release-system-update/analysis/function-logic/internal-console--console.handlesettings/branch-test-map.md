# Branch Test Map: `Console.handleSettings`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | Adoption settings seam is wired and loaded | existing settings rendering tests | baseline | pass |
| B2 | Adoption load error is rendered without suppressing other sections | existing settings load-error test | baseline | pass |
| B3 | Limit settings seam is wired and loaded | existing limits rendering tests | baseline | pass |
| B4 | Limit load error is rendered without suppressing other sections | existing limits load-error test | baseline | pass |
| B5 | Trading policy seam is wired and loaded | existing operating-settings tests | baseline | pass |
| B6 | Trading policy load error is rendered without suppressing other sections | existing operating-settings load-error test | baseline | pass |
| B7 | Fixed updater inspection is paired only with a matching in-process signed receipt | `TestSettingsRendersSignedReleaseActionWithoutFetching` | unsigned candidate could appear signed | pass |
