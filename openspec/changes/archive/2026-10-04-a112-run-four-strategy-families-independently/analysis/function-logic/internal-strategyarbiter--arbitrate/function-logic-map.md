# Function Logic Map: `Arbitrate`

- Source: `internal/strategyarbiter/arbiter.go` (117-189)
- Function: `Arbitrate` in package `strategyarbiter`
- Signature: `Arbitrate(params=1, results=1)`
- File SHA-256: `ba484e68bd49e73081afaf031129c1a418af4103a436b94382b4105b2f68da2a`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file.
- AST evidence: `ast.json` — AST branches 15.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

한 소유자 범위 `(AccountRef, Market, Symbol, PositionGeneration)` 의 봉인된 제안 목록을 받아 최대 하나를 고른다.
기대 범위는 호출자가 넘기고, 각 제안이 들고 온 `ProductionRouteAuthority` 와 `strategyflow.Result` 는 그 기대와 대조된다.
채점 기준(`Calibration`)과 가족 점수표(`FamilyScores`)는 호출자의 주장이 아니라 봉인된 권한 객체에서 읽으며,
`SealsValid()` 가 재료와 봉인의 일치를 다시 계산한다. 거절 코드는 동결 골든 `refusal_enums.arbitration` 의 여섯 개뿐이고,
그 안에서 무엇이 발화했는지는 계약이 아닌 `Outcome.Detail` 이 들고 간다.
자격 집합은 `strategyrouter.RouteSet` 을 이 함수가 직접 유도한다 —
호출자가 넘긴 자격 집합을 믿으면 자격을 스스로 만들어 오는 호출자를 막을 수 없다.
거절은 언제나 소유자 범위 전체를 닫는다. 문제 있는 제안 하나만 빼고 나머지를 견주지 않는다.

The signature above is the exhaustive input/result record; this map does not infer state the AST does not show.

## Branches and early returns

- Measurement regime: Go coverage profiles, count mode.
- arbiter untagged suite: `go test -count=1 -covermode=count -coverpkg=./internal/strategyarbiter ./internal/strategyarbiter/`
- arbiter tagged suite: `go test -count=1 -tags tossos_testseams -covermode=count -coverpkg=./internal/strategyarbiter ./internal/strategyarbiter/`
- engine tagged suite: `go test -count=1 -tags tossos_testseams -covermode=count -coverpkg=./internal/strategyarbiter,./internal/app/engine ./internal/app/engine/`
- engine untagged suite: `go test -count=1 -covermode=count -coverpkg=./internal/strategyarbiter,./internal/app/engine ./internal/app/engine/`
- Per-test attribution set: every `Test*` function in `internal/strategyarbiter`, each run alone under the seam tag. This is the whole package, not a sample.
- Measured entry: the function body was executed in the arbiter untagged and tagged suites and in the engine tagged suite; the engine untagged suite never reaches it because the engine fixtures that build proposals live behind the seam tag.

