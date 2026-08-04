# Function Logic Map: `marketVerdictSeal`

- Source: `internal/protectionreadiness/dispatch.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| verdict and provenance | all fields that can authorize a protection mutation, including quantity bounds and capability scope | verified attestation | omitted or modified field changes seal and snapshot identity |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | all canonical fields | hashes length-prefixed preimage | deterministic SHA-256 | seal tamper tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `hashStrings` | collision-safe length-prefixed digest | pure/no retry | current HEAD |

## State mutations and fallbacks

- Pure; no external state. Field order is protocol and changes require dispatch tests.

## Safety conclusion

- Safe edit boundary: append every authority-bearing scope field to the market seal and paired snapshot seal.
- High-risk impact: yes — omission permits scope substitution.
