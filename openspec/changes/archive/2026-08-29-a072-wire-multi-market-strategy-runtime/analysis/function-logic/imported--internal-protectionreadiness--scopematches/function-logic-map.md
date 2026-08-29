# Function Logic Map: `scopeMatches`

- Source: `internal/protectionreadiness/attestation.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| signed body and runtime scope | every authority field exactly equal | signed body plus sealed runtime manifest | false produces typed scope mismatch |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | any account/profile/market/order/session/quantity bounds/trigger/replace/broker/tool/build/evidence field differs | none | false | exact scope matrix |
| B2 | all fields equal | none | true | valid signed fixture |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| none | direct exact equality including broker struct | no error/retry | CodeGraph + AST |

## State mutations and fallbacks

- Pure comparison; no defaults, coercion, or mutation.

## Safety conclusion

- Safe edit boundary: exact attestation/runtime intersection only
- High-risk impact: yes; authority must fail closed
