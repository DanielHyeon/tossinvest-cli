# Branch Test Map: `Arbitrate`

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
| M1 | delete the top-score tie check | KILLED | `TestATieAtTheTopIsRefusedRatherThanBrokenArbitrarily` |
| M2 | skip the calibration requirement when only one proposal exists | SURVIVED — the checked value was dead code (`SealsValid()` already rejects an empty calibration); the check was removed rather than re-tested | n/a |
| M2a | drop the `SealsValid()` conjunct | KILLED | `TestASingletonProposalWithoutApprovedScoreAuthorityIsRefused` |
| M2b | drop the calibration-equality conjunct | KILLED | `TestProposalsUnderDifferentScoreVersionsAreIncomparable` |
| M3 | let the score path run even when an active owner exists | KILLED | `TestAnActiveWeeklyOwnerIsNotReplacedByAHigherScore` |
| M4 | treat the expiry instant as still fresh (`<=` to `<`) | KILLED | `TestAStaleProposalClosesTheWholeScope` |
| M7 | drop the duplicate-lane check | KILLED | `TestTheSameLaneTwiceIsRefused` |
| M8 | drop the route-set agreement check | KILLED | `TestProposalsBoundToDifferentRouteSetsAreRefused` |
| M9 | drop the proposal seal check | KILLED | `TestAProposalMutatedAfterSealingIsRefused` |
| M13 | let a refused outcome name index 0 instead of -1 | KILLED | `TestATieAtTheTopIsRefusedRatherThanBrokenArbitrarily` |
| M14 | drop position generation from the lineage scope check | KILLED | `TestAProposalWhoseLineageLeavesTheScopeIsRefused` |
| M15 | drop symbol from the lineage scope check | KILLED | `TestAProposalWhoseLineageLeavesTheScopeIsRefused` |
| M16 | drop account from the lineage scope check | KILLED | `TestAProposalWhoseLineageLeavesTheScopeIsRefused` |
| M17 | drop market from the lineage scope check | KILLED | `TestAProposalWhoseLineageLeavesTheScopeIsRefused` |
| M18 | drop the authority owner-key check | KILLED | `TestAProposalMeasuredAgainstAnotherSymbolsAuthorityIsRefused` |
| M0 | make the package import `internal/journal` | KILLED | `TestTheArbiterCannotReachAnyMutationCapability` |
| M23 | drop the owner-snapshot classifier so every route-set failure reports one code | KILLED | `TestAStaleOwnerRevisionIsItsOwnRefusal`, `TestTwoActiveOwnersAreItsOwnRefusal` |
| M26 | rename a contract code off the frozen golden (`ARBITRATION_TIE` to `ARBITRATION_SCORE_TIE`) | KILLED | `TestTheRefusalContractIsExactlyTheFrozenGolden` |

