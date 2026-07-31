# Branch Test Map: `Journal.OpenExitState`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | missing position ID | existing validation tests | existing | yes |
| B2 | empty kind selects ratchet | existing opening tests | existing | yes |
| B3 | invalid kind | existing validation tests | existing | yes |
| B4 | ratchet carries ladder ID | existing validation tests | existing | yes |
| B5 | ladder empty ID selects default | existing legacy tests | existing | yes |
| B6 | unknown ladder ID | existing policy snapshot tests | existing | yes |
| B7 | t0 arithmetic invalid | existing opening tests | existing | yes |
| B8 | transaction begin fails | journal failure tests | existing | yes |
| B9 | position not found | existing opening tests | existing | yes |
| B10 | position read fails | journal failure tests | existing | yes |
| B11 | ineligible position | existing unmanaged tests | existing | yes |
| B12 | exact omitted legacy identity remains compatible | journal identity test | yes | yes |
| B13 | supplied mismatched digest is refused | journal identity test | yes | yes |
| B14 | duplicate state | existing duplicate tests | existing | yes |
| B15 | insert failure | journal failure tests | existing | yes |
| B16 | opening event failure | journal failure tests | existing | yes |
| B17 | commit failure | journal failure tests | existing | yes |
| B18 | successful state returns runtime identity | runtime preservation test | yes | yes |
| B19 | adopted variant selects distinct pinned identity | adoption policy tests | existing | yes |
| B20 | no schema identity column is invented | schema v9 test + a042 handoff | yes | yes |