Exact AST return positions: 122:3, 125:3, 137:4, 139:3, 143:4, 146:4, 154:4, 159:4, 162:4, 166:4, 178:4, 184:3, 188:2.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 120:2 | arm entered 5x (arbiter untagged suite); arm entered 5x (arbiter tagged suite); arm not entered (engine tagged suite); arm not entered (engine untagged suite); entered by `TestARequestWithoutAnIdentityIsRefused` |
| B2 | if | 124:2 | arm entered 1x (arbiter untagged suite); arm entered 1x (arbiter tagged suite); arm not entered (engine tagged suite); arm not entered (engine untagged suite); entered by `TestNoProposalIsNotASelection` |
| B3 | if | 135:2 | arm entered 1x (arbiter untagged suite); arm entered 4x (arbiter tagged suite); arm not entered (engine tagged suite); arm not entered (engine untagged suite); entered by `TestAStaleOwnerRevisionIsItsOwnRefusal`, `TestAZeroProposalIsNeverSelected`, `TestAnOwnerSnapshotOutsideItsFreshnessWindowIsAStaleOwner`, `TestTwoActiveOwnersAreItsOwnRefusal` |
| B4 | if | 136:3 | arm not entered (arbiter untagged suite); arm entered 3x (arbiter tagged suite); arm not entered (engine tagged suite); arm not entered (engine untagged suite); entered by `TestAStaleOwnerRevisionIsItsOwnRefusal`, `TestAnOwnerSnapshotOutsideItsFreshnessWindowIsAStaleOwner`, `TestTwoActiveOwnersAreItsOwnRefusal` |
| B5 | range | 141:2 | arm not entered (arbiter untagged suite); arm entered 47x (arbiter tagged suite); arm entered 14x (engine tagged suite); arm not entered (engine untagged suite); entered by `TestALaneMatchingTwoScoreRowsIsUnknown`, `TestALaneOutsideTheSealedEligibleSetIsRefused`, `TestALaneWithNoFamilyScoreRowIsUnknown`, `TestAProposalBuiltFromOtherEvidenceIsRefused`, `TestAProposalFromAnotherCampaignDoesNotInheritTheOwner`, `TestAProposalMeasuredAgainstAnotherSymbolsAuthorityIsRefused`, `TestAProposalMutatedAfterSealingIsRefused`, `TestAProposalOutsideTheExpectedScopeIsRefused`, `TestAProposalWhoseLineageLeavesTheScopeIsRefused`, `TestAScoreAboveTheApprovedCeilingIsRefused`, `TestAScoreRowNamingAnUnapprovedFamilyIsUnknown`, `TestASingletonProposalWithoutApprovedScoreAuthorityIsRefused`, `TestAStaleProposalClosesTheWholeScope`, `TestATieAtTheTopIsRefusedRatherThanBrokenArbitrarily`, `TestATieBelowTheTopStillLeavesAUniqueWinner`, `TestAnActiveOwnerOnALaneWithNoScoreRowIsRefused`, `TestAnActiveWeeklyOwnerIsNotReplacedByAHigherScore`, `TestProposalsBoundToDifferentRouteSetsAreRefused`, `TestProposalsUnderDifferentScoreVersionsAreIncomparable`, `TestTheSameLaneTwiceIsRefused`, `TestThreeFamiliesOnOneSymbolYieldTheSingleHighestScore` |
| B6 | if | 142:3 | arm not entered (arbiter untagged suite); arm entered 4x (arbiter tagged suite); arm entered 1x (engine tagged suite); arm not entered (engine untagged suite); entered by `TestAProposalMeasuredAgainstAnotherSymbolsAuthorityIsRefused`, `TestAProposalOutsideTheExpectedScopeIsRefused` |
| B7 | if | 145:3 | arm not entered (arbiter untagged suite); arm entered 1x (arbiter tagged suite); arm not entered (engine tagged suite); arm not entered (engine untagged suite); entered by `TestProposalsBoundToDifferentRouteSetsAreRefused` |
| B8 | range | 152:2 | arm not entered (arbiter untagged suite); arm entered 40x (arbiter tagged suite); arm entered 13x (engine tagged suite); arm not entered (engine untagged suite); entered by `TestALaneMatchingTwoScoreRowsIsUnknown`, `TestALaneOutsideTheSealedEligibleSetIsRefused`, `TestALaneWithNoFamilyScoreRowIsUnknown`, `TestAProposalBuiltFromOtherEvidenceIsRefused`, `TestAProposalFromAnotherCampaignDoesNotInheritTheOwner`, `TestAProposalMutatedAfterSealingIsRefused`, `TestAProposalWhoseLineageLeavesTheScopeIsRefused`, `TestAScoreAboveTheApprovedCeilingIsRefused`, `TestAScoreRowNamingAnUnapprovedFamilyIsUnknown`, `TestASingletonProposalWithoutApprovedScoreAuthorityIsRefused`, `TestAStaleProposalClosesTheWholeScope`, `TestATieAtTheTopIsRefusedRatherThanBrokenArbitrarily`, `TestATieBelowTheTopStillLeavesAUniqueWinner`, `TestAnActiveOwnerOnALaneWithNoScoreRowIsRefused`, `TestAnActiveWeeklyOwnerIsNotReplacedByAHigherScore`, `TestProposalsUnderDifferentScoreVersionsAreIncomparable`, `TestTheSameLaneTwiceIsRefused`, `TestThreeFamiliesOnOneSymbolYieldTheSingleHighestScore` |
| B9 | if | 153:3 | arm not entered (arbiter untagged suite); arm entered 1x (arbiter tagged suite); arm not entered (engine tagged suite); arm not entered (engine untagged suite); entered by `TestAProposalMutatedAfterSealingIsRefused` |
| B10 | if | 157:3 | arm not entered (arbiter untagged suite); arm entered 4x (arbiter tagged suite); arm not entered (engine tagged suite); arm not entered (engine untagged suite); entered by `TestAProposalWhoseLineageLeavesTheScopeIsRefused` |
| B11 | if | 161:3 | arm not entered (arbiter untagged suite); arm entered 1x (arbiter tagged suite); arm not entered (engine tagged suite); arm not entered (engine untagged suite); entered by `TestAStaleProposalClosesTheWholeScope` |
| B12 | if | 165:3 | arm not entered (arbiter untagged suite); arm entered 1x (arbiter tagged suite); arm not entered (engine tagged suite); arm not entered (engine untagged suite); entered by `TestTheSameLaneTwiceIsRefused` |
| B13 | range | 176:2 | arm not entered (arbiter untagged suite); arm entered 32x (arbiter tagged suite); arm entered 13x (engine tagged suite); arm not entered (engine untagged suite); entered by `TestALaneMatchingTwoScoreRowsIsUnknown`, `TestALaneOutsideTheSealedEligibleSetIsRefused`, `TestALaneWithNoFamilyScoreRowIsUnknown`, `TestAProposalBuiltFromOtherEvidenceIsRefused`, `TestAProposalFromAnotherCampaignDoesNotInheritTheOwner`, `TestAScoreAboveTheApprovedCeilingIsRefused`, `TestAScoreRowNamingAnUnapprovedFamilyIsUnknown`, `TestASingletonProposalWithoutApprovedScoreAuthorityIsRefused`, `TestAStaleProposalClosesTheWholeScope`, `TestATieAtTheTopIsRefusedRatherThanBrokenArbitrarily`, `TestATieBelowTheTopStillLeavesAUniqueWinner`, `TestAnActiveOwnerOnALaneWithNoScoreRowIsRefused`, `TestAnActiveWeeklyOwnerIsNotReplacedByAHigherScore`, `TestProposalsUnderDifferentScoreVersionsAreIncomparable`, `TestThreeFamiliesOnOneSymbolYieldTheSingleHighestScore` |
| B14 | if | 177:3 | arm not entered (arbiter untagged suite); arm entered 3x (arbiter tagged suite); arm entered 1x (engine tagged suite); arm not entered (engine untagged suite); entered by `TestASingletonProposalWithoutApprovedScoreAuthorityIsRefused`, `TestProposalsUnderDifferentScoreVersionsAreIncomparable` |
| B15 | if | 183:2 | arm not entered (arbiter untagged suite); arm entered 4x (arbiter tagged suite); arm not entered (engine tagged suite); arm not entered (engine untagged suite); entered by `TestAProposalFromAnotherCampaignDoesNotInheritTheOwner`, `TestAnActiveOwnerOnALaneWithNoScoreRowIsRefused`, `TestAnActiveWeeklyOwnerIsNotReplacedByAHigherScore` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `request.ObservedAt.IsZero` | 121:97 |
| `refuse` | 122:10 |
| `len` | 124:5 |
| `refuse` | 125:10 |
| `request.Proposals.Authority.Request` | 133:13 |
| `strategyrouter.RouteSet` | 134:12 |
| `routed.Valid` | 135:51 |
| `len` | 135:69 |
| `ownerSnapshotFault` | 136:22 |
| `refuse` | 137:11 |
| `refuse` | 139:10 |
| `proposal.Authority.Request` | 142:6 |
| `refuse` | 143:11 |
| `SetDigest` | 145:6 |
| `strategyrouter.RouteSet` | 145:6 |
| `proposal.Authority.Request` | 145:30 |
| `routed.SetDigest` | 145:75 |
| `refuse` | 146:11 |
| `make` | 151:10 |
| `len` | 151:33 |
| `proposal.Result.ValidProposal` | 153:7 |
| `refuse` | 154:11 |
| `refuse` | 159:11 |
| `request.ObservedAt.UnixNano` | 161:39 |
| `refuse` | 162:11 |
| `refuse` | 166:11 |
| `request.Proposals.Authority.Calibration` | 175:17 |
| `proposal.Authority.SealsValid` | 177:7 |
| `proposal.Authority.Calibration` | 177:42 |
| `refuse` | 178:11 |
| `continueExistingOwner` | 184:10 |
| `selectHighestScore` | 188:9 |

## State mutations and fallbacks

- AST assignments: 9. Defers: 0. Goroutine statements: 0.

## Safety conclusion

이 함수는 아무것도 바꾸지 않는다. 패키지의 직접 import 는 `time`, `strategyflow`, `strategyrouter` 뿐이며
`TestTheArbiterCannotReachAnyMutationCapability` 가 그 목록을 지킨다(저널·주문 게이트웨이를 들여오면 실패한다).
선택은 dispatch 권한이 아니라 후보다 — 최종 권한은 여전히 저널의 원자적 owner/q_final 승인과 dispatch lease 다.
