# Branch Test Map: `selectHighestScore`

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
| M5 | drop the eligible-lane membership check | KILLED | `TestALaneOutsideTheSealedEligibleSetIsRefused` |
| M1 | delete the top-score tie check | KILLED | `TestATieAtTheTopIsRefusedRatherThanBrokenArbitrarily` |
| M19 | let a new leader keep the previous tie count (`1` to `ties+1`) | KILLED | `TestATieBelowTheTopStillLeavesAUniqueWinner` |
| M20 | drop the evidence/config digest binding to the route decision | KILLED | `TestAProposalBuiltFromOtherEvidenceIsRefused` |
| M21 | bind only the evidence digest, not the config digest | KILLED | `TestAProposalBuiltFromOtherEvidenceIsRefused` |
| M22 | bind only the config digest, not the evidence digest | KILLED | `TestAProposalBuiltFromOtherEvidenceIsRefused` |
| B8-dead | `if best < 0` measured as never entered in every suite | REMOVED as unreachable — `Arbitrate` refuses an empty proposal list before calling | n/a |

| Branch | Anchor | Measured disposition |
|---|---|---|
| B1 | range at 242:2 | arm not entered (arbiter untagged suite); arm entered 21x (arbiter tagged suite); arm entered 12x (engine tagged suite); arm not entered (engine untagged suite); entered by `TestALaneMatchingTwoScoreRowsIsUnknown`, `TestALaneOutsideTheSealedEligibleSetIsRefused`, `TestALaneWithNoFamilyScoreRowIsUnknown`, `TestAProposalBuiltFromOtherEvidenceIsRefused`, `TestAScoreAboveTheApprovedCeilingIsRefused`, `TestAScoreRowNamingAnUnapprovedFamilyIsUnknown`, `TestASingletonProposalWithoutApprovedScoreAuthorityIsRefused`, `TestAStaleProposalClosesTheWholeScope`, `TestATieAtTheTopIsRefusedRatherThanBrokenArbitrarily`, `TestATieBelowTheTopStillLeavesAUniqueWinner`, `TestThreeFamiliesOnOneSymbolYieldTheSingleHighestScore` |
| B2 | range at 247:2 | arm not entered (arbiter untagged suite); arm entered 22x (arbiter tagged suite); arm entered 12x (engine tagged suite); arm not entered (engine untagged suite); entered by `TestALaneMatchingTwoScoreRowsIsUnknown`, `TestALaneOutsideTheSealedEligibleSetIsRefused`, `TestALaneWithNoFamilyScoreRowIsUnknown`, `TestAProposalBuiltFromOtherEvidenceIsRefused`, `TestAScoreAboveTheApprovedCeilingIsRefused`, `TestAScoreRowNamingAnUnapprovedFamilyIsUnknown`, `TestASingletonProposalWithoutApprovedScoreAuthorityIsRefused`, `TestAStaleProposalClosesTheWholeScope`, `TestATieAtTheTopIsRefusedRatherThanBrokenArbitrarily`, `TestATieBelowTheTopStillLeavesAUniqueWinner`, `TestThreeFamiliesOnOneSymbolYieldTheSingleHighestScore` |
| B3 | if at 250:3 | arm not entered (arbiter untagged suite); arm entered 1x (arbiter tagged suite); arm not entered (engine tagged suite); arm not entered (engine untagged suite); entered by `TestALaneOutsideTheSealedEligibleSetIsRefused` |
| B4 | if at 256:3 | arm not entered (arbiter untagged suite); arm entered 3x (arbiter tagged suite); arm not entered (engine tagged suite); arm not entered (engine untagged suite); entered by `TestAProposalBuiltFromOtherEvidenceIsRefused` |
| B5 | if at 260:3 | arm not entered (arbiter untagged suite); arm entered 4x (arbiter tagged suite); arm not entered (engine tagged suite); arm not entered (engine untagged suite); entered by `TestALaneMatchingTwoScoreRowsIsUnknown`, `TestALaneWithNoFamilyScoreRowIsUnknown`, `TestAScoreAboveTheApprovedCeilingIsRefused`, `TestAScoreRowNamingAnUnapprovedFamilyIsUnknown` |
| B6 | switch at 263:3 | arm not entered (arbiter untagged suite); arm entered 10x (arbiter tagged suite); arm entered 10x (engine tagged suite); arm not entered (engine untagged suite); entered by `TestALaneOutsideTheSealedEligibleSetIsRefused`, `TestAProposalBuiltFromOtherEvidenceIsRefused`, `TestAScoreAboveTheApprovedCeilingIsRefused`, `TestASingletonProposalWithoutApprovedScoreAuthorityIsRefused`, `TestAStaleProposalClosesTheWholeScope`, `TestATieAtTheTopIsRefusedRatherThanBrokenArbitrarily`, `TestATieBelowTheTopStillLeavesAUniqueWinner`, `TestThreeFamiliesOnOneSymbolYieldTheSingleHighestScore` |
| B7 | case at 264:3 | arm not entered (arbiter untagged suite); arm entered 10x (arbiter tagged suite); arm entered 10x (engine tagged suite); arm not entered (engine untagged suite); entered by `TestALaneOutsideTheSealedEligibleSetIsRefused`, `TestAProposalBuiltFromOtherEvidenceIsRefused`, `TestAScoreAboveTheApprovedCeilingIsRefused`, `TestASingletonProposalWithoutApprovedScoreAuthorityIsRefused`, `TestAStaleProposalClosesTheWholeScope`, `TestATieAtTheTopIsRefusedRatherThanBrokenArbitrarily`, `TestATieBelowTheTopStillLeavesAUniqueWinner`, `TestThreeFamiliesOnOneSymbolYieldTheSingleHighestScore` |
| B8 | case at 266:3 | arm not entered (arbiter untagged suite); arm entered 1x (arbiter tagged suite); arm entered 1x (engine tagged suite); arm not entered (engine untagged suite); entered by `TestATieAtTheTopIsRefusedRatherThanBrokenArbitrarily` |
| B9 | if at 274:2 | arm not entered (arbiter untagged suite); arm entered 1x (arbiter tagged suite); arm entered 1x (engine tagged suite); arm not entered (engine untagged suite); entered by `TestATieAtTheTopIsRefusedRatherThanBrokenArbitrarily` |

A row states what was measured, not what is intended. An arm recorded as not entered is a coverage gap, not a pass.
