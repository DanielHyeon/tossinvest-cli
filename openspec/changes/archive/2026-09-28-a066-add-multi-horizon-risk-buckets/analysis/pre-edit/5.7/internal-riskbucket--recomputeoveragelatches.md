# Pre-edit branch table: `internal-riskbucket--recomputeoveragelatches` (before a066 5.7 shared overage)

| Branch | Position | Condition (AST source line at `cf381447`) | Coverage (pre-edit) |
|---|---|---|---|
| B1 | range at 352:2 | `for key, usage := range state.Buckets {` | covered |
| B2 | if at 354:3 | `if err != nil {` | NOT covered |
| B3 | if at 358:3 | `if err != nil {` | NOT covered |
| B4 | if at 362:3 | `if err != nil {` | NOT covered |
| B5 | if at 366:3 | `if err != nil {` | NOT covered |
| B6 | if at 370:3 | `if overage.Sign() > 0 {` | covered |
| B7 | if at 373:4 | `if err != nil {` | NOT covered |
| B8 | if at 376:4 | `if overage.Cmp(previous) > 0 {` | covered |
| B9 | if at 382:2 | `if !anyOverage {` | covered |
| B10 | range at 386:2 | `for key, usage := range state.Buckets {` | covered |
