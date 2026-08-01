# Function Logic Map: `Store.snapshot`

- Source: `internal/optimization/store.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| version and stored snapshot columns | one complete, canonical, internally consistent immutable snapshot with matching digest | control DB | reject corrupted snapshot rather than return partial state |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | no row/scan error | none | database error | existing reopen tests |
| B2 | malformed JSON, booleans, versions, actor/audit ID | none | corrupt snapshot error | `TestSnapshotAndAuditCorruptionFailClosed` |
| B3 | invalid timestamp or settings digest | none | corrupt snapshot error | `TestSnapshotAndAuditCorruptionFailClosed` |
| B4 | valid row | clone maps for caller | immutable View projection | lifecycle tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| SQLite query row | fetches one version | scan errors propagate | AST B1 |
| `parseStoredTime`, `digestSnapshot` | validates canonical timestamp/integrity | any mismatch fails closed | AST B2-B3 |

## State mutations and fallbacks

- The result is cloned; callers cannot mutate database-backed map state.

## Safety conclusion

- Safe edit boundary: private snapshot read integrity.
- High-risk impact: no LIVE authority.
