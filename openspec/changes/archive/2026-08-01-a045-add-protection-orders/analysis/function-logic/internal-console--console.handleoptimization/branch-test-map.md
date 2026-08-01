# Branch Test Map: `Console.handleOptimization`
| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | non-read method rejected | console method/static suite | existing coverage | pass |
| B2 | exit-policy seam wired | optimization suite | existing coverage | pass |
| B3 | exit-policy load failure displayed | optimization suite | existing coverage | pass |
| B4 | fixed policies rendered | optimization suite | existing coverage | pass |
| B5 | protection seam wired | `TestExitProtectionCurrentRowUsesOpaqueActionOnly` | missing status lane | pass |
| B6 | protection list failure displayed | exit-protection UI suite | missing failure state | pass |
| B7 | protection list success rendered | `TestExitProtectionCurrentRowUsesOpaqueActionOnly` | missing status lane | pass |
