# Branch Test Map: `consolePerformanceCapabilities.Close`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | nil close is safe; real close stops reads | `TestConsolePerformanceCapabilitiesOpenOneProfileDatabaseForBothReadSeams` | no lifecycle | yes |
