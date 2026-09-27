# Branch Test Map: `readProductionRiskUsage`

Measured at HEAD `b8211926` with `go test -coverpkg ./internal/journal,./internal/riskbucket (journal untagged + riskbucket tossos_testseams, isolated copy)` (statement coverage; `covered` = the branch body ran at least once in the package suite, it does not say which test). Named tests are attributed per test with a single-test `-coverprofile` run. Harness: `analysis/harness/branch_coverage_rows.py`.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 453:2 — `if err != nil {`; then `return nil, err` (line last changed by `8022f578`) | package suite `go test -coverpkg ./internal/journal,./internal/riskbucket (journal untagged + riskbucket tossos_testseams, isolated copy)` | n/a — branch line last changed by `8022f578`, not an a066 commit | NOT covered at `b8211926` |
| B2 | for at 458:2 — `for rows.Next() {`; then `var row productionRiskUsageRow` (line last changed by `8022f578`) | package suite `go test -coverpkg ./internal/journal,./internal/riskbucket (journal untagged + riskbucket tossos_testseams, isolated copy)` | n/a — branch line last changed by `8022f578`, not an a066 commit | covered at `b8211926` |
| B3 | if at 460:3 — `if err := rows.Scan(&row.ReservationID, &row.PolicyVersion, &row.HeldMinor, &row.FilledMinor, &row.State,`; then `return nil, err` (line last changed by `8022f578`) | package suite `go test -coverpkg ./internal/journal,./internal/riskbucket (journal untagged + riskbucket tossos_testseams, isolated copy)` | n/a — branch line last changed by `8022f578`, not an a066 commit | NOT covered at `b8211926` |
