# Function Logic Map: `Client.doRequest`

- Source: `internal/official/client.go:191-207`
- Qualified function: `Client.doRequest`
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
| B1 | `if` at `internal/official/client.go:194` — `if err != nil {` | Preserve source ordering; missing causal authority must HOLD | `TestAttemptObserverTraces401RefreshThenSuccessfulRetry` |
| B2 | `if` at `internal/official/client.go:201` — `if err != nil {` | Preserve source ordering; missing causal authority must HOLD | `TestAttemptObserverTraces401RefreshThenSuccessfulRetry` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `time.Now` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `c.hc.Do` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `observeAttempt` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `req.Context` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `fmt.Errorf` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `resp.Body.Close` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `c.rates.record` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `readRateBudget` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `io.ReadAll` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `append` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `call` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |

## State mutations and fallbacks

- Receipt/checkpoint persistence precedes the broker action or read it authorizes.
- Pending recovery is read-only; parent/child unresolved states are manual-only and never cleanup targets.
- Retry success cannot erase an earlier critical attempt failure.

## Safety conclusion

- Safe edit boundary: exact same-client authority, exclusive receipt lease, causal fsync order, terminal HOLD, and six-method mutation surface.
- High-risk impact: yes; every AST branch is linked to the named M0 or preservation test.
