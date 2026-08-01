# Branch Test Map: `Console.handleProtectionApply`
| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | unwired seam refuses | default-OFF UI test | route absent | pass |
| B2 | empty capability refuses | apply validation table | unchecked capability | pass |
| B3 | commander error enters fail-closed switch | apply validation table | missing mapping | pass |
| B4 | error classification switch | apply validation table | missing mapping | pass |
| B5 | three-second delay not elapsed | `TestExitProtectionApplyRequiresCheckboxAndThreeSecondDelay` | early apply accepted | pass |
| B6 | stale capability | apply validation table | stale apply accepted | pass |
| B7 | checkbox missing | `TestExitProtectionApplyRequiresCheckboxAndThreeSecondDelay` | unchecked apply accepted | pass |
| B8 | other error | apply validation table | leaked success | pass |