| Branch | Anchor | Measured disposition |
|---|---|---|
| B1 | if at 120:2 | arm entered 5x (arbiter untagged suite); arm entered 5x (arbiter tagged suite); arm not entered (engine tagged suite); arm not entered (engine untagged suite); entered by `TestARequestWithoutAnIdentityIsRefused` |
| B2 | if at 124:2 | arm entered 1x (arbiter untagged suite); arm entered 1x (arbiter tagged suite); arm not entered (engine tagged suite); arm not entered (engine untagged suite); entered by `TestNoProposalIsNotASelection` |
| B3 | if at 135:2 | arm entered 1x (arbiter untagged suite); arm entered 4x (arbiter tagged suite); arm not entered (engine tagged suite); arm not entered (engine untagged suite); entered by `TestAStaleOwnerRevisionIsItsOwnRefusal`, `TestAZeroProposalIsNeverSelected`, `TestAnOwnerSnapshotOutsideItsFreshnessWindowIsAStaleOwner`, `TestTwoActiveOwnersAreItsOwnRefusal` |
| B4 | if at 136:3 | arm not entered (arbiter untagged suite); arm entered 3x (arbiter tagged suite); arm not entered (engine tagged suite); arm not entered (engine untagged suite); entered by `TestAStaleOwnerRevisionIsItsOwnRefusal`, `TestAnOwnerSnapshotOutsideItsFreshnessWindowIsAStaleOwner`, `TestTwoActiveOwnersAreItsOwnRefusal` |
| B5 | range at 141:2 | arm not entered (arbiter untagged suite); arm entered 47x (arbiter tagged suite); arm entered 14x (engine tagged suite); arm not entered (engine untagged suite); entered by `TestALaneMatchingTwoScoreRowsIsUnknown`, `TestALaneOutsideTheSealedEligibleSetIsRefused`, `TestALaneWithNoFamilyScoreRowIsUnknown`, `TestAProposalBuiltFromOtherEvidenceIsRefused`, `TestAProposalFromAnotherCampaignDoesNotInheritTheOwner`, `TestAProposalMeasuredAgainstAnotherSymbolsAuthorityIsRefused`, `TestAProposalMutatedAfterSealingIsRefused`, `TestAProposalOutsideTheExpectedScopeIsRefused`, `TestAProposalWhoseLineageLeavesTheScopeIsRefused`, `TestAScoreAboveTheApprovedCeilingIsRefused`, `TestAScoreRowNamingAnUnapprovedFamilyIsUnknown`, `TestASingletonProposalWithoutApprovedScoreAuthorityIsRefused`, `TestAStaleProposalClosesTheWholeScope`, `TestATieAtTheTopIsRefusedRatherThanBrokenArbitrarily`, `TestATieBelowTheTopStillLeavesAUniqueWinner`, `TestAnActiveOwnerOnALaneWithNoScoreRowIsRefused`, `TestAnActiveWeeklyOwnerIsNotReplacedByAHigherScore`, `TestProposalsBoundToDifferentRouteSetsAreRefused`, `TestProposalsUnderDifferentScoreVersionsAreIncomparable`, `TestTheSameLaneTwiceIsRefused`, `TestThreeFamiliesOnOneSymbolYieldTheSingleHighestScore` |
| B6 | if at 142:3 | arm not entered (arbiter untagged suite); arm entered 4x (arbiter tagged suite); arm entered 1x (engine tagged suite); arm not entered (engine untagged suite); entered by `TestAProposalMeasuredAgainstAnotherSymbolsAuthorityIsRefused`, `TestAProposalOutsideTheExpectedScopeIsRefused` |
| B7 | if at 145:3 | arm not entered (arbiter untagged suite); arm entered 1x (arbiter tagged suite); arm not entered (engine tagged suite); arm not entered (engine untagged suite); entered by `TestProposalsBoundToDifferentRouteSetsAreRefused` |
| B8 | range at 152:2 | arm not entered (arbiter untagged suite); arm entered 40x (arbiter tagged suite); arm entered 13x (engine tagged suite); arm not entered (engine untagged suite); entered by `TestALaneMatchingTwoScoreRowsIsUnknown`, `TestALaneOutsideTheSealedEligibleSetIsRefused`, `TestALaneWithNoFamilyScoreRowIsUnknown`, `TestAProposalBuiltFromOtherEvidenceIsRefused`, `TestAProposalFromAnotherCampaignDoesNotInheritTheOwner`, `TestAProposalMutatedAfterSealingIsRefused`, `TestAProposalWhoseLineageLeavesTheScopeIsRefused`, `TestAScoreAboveTheApprovedCeilingIsRefused`, `TestAScoreRowNamingAnUnapprovedFamilyIsUnknown`, `TestASingletonProposalWithoutApprovedScoreAuthorityIsRefused`, `TestAStaleProposalClosesTheWholeScope`, `TestATieAtTheTopIsRefusedRatherThanBrokenArbitrarily`, `TestATieBelowTheTopStillLeavesAUniqueWinner`, `TestAnActiveOwnerOnALaneWithNoScoreRowIsRefused`, `TestAnActiveWeeklyOwnerIsNotReplacedByAHigherScore`, `TestProposalsUnderDifferentScoreVersionsAreIncomparable`, `TestTheSameLaneTwiceIsRefused`, `TestThreeFamiliesOnOneSymbolYieldTheSingleHighestScore` |
| B9 | if at 153:3 | arm not entered (arbiter untagged suite); arm entered 1x (arbiter tagged suite); arm not entered (engine tagged suite); arm not entered (engine untagged suite); entered by `TestAProposalMutatedAfterSealingIsRefused` |
| B10 | if at 157:3 | arm not entered (arbiter untagged suite); arm entered 4x (arbiter tagged suite); arm not entered (engine tagged suite); arm not entered (engine untagged suite); entered by `TestAProposalWhoseLineageLeavesTheScopeIsRefused` |
| B11 | if at 161:3 | arm not entered (arbiter untagged suite); arm entered 1x (arbiter tagged suite); arm not entered (engine tagged suite); arm not entered (engine untagged suite); entered by `TestAStaleProposalClosesTheWholeScope` |
| B12 | if at 165:3 | arm not entered (arbiter untagged suite); arm entered 1x (arbiter tagged suite); arm not entered (engine tagged suite); arm not entered (engine untagged suite); entered by `TestTheSameLaneTwiceIsRefused` |
| B13 | range at 176:2 | arm not entered (arbiter untagged suite); arm entered 32x (arbiter tagged suite); arm entered 13x (engine tagged suite); arm not entered (engine untagged suite); entered by `TestALaneMatchingTwoScoreRowsIsUnknown`, `TestALaneOutsideTheSealedEligibleSetIsRefused`, `TestALaneWithNoFamilyScoreRowIsUnknown`, `TestAProposalBuiltFromOtherEvidenceIsRefused`, `TestAProposalFromAnotherCampaignDoesNotInheritTheOwner`, `TestAScoreAboveTheApprovedCeilingIsRefused`, `TestAScoreRowNamingAnUnapprovedFamilyIsUnknown`, `TestASingletonProposalWithoutApprovedScoreAuthorityIsRefused`, `TestAStaleProposalClosesTheWholeScope`, `TestATieAtTheTopIsRefusedRatherThanBrokenArbitrarily`, `TestATieBelowTheTopStillLeavesAUniqueWinner`, `TestAnActiveOwnerOnALaneWithNoScoreRowIsRefused`, `TestAnActiveWeeklyOwnerIsNotReplacedByAHigherScore`, `TestProposalsUnderDifferentScoreVersionsAreIncomparable`, `TestThreeFamiliesOnOneSymbolYieldTheSingleHighestScore` |
| B14 | if at 177:3 | arm not entered (arbiter untagged suite); arm entered 3x (arbiter tagged suite); arm entered 1x (engine tagged suite); arm not entered (engine untagged suite); entered by `TestASingletonProposalWithoutApprovedScoreAuthorityIsRefused`, `TestProposalsUnderDifferentScoreVersionsAreIncomparable` |
| B15 | if at 183:2 | arm not entered (arbiter untagged suite); arm entered 4x (arbiter tagged suite); arm not entered (engine tagged suite); arm not entered (engine untagged suite); entered by `TestAProposalFromAnotherCampaignDoesNotInheritTheOwner`, `TestAnActiveOwnerOnALaneWithNoScoreRowIsRefused`, `TestAnActiveWeeklyOwnerIsNotReplacedByAHigherScore` |

A row states what was measured, not what is intended. An arm recorded as not entered is a coverage gap, not a pass.
