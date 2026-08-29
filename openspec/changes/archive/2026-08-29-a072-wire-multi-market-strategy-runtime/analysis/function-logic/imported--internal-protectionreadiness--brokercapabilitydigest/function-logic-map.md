# Function Logic Map: `brokerCapabilityDigest`

- Source: `internal/protectionreadiness/types.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| broker capability | exact eight-field capability tuple | attested manifest/body | deterministic SHA-256 digest |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | any capability field differs | none | different digest | dispatch substitution matrix |
| B2 | identical capability | none | identical digest | valid dispatch fixture |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `hashStrings`, `boolString`, `hexBytes` | unambiguous deterministic binding | no error/retry | CodeGraph + AST |

## State mutations and fallbacks

- Pure digest; no mutation or defaults.

## Safety conclusion

- Safe edit boundary: include every capability field in stable order
- High-risk impact: yes; dispatch authority binding
