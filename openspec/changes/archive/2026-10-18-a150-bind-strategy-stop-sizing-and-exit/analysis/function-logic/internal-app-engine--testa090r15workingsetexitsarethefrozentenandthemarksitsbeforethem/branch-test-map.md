# Branch Test Map: `TestA090R15WorkingSetExitsAreTheFrozenTenAndTheMarkSitsBeforeThem`

Current AST branch inventory. Scenario rows name conditions; they do not claim isolated branch coverage. Focused receipts are in the implementation evidence; full regression remains pending.

| Branch | Source line and condition | Verification status |
|---|---|---|
| B1 | 219: `if problems := a090CheckWorkingSet(t, a090ReadExitloop(t)); len(problems) > 0 {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
