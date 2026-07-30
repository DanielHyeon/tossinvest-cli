# Branch Test Map: `newConsoleCmd`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | Help and sole `--port` flag describe the authenticated fixed-candidate update surface | `TestConsoleIsRegisteredAndAnnotated`, `TestConsoleOffersOnlyThePortFlag` | old help claimed all non-verification behavior was read-only | pass |
