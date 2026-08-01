# Branch Test Map: `scanExitProgress`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | missing exit state maps typed error | `exit judgement suite` | yes | yes |
| B2 | query failure propagates | `exit judgement suite` | yes | yes |
| B3 | nullable rung hydrates only when present | `ratchet suite` | yes | yes |
| B4 | effective snapshot decodes when present | `exit snapshot suite` | yes | yes |
| B5 | corrupt effective snapshot fails closed | `exit snapshot corruption suite` | yes | yes |
