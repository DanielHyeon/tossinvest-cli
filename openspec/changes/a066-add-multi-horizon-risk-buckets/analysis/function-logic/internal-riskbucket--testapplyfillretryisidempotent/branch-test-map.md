# Branch Test Map: `TestApplyFillRetryIsIdempotent`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | first apply fails (if at 37) | this test | n/a — body unchanged since the old base | GREEN (`go test ./internal/riskbucket`, 2026-09-29) |
| B2 | retry apply fails (if at 41) | this test | n/a — body unchanged | GREEN |
| B3 | retry is not a duplicate (if at 44) | this test | n/a — body unchanged | GREEN |
