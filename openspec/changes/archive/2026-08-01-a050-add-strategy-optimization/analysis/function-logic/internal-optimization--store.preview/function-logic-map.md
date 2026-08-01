# Function Logic Map: `Store.preview`

- Source: `internal/optimization/store.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| current snapshot, request, registry, evidence and capability entropy | request/category/source/reason/changes validate before insert; payload is capability-authenticated | current registry/snapshot and evidence provider | typed preview error; no candidate insert |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B20 | version/request/evidence/field/entropy/MAC/insert validation | stores one signed immutable candidate only after all checks | typed error or Preview | preview and hardening tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `currentSnapshot`, registry, evidence, `signCandidatePayload` | derives and signs canonical candidate payload | error aborts before insert | AST |

## State mutations and fallbacks

- A mixed-category or read-only configuration-error field is rejected before candidate persistence.
- The raw capability is not persisted; it is used only to HMAC the canonical payload.

## Safety conclusion

- Safe edit boundary: optimization-private candidate construction.
- High-risk impact: no LIVE authority.
