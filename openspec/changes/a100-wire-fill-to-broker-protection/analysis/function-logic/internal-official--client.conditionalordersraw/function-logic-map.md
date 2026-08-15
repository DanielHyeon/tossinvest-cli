# Function Logic Map: `Client.ConditionalOrdersRaw`

- Source: `internal/official/conditional_reads.go:156-211`
- Qualified function: `Client.ConditionalOrdersRaw`
- Revision: `current`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Inputs and durable state | Exact typed/current values | `internal/official/conditional_reads.go` plus A100 tasks 0.2a.1–0.2a.9 | Reject or terminal HOLD; never infer evidence |
| Receipt/official evidence | Same-client raw result and attempts, active exclusive lease | Sealed official source and causal receipt | Any read/decode/identity/write/sync gap remains HOLD |
| Mutation authority | Exact M0 prerequisites and existing six methods | CLI/New gates and `MutationMethods()` | No factory/mutation outside the authorized trigger-only mode |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if` at `internal/official/conditional_reads.go:158` — `if strings.TrimSpace(status) == "" {` | Preserve source ordering; missing causal authority must HOLD | `TestAttemptObserverTraces401RefreshThenSuccessfulRetry` |
| B2 | `if` at `internal/official/conditional_reads.go:168` — `if status != "" {` | Preserve source ordering; missing causal authority must HOLD | `TestAttemptObserverTraces401RefreshThenSuccessfulRetry` |
| B3 | `if` at `internal/official/conditional_reads.go:171` — `if symbol != "" {` | Preserve source ordering; missing causal authority must HOLD | `TestAttemptObserverTraces401RefreshThenSuccessfulRetry` |
| B4 | `if` at `internal/official/conditional_reads.go:174` — `if cursor != "" {` | Preserve source ordering; missing causal authority must HOLD | `TestAttemptObserverTraces401RefreshThenSuccessfulRetry` |
| B5 | `if` at `internal/official/conditional_reads.go:177` — `if limit > 0 {` | Preserve source ordering; missing causal authority must HOLD | `TestAttemptObserverTraces401RefreshThenSuccessfulRetry` |
| B6 | `if` at `internal/official/conditional_reads.go:181` — `if err := c.getAcct(ctx, "/api/v1/conditional-orders", q, &raw); err != nil {` | Preserve source ordering; missing causal authority must HOLD | `TestAttemptObserverTraces401RefreshThenSuccessfulRetry` |
| B7 | `range` at `internal/official/conditional_reads.go:189` — `for _, o := range raw.ConditionalOrders {` | Preserve source ordering; missing causal authority must HOLD | `TestAttemptObserverTraces401RefreshThenSuccessfulRetry` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strings.TrimSpace` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `fmt.Errorf` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `q.Set` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `strconv.Itoa` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `c.getAcct` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `make` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `len` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `append` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |

## State mutations and fallbacks

- Receipt/checkpoint persistence precedes the broker action or read it authorizes.
- Pending recovery is read-only; parent/child unresolved states are manual-only and never cleanup targets.
- Retry success cannot erase an earlier critical attempt failure.

## Safety conclusion

- Safe edit boundary: exact same-client authority, exclusive receipt lease, causal fsync order, terminal HOLD, and six-method mutation surface.
- High-risk impact: yes; every AST branch is linked to the named M0 or preservation test.
