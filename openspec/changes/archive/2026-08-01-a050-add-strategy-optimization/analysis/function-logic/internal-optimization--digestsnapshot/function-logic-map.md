# Function Logic Map: `digestSnapshot`

- Source: `internal/optimization/store.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| snapshot | validated immutable lifecycle metadata | insert/scan paths | deterministic SHA-256 covers version/effective version, desired/effective maps, evidence/manifest, flags, actor/reason/audit/time |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | branchless happy path canonical-marshals fixed typed payload and hashes bytes | local allocation only | lowercase hex SHA-256 | snapshot metadata coverage test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| JSON marshal | deterministic encoding of fixed typed payload/maps | types are JSON-supported; no fallback omission | metadata tamper matrix |
| SHA-256 | immutable snapshot integrity identity | deterministic | digest corruption tests |

## State mutations and fallbacks

- No persistence or external state. Created time is normalized to UTC before hashing; every immutable field is explicit.

## Safety conclusion

- Safe edit boundary: complete snapshot integrity digest.
- High-risk impact: yes; omitted metadata could be tampered without detection.
