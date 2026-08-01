# Function Logic Map: `Console.render`

- Source: `internal/console/pages.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| template name/data | parsed server template and typed view | console renderer | delegated renderer emits either complete 200 HTML or a render error |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | ordinary page render | HTTP response only | delegates status/CSP/header contract | console page and security header tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Console.renderHTML` | centralizes successful HTML headers | no retry | CodeGraph + AST |

## State mutations and fallbacks

- No domain mutation; only an HTTP response is written.

## Safety conclusion

- Safe edit boundary: normal 200 rendering delegates to the shared hardened writer.
- High-risk impact: no.
