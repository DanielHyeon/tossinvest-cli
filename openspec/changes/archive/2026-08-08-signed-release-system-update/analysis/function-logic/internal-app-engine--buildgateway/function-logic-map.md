# Function Logic Map: `buildGateway`

- Source: `internal/app/engine/gateway.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| gateway inputs | real journal/trading/official/account/clock | `NewContext` | validate/restore/construct refusal |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | input validation fails | none | error | gateway tests |
| B2 | reconciliation restore fails | journal read only | wrapped error | restart recovery tests |
| B3 | execution gateway construction fails | no published wiring | wrapped error | gateway tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Tracker.Restore` | rebuild reconcile blocks | fail closed | CodeGraph + AST |
| `execgw.New` | create sole mutation gateway | validates wiring | CodeGraph + AST |
| notifier/retrier constructors | bind observation/alert paths | returned together | CodeGraph |

## State mutations and fallbacks

- This change only applies gofmt alignment to existing option fields; logic is unchanged.

## Safety conclusion

- Safe edit boundary: mechanical formatting only.
- High-risk impact: yes — sole official order gateway assembly, covered by full/race suites.
