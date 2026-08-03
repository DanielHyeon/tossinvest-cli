# Branch Test Map: `ReadOfficial`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | invalid scope or zero/tampered haircut policy cannot mint | invalid-authority tests | yes (atomic API absent) | yes |
| B2 | official read error produces no evidence | source and configured-origin tests | yes (separate check/read was mutable) | yes |
| B3 | configured origin maps to fail-closed invalid evidence | TestReadOfficialRefusesInvalidAuthority | yes (origin could change after check) | yes |
