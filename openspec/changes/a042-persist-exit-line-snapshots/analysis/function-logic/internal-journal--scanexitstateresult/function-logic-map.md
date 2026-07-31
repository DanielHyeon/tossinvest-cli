# Function Logic Map: `scanExitStateResult`

- Source: `internal/journal/exit_snapshot.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| one `exit_states` row with nullable v10 columns | all v10 NULL is legacy; SEED has no output evidence; EVALUATED is complete and coherent | additive schema evidence, never current-policy recomputation | per-row typed `ErrExitSnapshotCorrupt` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | SQL no-row/scan failure | none | typed not-found/driver error | state read tests |
| B2 | every v10 column NULL | resolve only pinned legacy identity in memory | legacy unknown view | legacy no-backfill test |
| B3 | status absent but any v10 column present | none | `partial_snapshot_tuple` | single-column evidence test |
| B4 | policy identity tuple partial/invalid | none | typed corruption | corruption matrix |
| B5 | SEED carries any evaluation/output column | none | `partial_seed_tuple` | per-column SEED table |
| B6 | EVALUATED tuple incomplete | none | `partial_evaluated_tuple` | corruption matrix |
| B7 | JSON invalid or flattened values differ | none | typed corruption | digest/flattened forgery tests |
| B8 | complete exact tuple | attach stored snapshot | success | persistence/reopen tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `decodeStoredSnapshot` | validate sealed semantic snapshot | exact error becomes per-row corruption | CodeGraph + AST |
| `legacyPolicyIdentity` | resolve only pinned a041 compatibility meanings | unknown remains typed unknown; no backfill | CodeGraph + AST |

## State mutations and fallbacks

- Read-only scan; it never writes, repairs, or recomputes missing output.

## Safety conclusion

- Safe edit boundary: nullable v10 decoding and corruption classification.
- High-risk impact: yes; fail-open legacy classification would bypass quarantine.
