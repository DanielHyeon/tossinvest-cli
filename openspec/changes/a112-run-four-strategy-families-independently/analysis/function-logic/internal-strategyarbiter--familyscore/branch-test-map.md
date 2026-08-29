# Branch Test Map: `familyScore`

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
| M6 | drop the score ceiling check | KILLED | `TestAScoreAboveTheApprovedCeilingIsRefused` |
| M10 | accept a family name outside the approved enum | KILLED | `TestAScoreRowNamingAnUnapprovedFamilyIsUnknown` |
| M11 | accept a lane that matches two score rows (`!= 1` to `< 1`) | KILLED | `TestALaneMatchingTwoScoreRowsIsUnknown` |

| Branch | Anchor | Measured disposition |
|---|---|---|
| B1 | range at 287:2 | arm not entered (arbiter untagged suite); arm entered 39x (arbiter tagged suite); arm entered 46x (engine tagged suite); arm not entered (engine untagged suite); entered by `TestALaneMatchingTwoScoreRowsIsUnknown`, `TestALaneOutsideTheSealedEligibleSetIsRefused`, `TestALaneWithNoFamilyScoreRowIsUnknown`, `TestAProposalBuiltFromOtherEvidenceIsRefused`, `TestAScoreAboveTheApprovedCeilingIsRefused`, `TestAScoreRowNamingAnUnapprovedFamilyIsUnknown`, `TestASingletonProposalWithoutApprovedScoreAuthorityIsRefused`, `TestAStaleProposalClosesTheWholeScope`, `TestATieAtTheTopIsRefusedRatherThanBrokenArbitrarily`, `TestATieBelowTheTopStillLeavesAUniqueWinner`, `TestAnActiveOwnerOnALaneWithNoScoreRowIsRefused`, `TestAnActiveWeeklyOwnerIsNotReplacedByAHigherScore`, `TestThreeFamiliesOnOneSymbolYieldTheSingleHighestScore` |
| B2 | if at 288:3 | arm not entered (arbiter untagged suite); arm entered 19x (arbiter tagged suite); arm entered 12x (engine tagged suite); arm not entered (engine untagged suite); entered by `TestALaneMatchingTwoScoreRowsIsUnknown`, `TestALaneOutsideTheSealedEligibleSetIsRefused`, `TestAProposalBuiltFromOtherEvidenceIsRefused`, `TestAScoreAboveTheApprovedCeilingIsRefused`, `TestAScoreRowNamingAnUnapprovedFamilyIsUnknown`, `TestASingletonProposalWithoutApprovedScoreAuthorityIsRefused`, `TestAStaleProposalClosesTheWholeScope`, `TestATieAtTheTopIsRefusedRatherThanBrokenArbitrarily`, `TestATieBelowTheTopStillLeavesAUniqueWinner`, `TestAnActiveWeeklyOwnerIsNotReplacedByAHigherScore`, `TestThreeFamiliesOnOneSymbolYieldTheSingleHighestScore` |
| B3 | if at 293:2 | arm not entered (arbiter untagged suite); arm entered 4x (arbiter tagged suite); arm not entered (engine tagged suite); arm not entered (engine untagged suite); entered by `TestALaneMatchingTwoScoreRowsIsUnknown`, `TestALaneWithNoFamilyScoreRowIsUnknown`, `TestAScoreRowNamingAnUnapprovedFamilyIsUnknown`, `TestAnActiveOwnerOnALaneWithNoScoreRowIsRefused` |
| B4 | if at 296:2 | arm not entered (arbiter untagged suite); arm entered 1x (arbiter tagged suite); arm not entered (engine tagged suite); arm not entered (engine untagged suite); entered by `TestAScoreAboveTheApprovedCeilingIsRefused` |

A row states what was measured, not what is intended. An arm recorded as not entered is a coverage gap, not a pass.
