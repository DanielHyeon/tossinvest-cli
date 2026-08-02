# Function Logic Map: `Console.decoratePositionRows`

- Source: `internal/console/portfolio_pages.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `ctx` | Request context | HTTP handler / dashboard assembly | Read adapters may fail closed without mutating state |
| `rows` | Request-local joined broker/journal rows, including empty | `positions()` / `joinPositions()` | Empty slice remains empty |
| `asOf` | Console clock instant | `c.now()` | Used only for age/freshness projections |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B3 | Position-policy adapter exists; runtime/list succeeds or fails; iterate returned states | Populate request-local runtime/index | Failure preserves unknown effective state | policy runtime/list tests + projection parity |
| B4-B6 | Settings adapter exists; load succeeds/fails; iterate rows | Set request-local designated/excluded flags | Load failure leaves flags unset | settings/adoption tests |
| B7-B12 | Runtime was attempted; iterate rows; resolve lifecycle state/release/block | Populate request-local management projections | Missing lifecycle fails closed | lifecycle/reconcile/US tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `PositionPolicies.Runtime` / `List` | Read effective policy and lifecycle evidence | No retry; errors become unknown | CodeGraph + AST |
| `Settings.Load` | Read desired include/exclude snapshot once | No retry; error leaves desired flags unknown | CodeGraph + AST |
| `positionpolicy.ProjectManagement` | Produce the canonical management verdict | Pure projection | AST |
| `attachPositionExitLines` | Attach canonical fail-closed exit-line and reference views | Pure/request-local mutation; stale values suppressed | CodeGraph + AST |

## State mutations and fallbacks

- Mutates only `rows`, a request-local display slice. It cannot place orders,
  write the journal, save configuration, change toggles, or call the broker.
- Runtime, settings, and lifecycle failures stay unknown instead of falling
  back to desired policy or stored raw prices.

## Safety conclusion

- Safe edit boundary: shared console display enrichment only.
- High-risk impact: reviewed-safe; trading-state reads are fail-closed and all
  operational safety mutations remain unavailable from the dependency graph.
