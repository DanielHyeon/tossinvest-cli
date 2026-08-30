# Branch Test Map: `ownerSnapshotFault`

- Source: `internal/strategyarbiter/arbiter.go`; file SHA-256 `ba484e68bd49e73081afaf031129c1a418af4103a436b94382b4105b2f68da2a`. AST branch positions are authoritative.
- Rows carry measured counts from Go coverage profiles, count mode.
- arbiter untagged suite: `go test -count=1 -covermode=count -coverpkg=./internal/strategyarbiter ./internal/strategyarbiter/`
- arbiter tagged suite: `go test -count=1 -tags tossos_testseams -covermode=count -coverpkg=./internal/strategyarbiter ./internal/strategyarbiter/`
- engine tagged suite: `go test -count=1 -tags tossos_testseams -covermode=count -coverpkg=./internal/strategyarbiter,./internal/app/engine ./internal/app/engine/`
- engine untagged suite: `go test -count=1 -covermode=count -coverpkg=./internal/strategyarbiter,./internal/app/engine ./internal/app/engine/`
- Per-test attribution set: every `Test*` function in `internal/strategyarbiter`, each run alone under the seam tag. This is the whole package, not a sample.

Mutation receipts for this function (production source mutated, run, restored from a pristine copy taken before the mutation):

| # | mutation | result | killed by |
|---|---|---|---|
| M23 | drop the classifier so every route-set failure reports one code | KILLED | `TestAStaleOwnerRevisionIsItsOwnRefusal`, `TestTwoActiveOwnersAreItsOwnRefusal` |
| M24 | report multiple active owners as a stale owner revision | KILLED | `TestTwoActiveOwnersAreItsOwnRefusal` |
| M25 | report a stale owner revision as a score tie | KILLED | `TestAStaleOwnerRevisionIsItsOwnRefusal` |

| Branch | Anchor | Measured disposition |
|---|---|---|
| B1 | range at 196:2 | arm not entered (arbiter untagged suite); arm entered 3x (arbiter tagged suite); arm not entered (engine tagged suite); arm not entered (engine untagged suite); entered by `TestAStaleOwnerRevisionIsItsOwnRefusal`, `TestTwoActiveOwnersAreItsOwnRefusal` |
| B2 | if at 197:3 | arm not entered (arbiter untagged suite); arm entered 3x (arbiter tagged suite); arm not entered (engine tagged suite); arm not entered (engine untagged suite); entered by `TestAStaleOwnerRevisionIsItsOwnRefusal`, `TestTwoActiveOwnersAreItsOwnRefusal` |
| B3 | if at 201:2 | arm not entered (arbiter untagged suite); arm entered 1x (arbiter tagged suite); arm not entered (engine tagged suite); arm not entered (engine untagged suite); entered by `TestTwoActiveOwnersAreItsOwnRefusal` |
| B4 | if at 204:2 | arm not entered (arbiter untagged suite); arm entered 1x (arbiter tagged suite); arm not entered (engine tagged suite); arm not entered (engine untagged suite); entered by `TestAStaleOwnerRevisionIsItsOwnRefusal` |
| B5 | if at 207:2 | arm not entered (arbiter untagged suite); arm entered 1x (arbiter tagged suite); arm not entered (engine tagged suite); arm not entered (engine untagged suite); entered by `TestAnOwnerSnapshotOutsideItsFreshnessWindowIsAStaleOwner` |

A row states what was measured, not what is intended. An arm recorded as not entered is a coverage gap, not a pass.
