# Function Logic Map: `boolString`

- Source: `internal/protectionreadiness/types.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| boolean | true or false | broker capability field | canonical lowercase token |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | true | none | `"true"` | digest fixture |
| B2 | false | none | `"false"` | capability substitution |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| none | canonical conversion | no error/retry | CodeGraph + AST |

## State mutations and fallbacks

- Pure conversion.

## Safety conclusion

- Safe edit boundary: stable lowercase tokens
- High-risk impact: no
