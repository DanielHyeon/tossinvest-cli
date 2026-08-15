# Branch Test Map: `Client.ConditionalOrdersRaw`

- Source: `internal/official/conditional_reads.go:156-211`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if` at `internal/official/conditional_reads.go:158` — `if strings.TrimSpace(status) == "" {` | `TestAttemptObserverTraces401RefreshThenSuccessfulRetry` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B2 | `if` at `internal/official/conditional_reads.go:168` — `if status != "" {` | `TestAttemptObserverTraces401RefreshThenSuccessfulRetry` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B3 | `if` at `internal/official/conditional_reads.go:171` — `if symbol != "" {` | `TestAttemptObserverTraces401RefreshThenSuccessfulRetry` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B4 | `if` at `internal/official/conditional_reads.go:174` — `if cursor != "" {` | `TestAttemptObserverTraces401RefreshThenSuccessfulRetry` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B5 | `if` at `internal/official/conditional_reads.go:177` — `if limit > 0 {` | `TestAttemptObserverTraces401RefreshThenSuccessfulRetry` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B6 | `if` at `internal/official/conditional_reads.go:181` — `if err := c.getAcct(ctx, "/api/v1/conditional-orders", q, &raw); err != nil {` | `TestAttemptObserverTraces401RefreshThenSuccessfulRetry` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B7 | `range` at `internal/official/conditional_reads.go:189` — `for _, o := range raw.ConditionalOrders {` | `TestAttemptObserverTraces401RefreshThenSuccessfulRetry` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
