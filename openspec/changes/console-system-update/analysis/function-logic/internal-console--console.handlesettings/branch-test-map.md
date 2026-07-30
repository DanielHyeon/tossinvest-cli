# Branch Test Map: `Console.handleSettings`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | adoption seam wired | existing settings tests | baseline | pass |
| B2 | adoption load error remains local | existing settings error test | baseline | pass |
| B3 | limit seam wired | existing limit tests | baseline | pass |
| B4 | limit load error remains local | existing limit error test | baseline | pass |
| B5 | trading seam wired | existing policy tests | baseline | pass |
| B6 | trading load error remains local | existing policy error test | baseline | pass |
| B7 | fixed updater metadata is rendered without path input | `TestSettingsRendersTheFixedReviewedSystemUpdate` | no update panel | pass |
