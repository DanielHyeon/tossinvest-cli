# Function Logic Map: `RequiredEndpoints`

- Source: `internal/app/engine/interlock.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| static engine dependency catalog | exact method/path for every engine read and mutation | official client calls plus a072 production FX service | capability attestation missing any item refuses automation startup |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | direct return | none; returns a fresh slice including exchange-rate read | complete required endpoint set | endpoint coverage tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| none | static catalog only | no error, timeout, retry, I/O or fallback | CodeGraph + AST |

## State mutations and fallbacks

- No durable or network mutation. Adding the official exchange read intentionally invalidates older
  attestations until separately executed OAuth evidence covers the new endpoint.

## Safety conclusion

- Safe edit boundary: append one read-only dependency without removing or weakening existing requirements.
- High-risk impact: **yes** — startup capability attestation is fail-closed.
