# Function Logic Map: `canonicalProtectionAccount`

- Source: `internal/attest/protection_matrix.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| account ref | exactly 8-14 ASCII digits, no separators or padding | signed/runtime scope contract | invalid account/scope error |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B3 | nonempty/trimmed; every byte digit; length 8-14 | local byte scan | error or unchanged canonical string | account grammar table |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| none | deliberately avoids legacy normalization | no fallback | CodeGraph + AST |

## State mutations and fallbacks

- None.

## Safety conclusion

- Safe edit boundary: remove hyphen normalization aliases.
- High-risk impact: yes.
