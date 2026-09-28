# Branch Test Map: `aggregateProductionRiskUsage`

Measured at HEAD `54e67495+5.7 working tree` with `go test -coverpkg ./internal/journal,./internal/riskbucket (journal untagged + riskbucket tossos_testseams, isolated copy)` (statement coverage; `covered` = the branch body ran at least once in the package suite, it does not say which test). Named tests are attributed per test with a single-test `-coverprofile` run. Harness: `analysis/harness/branch_coverage_rows.py`.

6.4 refresh (2026-09-28): rows measured NOT covered earlier carry, alongside the historical cell, the package-wide statement coverage at `97ea352e` (journal untagged `-coverpkg journal,riskbucket`, riskbucket and execgw `tossos_testseams`; isolated copy). The earlier cell is kept as the record of that moment.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | range at 479:2 — `for _, row := range rows {`; then `rowFilled, filledOK := new(big.Int).SetString(row.FilledMinor, 10)` (line last changed by `8022f578`) | package suite `go test -coverpkg ./internal/journal,./internal/riskbucket (journal untagged + riskbucket tossos_testseams, isolated copy)` | n/a — branch line last changed by `8022f578`, not an a066 commit | covered at `54e67495+5.7 working tree` |
| B2 | if at 482:3 — `if !filledOK \|\| !heldOK \|\| rowFilled.Sign() < 0 \|\| rowHeld.Sign() < 0 \|\| rowFilled.BitLen() > 256 \|\| rowHeld.BitLen() > 256 \|\|`; then `return JournalBucketUsage{}, ErrJournalUsageInvalid` (line last changed by `8022f578`) | package suite `go test -coverpkg ./internal/journal,./internal/riskbucket (journal untagged + riskbucket tossos_testseams, isolated copy)` | n/a — branch line last changed by `8022f578`, not an a066 commit | covered at `54e67495+5.7 working tree` |
| B3 | if at 494:3 — `if filled.BitLen() > 256 \|\| held.BitLen() > 256 {`; then `return JournalBucketUsage{}, fmt.Errorf("%w: journal usage overflow", ErrJournalUsageInvalid)` (line last changed by `8022f578`) | package suite `go test -coverpkg ./internal/journal,./internal/riskbucket (journal untagged + riskbucket tossos_testseams, isolated copy)` | n/a — branch line last changed by `8022f578`, not an a066 commit | NOT covered at `54e67495+5.7 working tree` · package-wide at `97ea352e`: still NOT covered — review.md 6.1 (B)(2) |

Positions re-read from the post-edit `ast.json` (2026-09-29): B1 477→479, B2 480→482, B3 490→494 — the new struct field and the two loop lines of `28629ec6` moved them; conditions unchanged. Coverage cells are the historical measurements at the commits they name.
