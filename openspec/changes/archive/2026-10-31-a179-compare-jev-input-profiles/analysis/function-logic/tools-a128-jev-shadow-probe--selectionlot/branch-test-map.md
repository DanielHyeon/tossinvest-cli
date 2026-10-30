# Branch Test Map: `selectionLot`

Relevant package test surfaces: TestSelectionLotSeparatesProfiles, TestReportSeparatesSelectionLots. These tests cover named scenarios, not every AST branch; rows below identify unproven individual branch coverage rather than asserting it. Existing socket-backed tests could not execute in this sandbox.

| Branch | Source branch | Individual branch coverage |
|---|---|---|
| B1 | if at 240:2: `if row.Selection != nil {` | Not individually established |
| B2 | if at 243:2: `if row.StateRev != "" {` | Not individually established |
| B3 | if at 246:2: `if row.InputProfile != "" {` | Not individually established |
