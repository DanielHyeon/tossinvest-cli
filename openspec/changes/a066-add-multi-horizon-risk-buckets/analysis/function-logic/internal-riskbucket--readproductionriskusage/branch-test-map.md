# Branch Test Map: `readProductionRiskUsage`

Measured at HEAD `54e67495+5.7 working tree` with `go test -coverpkg ./internal/journal,./internal/riskbucket (journal untagged + riskbucket tossos_testseams, isolated copy)` (statement coverage; `covered` = the branch body ran at least once in the package suite, it does not say which test). Named tests are attributed per test with a single-test `-coverprofile` run. Harness: `analysis/harness/branch_coverage_rows.py`.

6.4 refresh (2026-09-28): rows measured NOT covered earlier carry, alongside the historical cell, the package-wide statement coverage at `97ea352e` (journal untagged `-coverpkg journal,riskbucket`, riskbucket and execgw `tossos_testseams`; isolated copy). The earlier cell is kept as the record of that moment.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 457:2 — `if err != nil {`; then `return nil, err` (line last changed by `8022f578`) | package suite `go test -coverpkg ./internal/journal,./internal/riskbucket (journal untagged + riskbucket tossos_testseams, isolated copy)` | n/a — branch line last changed by `8022f578`, not an a066 commit | NOT covered at `54e67495+5.7 working tree` · package-wide at `97ea352e`: still NOT covered — review.md 6.1 (B)(2) |
| B2 | for at 462:2 — `for rows.Next() {`; then `var row productionRiskUsageRow` (line last changed by `8022f578`) | package suite `go test -coverpkg ./internal/journal,./internal/riskbucket (journal untagged + riskbucket tossos_testseams, isolated copy)` | n/a — branch line last changed by `8022f578`, not an a066 commit | covered at `54e67495+5.7 working tree` |
| B3 | if at 464:3 — `if err := rows.Scan(&row.ReservationID, &row.PolicyVersion, &row.HeldMinor, &row.FilledMinor, &row.State,`; then `return nil, err` (line last changed by `8022f578`) | package suite `go test -coverpkg ./internal/journal,./internal/riskbucket (journal untagged + riskbucket tossos_testseams, isolated copy)` | n/a — branch line last changed by `8022f578`, not an a066 commit | NOT covered at `54e67495+5.7 working tree` · package-wide at `97ea352e`: still NOT covered — review.md 6.1 (B)(2) |
