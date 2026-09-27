# Branch Test Map: `readProductionRiskUsage`

Measured at HEAD `54e67495+5.7 working tree` with `go test -coverpkg ./internal/journal,./internal/riskbucket (journal untagged + riskbucket tossos_testseams, isolated copy)` (statement coverage; `covered` = the branch body ran at least once in the package suite, it does not say which test). Named tests are attributed per test with a single-test `-coverprofile` run. Harness: `analysis/harness/branch_coverage_rows.py`.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 457:2 — `if err != nil {`; then `return nil, err` (line last changed by `8022f578`) | package suite `go test -coverpkg ./internal/journal,./internal/riskbucket (journal untagged + riskbucket tossos_testseams, isolated copy)` | n/a — branch line last changed by `8022f578`, not an a066 commit | NOT covered at `54e67495+5.7 working tree` |
| B2 | for at 462:2 — `for rows.Next() {`; then `var row productionRiskUsageRow` (line last changed by `8022f578`) | package suite `go test -coverpkg ./internal/journal,./internal/riskbucket (journal untagged + riskbucket tossos_testseams, isolated copy)` | n/a — branch line last changed by `8022f578`, not an a066 commit | covered at `54e67495+5.7 working tree` |
| B3 | if at 464:3 — `if err := rows.Scan(&row.ReservationID, &row.PolicyVersion, &row.HeldMinor, &row.FilledMinor, &row.State,`; then `return nil, err` (line last changed by `8022f578`) | package suite `go test -coverpkg ./internal/journal,./internal/riskbucket (journal untagged + riskbucket tossos_testseams, isolated copy)` | n/a — branch line last changed by `8022f578`, not an a066 commit | NOT covered at `54e67495+5.7 working tree` |
