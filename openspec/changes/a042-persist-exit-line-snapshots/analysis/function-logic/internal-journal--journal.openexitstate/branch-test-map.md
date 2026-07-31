# Branch Test Map: `Journal.OpenExitState`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | empty position id | existing invalid-seed tests | yes | pending |
| B2 | default policy kind | existing RATCHET open tests | yes | pending |
| B3 | invalid policy kind | existing policy validation tests | yes | pending |
| B4 | ratchet with ladder id | existing policy validation tests | yes | pending |
| B5 | ladder default/known id | policy snapshot tests | yes | pending |
| B6 | unknown ladder id | policy snapshot tests | yes | pending |
| B7 | invalid t0 arithmetic | existing exit-state tests | yes | pending |
| B8 | transaction begin failure | fault rollback test | yes | pending |
| B9 | missing position | existing not-found test | yes | pending |
| B10 | position read error | transaction fault test | yes | pending |
| B11 | ineligible position | existing eligibility test | yes | pending |
| B12 | invalid claimed identity | identity contract test | yes | pending |
| B13 | duplicate state | existing unique test | yes | pending |
| B14 | insert failure | v10 atomic seed test | yes | pending |
| B15 | event append failure | v10 fault-injection test | yes | pending |
| B16 | commit failure | v10 fault-injection test | yes | pending |
| B17 | successful state/event commit | v10 seed roundtrip test | yes | pending |
| B18 | post-commit state read | reopen test | yes | pending |
| B19 | policy identity return | identity roundtrip test | yes | pending |
| B20 | generation persistence | generation quarantine test | yes | pending |
