# Branch Test Map: `continueExistingOwner`

- Source: `internal/strategyarbiter/arbiter.go`; file SHA-256 `1788b0503479c4c4e8b4b17d7e2e6e2fd189f414ebde7877a7d5043157be6d03`. AST branch positions are authoritative.
- Rows carry measured counts from Go coverage profiles, count mode.
- arbiter untagged suite: `go test -count=1 -covermode=count -coverpkg=./internal/strategyarbiter ./internal/strategyarbiter/`
- arbiter tagged suite: `go test -count=1 -tags tossos_testseams -covermode=count -coverpkg=./internal/strategyarbiter ./internal/strategyarbiter/`
- engine tagged suite: `go test -count=1 -tags tossos_testseams -covermode=count -coverpkg=./internal/strategyarbiter,./internal/app/engine ./internal/app/engine/`
- engine untagged suite: `go test -count=1 -covermode=count -coverpkg=./internal/strategyarbiter,./internal/app/engine ./internal/app/engine/`
- Per-test attribution set: every `Test*` function in `internal/strategyarbiter`, each run alone under the seam tag. This is the whole package, not a sample.

Mutation receipts for this function (production source mutated, run, restored from a pristine copy taken before the mutation):

| # | mutation | result | killed by |
|---|---|---|---|
| M3 | let the score path run even when an active owner exists | KILLED | `TestAnActiveWeeklyOwnerIsNotReplacedByAHigherScore` |
| M12 | drop the campaign identity from the owner comparison | KILLED | `TestAProposalFromAnotherCampaignDoesNotInheritTheOwner` |

| Branch | Anchor | Measured disposition |
|---|---|---|
| B1 | if at 222:2 | arm not entered (arbiter untagged suite); arm entered 1x (arbiter tagged suite); arm not entered (engine tagged suite); arm not entered (engine untagged suite); entered by `TestAnActiveWeeklyOwnerIsNotReplacedByAHigherScore` |
| B2 | if at 227:2 | arm not entered (arbiter untagged suite); arm entered 1x (arbiter tagged suite); arm not entered (engine tagged suite); arm not entered (engine untagged suite); entered by `TestAProposalFromAnotherCampaignDoesNotInheritTheOwner` |
| B3 | if at 232:2 | arm not entered (arbiter untagged suite); arm entered 1x (arbiter tagged suite); arm not entered (engine tagged suite); arm not entered (engine untagged suite); entered by `TestAnActiveOwnerOnALaneWithNoScoreRowIsRefused` |

A row states what was measured, not what is intended. An arm recorded as not entered is a coverage gap, not a pass.
