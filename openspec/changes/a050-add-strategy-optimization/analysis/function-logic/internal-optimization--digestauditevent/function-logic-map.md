# Function Logic Map: `digestAuditEvent`

- Source: `internal/optimization/store.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| audit event | validated positive row/version IDs, exact candidate/key/options/actor/reason/time | append-only audit row | domain-separated SHA-256 covers every persisted event field |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | branchless canonical JSON plus SHA-256 happy path | local allocation only | lowercase hex digest | audit tamper/migration tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| JSON marshal and SHA-256 | authenticates full immutable event with domain tag | fixed supported types; UTC time; deterministic | event-field tamper matrix |

## State mutations and fallbacks

- Pure helper; no field omission, repair, or legacy fallback on ordinary reads.

## Safety conclusion

- Safe edit boundary: full audit-event integrity digest.
- High-risk impact: yes; valid-looking audit row edits must fail closed.
