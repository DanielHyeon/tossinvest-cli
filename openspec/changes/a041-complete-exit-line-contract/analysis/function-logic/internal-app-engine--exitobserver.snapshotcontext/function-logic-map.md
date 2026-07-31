# Function Logic Map: `ExitObserver.snapshotContext`

- Source: `internal/app/engine/exitloop.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| managed position and observed price | persisted position ID/generation/quantity; canonical market and symbol | frozen observer input | values are carried into evaluation; downstream validation refuses invalid quantities |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | happy path (branchless) | none; returns a value-only context | context with unique observation identity | a041 one-share observer integration tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `o.clk.Now` | binds one observation instant into the identity | deterministic injected clock; no retry | CodeGraph + AST |
| `fmt.Sprintf` and string normalization | creates a canonical observation identity | pure, cannot fail | CodeGraph + AST |

## State mutations and fallbacks

- Does not mutate observer or position state; all fields are copied into a new value.

## Safety conclusion

- Safe edit boundary: identity construction only; evaluation and proposal authority remain downstream.
- High-risk impact: yes, because deduplication identity protects exit proposal reuse.
