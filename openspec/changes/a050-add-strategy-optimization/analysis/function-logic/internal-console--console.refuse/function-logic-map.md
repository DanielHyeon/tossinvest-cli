# Function Logic Map: `Console.refuse`

- Source: `internal/console/pages.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| status/title/detail | explicit refusal status and server-created text | handler | hardened Korean refusal document, no redirect or mutation |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | any refusal | HTTP response only | exact requested status with normal security headers | optimization 400/412/425 security test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Console.renderHTML` | makes refusal headers identical to normal HTML | no retry | CodeGraph + AST |

## State mutations and fallbacks

- No domain mutation; refusal paths cannot invoke retry.

## Safety conclusion

- Safe edit boundary: status-preserving hardened HTML response.
- High-risk impact: yes; CSP, referrer and nosniff now remain present on errors.
