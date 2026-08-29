# Function Logic Map: `TestEngineRuntimeConstructionBranchesFailClosedAndAssembleExactSuccess`

- Source: `cmd/tossctl/engine_runtime_branch_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| table fixture | constructor-safe in-memory engine context | `runtimeBranchContext` | each deliberate missing dependency must return its sentinel |
| request context | background, non-cancelled | test | no broker read or mutation is invoked |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | reconcile tracker missing | recovery latch may already be armed by prior construction only | `ErrReconcileDriverUnavailable` | table case B2 |
| B2 | reduction issuer missing | no runtime returned | `ErrExitObserverUnavailable` | table case B3 |
| B3 | resolver missing | no runtime returned | `ErrRecoveryUnavailable` | table case B4 |
| B4 | exact fixture | constructs runtime only | success and recovery-incomplete entry latch | success assertion |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `engineRuntime` | exercise production constructor branches including paired schedule assembly | must return nil runtime on each constructor failure | AST |
| `Entry.CheckEntry` | prove construction arms restart recovery before loops run | exact recovery-incomplete refusal | assertion |

## State mutations and fallbacks

- Each case creates a fresh context, preventing one branch's latch or nil field from contaminating another.
- The schedule loader sees no production desired artifacts and performs no network read in this constructor-only fixture.

## Safety conclusion

- Safe edit boundary: update only the request-context argument required by the production runtime signature.
- High-risk impact: no — test-only proof of production fail-closed branches.
