# Function Logic Map: `Tracker.Restore`

- Source: `internal/reconcile/mismatch.go`
- Evidence: `ast.json`, `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Authority | Failure behavior |
|---|---|---|---|
| active reconcile rows | current account plus foreign rows for gate rebuild | journal | read error leaves prior runtime state unchanged |

## Branches and early returns

| Branches | Condition | Result | Test |
|---|---|---|---|
| B1 | no journal | no-op | nil journal tests |
| B2 | authority read fails | wrapped error | restore failure test |
| B3–B4 | row belongs to another configured account | skip tracker ownership | account isolation tests |
| B5 | durable account quantity row | permanent projection | durable permanent restart test |
| B6 | permanent restored | compatibility failures set to threshold | restore compatibility test |

## Calls and live bindings

| Callee | Purpose | Contract | Evidence |
|---|---|---|---|
| `ActiveReconcileStates` | restart authority | all active rows or error | B2 |
| `blockFromReconcileState` | durable shape projection | empty symbol account-wide | blank-symbol prevention test |
| `EntryGate.RebuildReconcileProjection` | rebuild all producer gates | preserves foreign causes | restore suites |

## State mutations and fallbacks

- Durable blocks replace memory; process-local streaks, pending identity and credits are deliberately cleared.
- Only an already durable account-wide quantity row restores permanent.

## Safety conclusion

- Safe boundary: comments/fields align transient streak non-reconstruction with durable operator-only authority.
- High-risk impact: yes; restart must neither invent nor lose a permanent block.
