# Function Logic Map: `ReadOnly.accountActiveQuarantines`

- Source: `internal/journal/account_views.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Base commit: `openspec/changes/a061-screens-show-what-they-already-know/base-commit.txt`

New in a061. It is a leaf read with no caller other than `LivePositionExits`.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `accountRef` | one account reference | `LivePositionExits` | unknown account -> empty map, not an error |
| `exit_snapshot_quarantines` | schemaV10 table; always present when the journal opened | `checkSchema` | a missing table would be a query error, which propagates |
| `positions.instance_seq` | the position's current generation | `positions` | joined, never assumed |

**Invariant**: only the current generation's unreleased quarantine is returned.
The active unique index is on `(position_id, position_generation)`, so a released
and re-adopted position keeps its old instance's row on disk unreleased forever --
nothing releases the quarantine of a generation that no longer exists. The engine
asks about the current generation only
(`exitloop.go`: `ActiveExitSnapshotQuarantine(ctx, p.ID, p.InstanceSeq)`), and
`q.position_generation = p.instance_seq` is what keeps a dead generation from
closing a live position's protection line.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | the query failed | none | **returns a wrapped error** | fail-closed; the caller degrades the ledger half |
| B2 | `for rows.Next()` | one map entry per active quarantine | none | `TestAQuarantinedPositionIsNotDrawnAsProtected` |
| B3 | a row scan failed | none | **returns a wrapped error** | fail-closed |
| B4 | `rows.Err()` | none | **returns a wrapped error** | fail-closed |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `r.db.QueryContext` | the one SELECT | error wrapped and propagated | AST |
| `rows.Scan` | six columns, all NOT NULL in the STRICT table | error wrapped and propagated | AST |

One SELECT on a local SQLite file opened read-only. No network, no lock, no
write, no retry.

## State mutations and fallbacks

- Builds and returns a map. Writes nothing.
- There is no fallback: every failure is returned. Returning a partial or empty
  map on error would tell the caller "no position is quarantined", which is the
  one wrong answer this read can give.

## Safety conclusion

- Safe edit boundary: a new read-only SELECT on an existing table.
- High-risk impact: no. Read-only handle, no schema change, no write path.
- Release and creation of quarantines stay where they are: the engine and the
  writable `Journal`. This function cannot do either.
