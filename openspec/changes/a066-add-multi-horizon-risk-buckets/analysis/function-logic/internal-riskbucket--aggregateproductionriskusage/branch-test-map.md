# Branch Test Map: `aggregateProductionRiskUsage`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | range at 448:2 — `for _, row := range rows {` | package suites (pre-edit) | n/a (pre-edit map) | covered at `f2decd0a` |
| B2 | if at 451:3 — `if !filledOK \ | package suites (pre-edit) | n/a (pre-edit map) | \ at `f2decd0a` |
| B3 | if at 459:3 — `if filled.BitLen() > 256 \ | package suites (pre-edit) | n/a (pre-edit map) | \ at `f2decd0a` |
