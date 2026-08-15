# Branch Test Map: `Client.doRequest`

- Source: `internal/official/client.go:191-207`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if` at `internal/official/client.go:194` — `if err != nil {` | `TestAttemptObserverTraces401RefreshThenSuccessfulRetry` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B2 | `if` at `internal/official/client.go:201` — `if err != nil {` | `TestAttemptObserverTraces401RefreshThenSuccessfulRetry` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
