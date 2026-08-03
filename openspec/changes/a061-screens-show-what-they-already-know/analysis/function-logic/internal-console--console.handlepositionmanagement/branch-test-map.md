# Branch Test Map: `Console.handlePositionManagement`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | no settings seam names itself | existing settings tests | n/a | pass |
| B2 | a failed desired load is reported | existing tests | n/a | pass |
| B3 | same condition, other arm | same | n/a | pass |
| B4 | a successful desired load fills the panel | existing tests | n/a | pass |
| B5 | no commander renders the unwired page and returns | existing unwired test | n/a | pass |
| B6 | a runtime error is reported | existing tests | n/a | pass |
| B7 | a known runtime fills effective settings | existing tests | n/a | pass |
| B8 | account-level reconcile blocks render | existing reconcile tests | n/a | pass |
| B9 | a list error returns early | existing tests | n/a | pass |
| B10 | one row per live position; the KR row is named; the US row is named the same way; a same-ticker holding in another market does not name this row; an empty cache renders the row with no name and no broker call | existing tests; `TestPositionManagementNamesTheStock`; same test; `TestPositionManagementStaysSilentAndCallsNobodyWhenTheNameIsUnknown` | yes | yes |
| B11 | a per-row reconcile block renders | existing tests | n/a | pass |
| B12 | managed rows offer inherit and override | existing action tests | n/a | pass |
| B13 | released and eligible rows offer re-adopt | existing action tests | n/a | pass |
| B14 | one override action per registered policy | existing action tests | n/a | pass |
| B15 | externally eligible rows offer release | existing action tests | n/a | pass |
| B16 | same condition as B13 | same | n/a | pass |
