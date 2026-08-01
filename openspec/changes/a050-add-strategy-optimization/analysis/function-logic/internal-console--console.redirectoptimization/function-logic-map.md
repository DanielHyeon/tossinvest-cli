# Function Logic Map: `Console.redirectOptimization`

- Source: `internal/console/optimization.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| notice | server-created operator result text | handler | URL escaped |
| category | current valid category | handler/registry | omitted or canonical ID only |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | always | response redirect only | 303 to optimization deep link | redirect escaping test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `url.QueryEscape` | encodes notice | no error contract | CodeGraph + AST |
| `http.Redirect` | PRG after POST | 303, no retry | CodeGraph + AST |

## State mutations and fallbacks

- No persistent mutation; carries only result notice and canonical category.

## Safety conclusion

- Safe edit boundary: fixed same-origin redirect.
- High-risk impact: no; output encoding and redirect destination are still tested.
