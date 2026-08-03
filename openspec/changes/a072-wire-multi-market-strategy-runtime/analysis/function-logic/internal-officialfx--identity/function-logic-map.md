# Function Logic Map: `Identity`

- Source: `internal/officialfx/evidence.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| identity snapshot | private sealed, canonical currency, bounded time window and snapshot identity | package-owned future snapshot loader | invalid/zero/tampered input refuses |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | snapshot seal/scope/window invalid | none | ErrInvalidEvidence | identity authority tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| identity seal validators | prove package-owned snapshot capability | pure | current source |

## State mutations and fallbacks

- Removes caller-supplied freshness; evidence window is copied from the opaque snapshot.

## Safety conclusion

- Safe edit boundary: same-currency rate and haircut remain exact canonical 1.
- High-risk impact: yes — KR q_final freshness authority.
