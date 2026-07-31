# Function Logic Map: `Journal.ExitEvents`

- Source: `internal/journal/exit_state.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| position id | append-only evaluation rows | journal v10 | SQL errors remain global; nullable evidence stays typed unknown |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B4 | query, scan loop, rows error, success | none | wrapped storage error or typed events | event read-model tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| SQLite query + `scanExitEvent` | restore stored evaluation without recomputation and validate typed arm-suppression evidence | corruption returns `ErrExitSnapshotCorrupt` | CodeGraph + AST |

## State mutations and fallbacks

- Reads exact saved/recomputed/effective JSON and scalar provenance columns.
- A nonempty arm-suppression reason must be the known enum and carry complete recomputed/effective orderable evidence with no action or intent.

## Safety conclusion

- Safe edit boundary: read-only projection.
- High-risk impact: yes.
