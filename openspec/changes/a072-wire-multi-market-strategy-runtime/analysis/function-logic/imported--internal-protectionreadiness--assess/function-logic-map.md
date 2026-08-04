# Function Logic Map: `Assess`

- Source: `internal/protectionreadiness/readiness.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| policy/state/time | sealed and monotonic | pinned policy and durable state | per-market UNWIRED |
| market evidence | exact signed KR or US scope | attestation + supervisor binding | peer market unchanged |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B3 | durable state/time floor validity | pure next-state only | fail closed | state tests |
| B4-B14 | each market evidence and verification result | per-market verdict | typed refusal | KR/US isolation tests |
| B15-B17 | state commit allowed | reseal or preserve preimage | assessment result | rollback tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `verifyAttestation` | signature/scope/supervisor proof | no retry or external mutation | CodeGraph + AST |

## State mutations and fallbacks

- No external mutation; result contains a pure durable-state successor and immutable paired snapshot.

## Safety conclusion

- Safe edit boundary: add exact account/profile/supervisor provenance to already-verified verdicts only.
- High-risk impact: yes; every new field is included in the snapshot seal.
