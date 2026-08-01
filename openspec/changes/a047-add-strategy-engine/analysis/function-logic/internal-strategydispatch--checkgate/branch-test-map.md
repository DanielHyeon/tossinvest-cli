# Branch Test Map: `checkGate`

| Branch | Scenario | Test | RED | GREEN |
|---|---|---|---|---|
| B1 | activation mismatch precedes operational blockers | initial gate precedence table | missing | pass |
| B2 | decision/order activation mismatch | 60/60 decision table + settings cases | existing partial | pass |
| B3 | ordered operational switch | initial gate table | missing | pass |
| B4 | lane desired/effective OFF | initial gate table | missing | pass |
| B5 | kill switch | initial gate table | missing | pass |
| B6 | protection unwired | initial gate table | missing | pass |
| B7 | reconciliation unhealthy | initial gate table | missing | pass |
| B8 | scheduler invalid | initial gate table | missing | pass |
| B9 | autostart disabled | initial gate table | missing | pass |
| B10 | gate closed | initial gate table | missing | pass |
| B11 | LIVE unapproved | initial gate table | missing | pass |
| Order | lane before kill, kill before protection, protection before later blockers | simultaneous-failure rows | missing | pass |
| Success | exact snapshot reaches post-validation planning core | post-validation confirmed spy test | existing | existing |
