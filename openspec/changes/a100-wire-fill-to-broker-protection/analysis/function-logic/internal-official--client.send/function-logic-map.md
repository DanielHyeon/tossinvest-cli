# Function Logic Map: `Client.send`

- Source: `internal/official/client.go:320-366`
- Qualified function: `Client.send`
- Revision: `current`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Inputs and durable state | Exact typed/current values | `internal/official/client.go` plus A100 tasks 0.2a.1–0.2a.9 | Reject or terminal HOLD; never infer evidence |
| Receipt/official evidence | Same-client raw result and attempts, active exclusive lease | Sealed official source and causal receipt | Any read/decode/identity/write/sync gap remains HOLD |
| Mutation authority | Exact M0 prerequisites and existing six methods | CLI/New gates and `MutationMethods()` | No factory/mutation outside the authorized trigger-only mode |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if` at `internal/official/client.go:322` — `if err != nil {` | Preserve source ordering; missing causal authority must HOLD | `TestAttemptObserverTraces401RefreshThenSuccessfulRetry` |
| B2 | `if` at `internal/official/client.go:326` — `if err != nil {` | Preserve source ordering; missing causal authority must HOLD | `TestAttemptObserverTraces401RefreshThenSuccessfulRetry` |
| B3 | `if` at `internal/official/client.go:330` — `if err != nil {` | Preserve source ordering; missing causal authority must HOLD | `TestAttemptObserverTraces401RefreshThenSuccessfulRetry` |
| B4 | `for` at `internal/official/client.go:344` — `for attempt := 0; attempt < 2 && code == http.StatusUnauthorized; attempt++ {` | Preserve source ordering; missing causal authority must HOLD | `TestAttemptObserverTraces401RefreshThenSuccessfulRetry` |
| B5 | `if` at `internal/official/client.go:347` — `if err != nil {` | Preserve source ordering; missing causal authority must HOLD | `TestAttemptObserverTraces401RefreshThenSuccessfulRetry` |
| B6 | `if` at `internal/official/client.go:351` — `if err != nil {` | Preserve source ordering; missing causal authority must HOLD | `TestAttemptObserverTraces401RefreshThenSuccessfulRetry` |
| B7 | `if` at `internal/official/client.go:355` — `if err != nil {` | Preserve source ordering; missing causal authority must HOLD | `TestAttemptObserverTraces401RefreshThenSuccessfulRetry` |
| B8 | `if` at `internal/official/client.go:358` — `if !adopted {` | Preserve source ordering; missing causal authority must HOLD | `TestAttemptObserverTraces401RefreshThenSuccessfulRetry` |
| B9 | `if` at `internal/official/client.go:362` — `if code < 200 \|\| code >= 300 {` | Preserve source ordering; missing causal authority must HOLD | `TestAttemptObserverTraces401RefreshThenSuccessfulRetry` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `c.tm.token` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `makeReq` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `c.doRequest` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `c.tm.refresh` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `classifyStatus` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `unwrapAndDecode` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |

## State mutations and fallbacks

- Receipt/checkpoint persistence precedes the broker action or read it authorizes.
- Pending recovery is read-only; parent/child unresolved states are manual-only and never cleanup targets.
- Retry success cannot erase an earlier critical attempt failure.

## Safety conclusion

- Safe edit boundary: exact same-client authority, exclusive receipt lease, causal fsync order, terminal HOLD, and six-method mutation surface.
- High-risk impact: yes; every AST branch is linked to the named M0 or preservation test.
