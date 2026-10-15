# Focused mutation receipt

- Scope: local temporary worktree only; no live request, order, or service access.
- Command: `go test ./internal/client -run 'TestListPendingOrders(Rejects|Accepts)|TestListCompletedOrdersRange(Rejects|Accepts)|TestListWatchlist(FromFixtures|RejectsInvalidCollections)' -count=1`
- Control: focused command passes at `c8ab315a`.
- Independent mutants: one temporary copy per guard, replacing its condition with `false`:
  - pending `result`: `TestListPendingOrdersRejectsAbsentOrNullResult` exited `1` on absent and null.
  - completed `result.body`: missing-collection and partial-market tests exited `1` on absent/null body.
  - bulk `watchlists`: `TestListWatchlistRejectsInvalidCollections` exited `1` on missing and null.
  - each group `items`: the same test exited `1` on missing items.
- Result: every changed guard was independently killed by the corresponding negative fixture. The combined inversion also failed, but is not the guard-level evidence.

The temporary mutant copy is outside this repository and is not source evidence.
