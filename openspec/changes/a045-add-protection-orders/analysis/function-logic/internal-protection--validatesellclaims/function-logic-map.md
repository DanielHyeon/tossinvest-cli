# Function Logic Map: `ValidateSellClaims`

- Source: `internal/protection/domain.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Function inputs | Nonnegative int64 holding and three nonnegative sell claims. | Current HEAD + OpenSpec | Fail closed with typed error/decision |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1+ | B1 negative value; B2 arithmetic total exceeds holding; else accept. | No mutation. Existing sum can overflow int64 and incorrectly admit oversell. | Typed refusal or validated result | See branch map |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| Current callees | No callees or side effects. | No implicit retry; errors propagate fail-closed | CodeGraph + AST |

## State mutations and fallbacks

- No mutation. Existing sum can overflow int64 and incorrectly admit oversell.

## Safety conclusion

- Safe edit boundary: Never add claims; subtract each claim from remaining holding after bounds checks.
- High-risk impact: yes; dormant logic only, no broker mutation or WIRED binding.
