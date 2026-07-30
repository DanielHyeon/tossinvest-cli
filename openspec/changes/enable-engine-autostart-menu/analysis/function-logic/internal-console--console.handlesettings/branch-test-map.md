# Branch Test Map: `Console.handleSettings`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | adoption settings seam is wired | existing adoption settings tests | yes | yes |
| B2 | adoption settings load fails visibly | existing adoption settings tests | yes | yes |
| B3 | Guardian limits seam is wired | existing limit settings tests | yes | yes |
| B4 | Guardian limits load fails visibly | existing limit settings tests | yes | yes |
| B5 | trading-policy seam is wired | existing operating-toggle tests | yes | yes |
| B6 | trading-policy load fails visibly | existing operating-toggle tests | yes | yes |
| B7 | engine autostart seam is wired and its value is loaded | `TestAutostartScreenRendersStateAndMeaning` | yes | yes |
| B8 | engine autostart load failure is rendered fail-closed | `TestAutostartUnwiredAndLoadErrorAreExplicit` | yes | yes |
| B9 | system updater seam is wired | existing system-update settings tests | yes | yes |
