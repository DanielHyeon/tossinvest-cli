# Function Logic Map: `Tracker.Refresh`

- Source: `internal/reconcile/mismatch.go`
- Evidence: `ast.json`, `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Authority | Failure behavior |
|---|---|---|---|
| durable states + runtime pending blocks | configured account | journal under tracker mutex | authority error retains runtime projection and gate |

## Branches and early returns

| Branches | Condition | Result | Test |
|---|---|---|---|
| B1 | no journal | no-op | nil journal cases |
| B2 | authority read fails | unlock/error | Refresh error test |
| B3–B5 | scan account authority for durable permanent | durable row wins over memory proposal | durable Refresh test |
| B6–B8 | retain runtime pending blocks except continuity-broken non-durable account proposal | preserve ordinary pending; withdraw only disproved proposal | `TestA110RefreshDoesNotManufacturePermanentFromContinuityBrokenPendingProposal` |
| B9–B10 | rebuild matching-account durable rows | foreign rows skipped | isolation tests |
| B11 | durable permanent or disproved proposal clears pending retry identity | no stale retry | Refresh A110 cases |
| B12 | disproved proposal clears transient streak scalar | no manufactured threshold | `TestA110RefreshDoesNotManufacturePermanentFromContinuityBrokenPendingProposal` |
| B13 | durable permanent lifts compatibility scalar only | threshold display remains compatible | durable Refresh test |

## Calls and live bindings

| Callee | Purpose | Contract | Evidence |
|---|---|---|---|
| `ActiveReconcileStates` | durable authority | serialized with Observe | Refresh/Observe serialization test |
| `syncGate` | publish combined projection | pending remains fail-closed | pending tests |

## State mutations and fallbacks

- Replaces durable projection while retaining ordinary pending blocks, streaks and adjustment credits.
- A successful authority read that proves a continuity-broken permanent proposal was never durable removes only that proposal and its stale threshold-shaped streak view.

## Safety conclusion

- Safe boundary: preserve process-local streak behavior while deriving permanence only from a durable account row; withdraw only a proposal whose continuity was already broken and whose durable absence is now authoritative.
- High-risk impact: yes; driver refresh occurs immediately before adoption.
