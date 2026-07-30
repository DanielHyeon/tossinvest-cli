# Function Logic Map: `Client.Accounts`

- Source: `internal/official/reads.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| context | live or cancelled request context | caller | token/HTTP cancellation propagates |
| context | live or cancelled request context | caller | helper propagates request error |
| account mutex | unlocked on entry | `Client.mu` | deferred unlock on every return |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| none | this wrapper has no conditional | acquires/releases `Client.mu` | exact helper result | public/scoped contention and cancellation tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `c.mu.Lock` / deferred `Unlock` | serialize public and lazy discovery | waits for current holder; always unlocks | post-edit AST |
| `accountsLocked` | perform the discovery and implicit-drift validation path | caller context; no retry | post-edit AST |

## State mutations and fallbacks

- This wrapper directly mutates only mutex state. Account cache priming and
  implicit-drift validation belong to the lock-assuming helper and therefore
  cannot escape discovery serialization.
- No HTTP retry, order call, journal write, or fallback account source is added.

## Safety conclusion

- Safe edit boundary: acquire the existing account mutex exactly once and call
  the lock-assuming helper; lazy resolution must never call this wrapper.
- High-risk impact: yes — this header scopes every official account read and
  mutation, so selection, invalid-response, concurrency, cancellation, and
  explicit override behavior require race-tested regressions.
