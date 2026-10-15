# Focused mutation receipt

- Scope: local temporary worktree only; no live request, order, or service access.
- Command: `go test ./internal/client -run 'TestListPendingOrders(Rejects|Accepts)|TestListCompletedOrdersRange(Rejects|Accepts)|TestListWatchlist(FromFixtures|RejectsInvalidCollections)' -count=1`
- Control: focused command passes at `c8ab315a`.
- Mutants: inverted the pending `result`, completed `result.body`, and bulk `watchlists` presence guards in the temporary copy.
- Result: command exited `1`; both explicit-empty success tests and absent/null failure tests failed. The fixture suite therefore reaches each changed presence guard.

The temporary mutant copy is outside this repository and is not source evidence.
