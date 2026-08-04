# Function Logic Map: `ReadOnly.LivePositionExits`

- Source: `internal/journal/account_views.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Base commit: `openspec/changes/a067-screens-show-what-they-already-know/base-commit.txt`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `accountRef` | one account reference, trimmed | `AccountRefs` | unknown account -> no rows, not an error |
| `exit_states` of that account | includes completed rows | `accountExitStates` | query error -> returned, caller degrades the whole ledger half |
| `exit_snapshot_quarantines` (a067) | unreleased rows joined on the position's current `instance_seq` | `accountActiveQuarantines` | query error -> returned; the caller shows the broker half only, so the failure is fail-closed |
| `positions` of that account | `state <> CLOSED` | this query | query error -> returned |

**Invariant**: this is a compile-time read-only handle. Nothing here writes.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `accountExitStates` failed | none | **returns the error** | existing journal tests |
| **B2 (a067)** | `accountActiveQuarantines` failed | none | **returns the error** | fail-closed by construction; see below |
| B3 | the positions query failed | none | **returns a wrapped error** | existing journal tests |
| B4 | `for rows.Next()` | one `PositionExit` per live position | none | existing |
| B5 | `scanPosition` failed | none | **returns the error** | existing |
| **B6 (a067)** | an active quarantine exists for this position | attaches a copy to the row | falls through | `TestAQuarantinedPositionIsNotDrawnAsProtected` |
| B7 | an exit state exists for this position | attaches it, sets `HasExit` | falls through | existing |
| B8 | the state is pre-v10 with no policy identity | resolves the legacy identity in memory only | falls through | existing legacy tests |
| B9 | the legacy identity resolved | sets `PolicyIdentity` | falls through | existing |
| B10 | `else` it did not resolve | sets a typed unknown reason | falls through | existing |
| B11 | `rows.Err()` | none | **returns a wrapped error** | existing |

**Ordering**: B6 is placed before B7 deliberately. A quarantined position may have
no exit state at all, and the row still has to carry the quarantine.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `accountExitStates` | the protection state of every position | error propagates | AST |
| `accountActiveQuarantines` (a067) | which positions the engine has taken out of judgement | error propagates; no partial answer | AST |
| `scanPosition` | one projected position | error propagates | AST |
| `legacyPolicyIdentity` | pre-v10 identity resolution, in memory only | error becomes a typed unknown reason | AST |

Three SELECTs on a local SQLite file opened read-only. No network, no lock, no
broker call. a067 adds the third; it does not change the first two.

## State mutations and fallbacks

- Builds and returns a slice. Writes nothing, in memory or on disk.
- The quarantine is copied into a fresh variable before its address is taken, so
  no two rows can share one loop variable.
- A quarantine read failure fails the whole account read rather than returning
  rows with unknown quarantine status. That is the fail-closed direction: the
  caller renders the broker half and every protection line closes, instead of
  drawing a line for a position that might be out of judgement.

## Safety conclusion

- Safe edit boundary: one additional read-only SELECT and one pointer copy.
- High-risk impact: no. This is a read path on a read-only handle; the ledger
  schema is unchanged and nothing is written.
- Safety invariant 0.6 (ledger schema) not engaged: no migration, no new column.
