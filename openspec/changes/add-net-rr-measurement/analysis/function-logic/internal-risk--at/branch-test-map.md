# Branch Test Map: `at`

Source: `internal/risk/chain.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `!d.Allowed` | `TestAnAllowedVerdictNamesNoStep` (the false arm), `TestEveryRungReportsItsOwnName` (the true arm) | pre-existing | yes |
