# Function Logic Map: `canonicalTriggerSQL`

- Source: `internal/optimization/store.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| trigger SQL | persisted or manifest trigger definition | sqlite schema/manifest | only whitespace, trailing semicolon, and install-time `IF NOT EXISTS` are normalized |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | branchless happy path canonicalizes syntax-preserving variations | local string only | canonical SQL | trigger drift test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `TrimSuffix`, `Fields`, `Join`, `Replace` | narrow canonical comparison | pure | same-name drift tests |

## State mutations and fallbacks

- No schema mutation. Trigger body/table/event semantics are not erased by canonicalization.

## Safety conclusion

- Safe edit boundary: narrow SQL normalization for integrity comparison.
- High-risk impact: yes; over-normalization could accept a weakened trigger.
