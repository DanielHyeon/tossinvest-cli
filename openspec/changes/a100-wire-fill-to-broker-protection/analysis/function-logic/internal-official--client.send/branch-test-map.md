# Branch Test Map: `Client.send`

- Source: `internal/official/client.go:320-366`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if` at `internal/official/client.go:322` — `if err != nil {` | `TestAttemptObserverTraces401RefreshThenSuccessfulRetry` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B2 | `if` at `internal/official/client.go:326` — `if err != nil {` | `TestAttemptObserverTraces401RefreshThenSuccessfulRetry` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B3 | `if` at `internal/official/client.go:330` — `if err != nil {` | `TestAttemptObserverTraces401RefreshThenSuccessfulRetry` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B4 | `for` at `internal/official/client.go:344` — `for attempt := 0; attempt < 2 && code == http.StatusUnauthorized; attempt++ {` | `TestAttemptObserverTraces401RefreshThenSuccessfulRetry` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B5 | `if` at `internal/official/client.go:347` — `if err != nil {` | `TestAttemptObserverTraces401RefreshThenSuccessfulRetry` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B6 | `if` at `internal/official/client.go:351` — `if err != nil {` | `TestAttemptObserverTraces401RefreshThenSuccessfulRetry` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B7 | `if` at `internal/official/client.go:355` — `if err != nil {` | `TestAttemptObserverTraces401RefreshThenSuccessfulRetry` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B8 | `if` at `internal/official/client.go:358` — `if !adopted {` | `TestAttemptObserverTraces401RefreshThenSuccessfulRetry` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B9 | `if` at `internal/official/client.go:362` — `if code < 200 \|\| code >= 300 {` | `TestAttemptObserverTraces401RefreshThenSuccessfulRetry` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
