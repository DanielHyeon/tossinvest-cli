# Function Logic Map: `readinessSnapshotSeal`

- Source: `internal/protectionreadiness/readiness.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| paired snapshot | KR and US exact verdicts | `Assess` | corrupt snapshot reads UNWIRED |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | each market verdict | hash only | sealed digest | provenance tamper test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| SHA-256 helpers | canonical immutable seal | no I/O | AST |

## State mutations and fallbacks

- Hash computation only; no fallback and no authority.

## Safety conclusion

- Safe edit boundary: bind newly exposed provenance fields into the seal.
- High-risk impact: yes; omitted scope fields would permit undetected drift.
