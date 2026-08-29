# Function Logic Map: `familyScore`

- Source: `internal/strategyarbiter/arbiter.go` (283-300)
- Function: `familyScore` in package `strategyarbiter`
- Signature: `familyScore(params=1, results=3)`
- File SHA-256: `1788b0503479c4c4e8b4b17d7e2e6e2fd189f414ebde7877a7d5043157be6d03`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file.
- AST evidence: `ast.json` — AST branches 4.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

제안의 레인 삼항을 봉인된 가족 점수 행 *정확히 하나* 에 붙인다. 한 행에도 안 붙거나 두 행에 걸치면
그 제안은 견줄 수 없는 것이다. 가족 이름은 승인된 네 개 열거 안에 있어야 하고,
점수는 `strategyrouter.ScorePPMMax` 를 넘지 않아야 한다. 그 상한은 매니페스트 검증이 쓰는 상수에서 유도한 값이며 사본이 아니다.

The signature above is the exhaustive input/result record; this map does not infer state the AST does not show.

## Branches and early returns

- Measurement regime: Go coverage profiles, count mode.
- arbiter untagged suite: `go test -count=1 -covermode=count -coverpkg=./internal/strategyarbiter ./internal/strategyarbiter/`
- arbiter tagged suite: `go test -count=1 -tags tossos_testseams -covermode=count -coverpkg=./internal/strategyarbiter ./internal/strategyarbiter/`
- engine tagged suite: `go test -count=1 -tags tossos_testseams -covermode=count -coverpkg=./internal/strategyarbiter,./internal/app/engine ./internal/app/engine/`
- engine untagged suite: `go test -count=1 -covermode=count -coverpkg=./internal/strategyarbiter,./internal/app/engine ./internal/app/engine/`
- Per-test attribution set: every `Test*` function in `internal/strategyarbiter`, each run alone under the seam tag. This is the whole package, not a sample.
- Measured entry: executed from both `continueExistingOwner` and `selectHighestScore` in the arbiter tagged suite and the engine tagged suite.

Exact AST return positions: 294:3, 297:3, 299:2.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | range | 287:2 | arm not entered (arbiter untagged suite); arm entered 39x (arbiter tagged suite); arm entered 46x (engine tagged suite); arm not entered (engine untagged suite); entered by `TestALaneMatchingTwoScoreRowsIsUnknown`, `TestALaneOutsideTheSealedEligibleSetIsRefused`, `TestALaneWithNoFamilyScoreRowIsUnknown`, `TestAProposalBuiltFromOtherEvidenceIsRefused`, `TestAScoreAboveTheApprovedCeilingIsRefused`, `TestAScoreRowNamingAnUnapprovedFamilyIsUnknown`, `TestASingletonProposalWithoutApprovedScoreAuthorityIsRefused`, `TestAStaleProposalClosesTheWholeScope`, `TestATieAtTheTopIsRefusedRatherThanBrokenArbitrarily`, `TestATieBelowTheTopStillLeavesAUniqueWinner`, `TestAnActiveOwnerOnALaneWithNoScoreRowIsRefused`, `TestAnActiveWeeklyOwnerIsNotReplacedByAHigherScore`, `TestThreeFamiliesOnOneSymbolYieldTheSingleHighestScore` |
| B2 | if | 288:3 | arm not entered (arbiter untagged suite); arm entered 19x (arbiter tagged suite); arm entered 12x (engine tagged suite); arm not entered (engine untagged suite); entered by `TestALaneMatchingTwoScoreRowsIsUnknown`, `TestALaneOutsideTheSealedEligibleSetIsRefused`, `TestAProposalBuiltFromOtherEvidenceIsRefused`, `TestAScoreAboveTheApprovedCeilingIsRefused`, `TestAScoreRowNamingAnUnapprovedFamilyIsUnknown`, `TestASingletonProposalWithoutApprovedScoreAuthorityIsRefused`, `TestAStaleProposalClosesTheWholeScope`, `TestATieAtTheTopIsRefusedRatherThanBrokenArbitrarily`, `TestATieBelowTheTopStillLeavesAUniqueWinner`, `TestAnActiveWeeklyOwnerIsNotReplacedByAHigherScore`, `TestThreeFamiliesOnOneSymbolYieldTheSingleHighestScore` |
| B3 | if | 293:2 | arm not entered (arbiter untagged suite); arm entered 4x (arbiter tagged suite); arm not entered (engine tagged suite); arm not entered (engine untagged suite); entered by `TestALaneMatchingTwoScoreRowsIsUnknown`, `TestALaneWithNoFamilyScoreRowIsUnknown`, `TestAScoreRowNamingAnUnapprovedFamilyIsUnknown`, `TestAnActiveOwnerOnALaneWithNoScoreRowIsRefused` |
| B4 | if | 296:2 | arm not entered (arbiter untagged suite); arm entered 1x (arbiter tagged suite); arm not entered (engine tagged suite); arm not entered (engine untagged suite); entered by `TestAScoreAboveTheApprovedCeilingIsRefused` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `proposal.Authority.FamilyScores` | 287:24 |
| `found.Family.Known` | 293:22 |

## State mutations and fallbacks

- AST assignments: 4. Defers: 0. Goroutine statements: 0.

## Safety conclusion

값을 만들지 않고 표에서 찾기만 한다. 못 찾거나 애매하면 닫는다.
