# Function Logic Map: `selectHighestScore`

- Source: `internal/strategyarbiter/arbiter.go` (240-279)
- Function: `selectHighestScore` in package `strategyarbiter`
- Signature: `selectHighestScore(params=2, results=1)`
- File SHA-256: `ba484e68bd49e73081afaf031129c1a418af4103a436b94382b4105b2f68da2a`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file.
- AST evidence: `ast.json` — AST branches 9.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

활성 소유자가 없을 때만 불린다. 봉인된 자격 집합의 결정들을 레인 삼항 `(Horizon, LaneID, LaneVersion)` 집합으로 바꾸고,
제안마다 그 집합에 드는지 확인한 뒤, **그 결정의 증거·설정 다이제스트가 제안 계보의 두 값과 같은지**를 확인하고,
가족 점수 한 행에 붙여 최고점 하나를 고른다. 레인 이름만 맞추면 지금이 아닌 어떤 증거로 만들어진 옛 제안이
현재 권한을 타고 들어온다 — `Propose` 는 그때의 결정에서 두 값을 계보에 박는다(`flow.go:67-68`).
최고점이 둘 이상이면 임의로 집지 않고 닫는다. 목록이 비어 있는 경우는 `Arbitrate` 가 이미 거절해 여기 오지 않는다.

The signature above is the exhaustive input/result record; this map does not infer state the AST does not show.

## Branches and early returns

- Measurement regime: Go coverage profiles, count mode.
- arbiter untagged suite: `go test -count=1 -covermode=count -coverpkg=./internal/strategyarbiter ./internal/strategyarbiter/`
- arbiter tagged suite: `go test -count=1 -tags tossos_testseams -covermode=count -coverpkg=./internal/strategyarbiter ./internal/strategyarbiter/`
- engine tagged suite: `go test -count=1 -tags tossos_testseams -covermode=count -coverpkg=./internal/strategyarbiter,./internal/app/engine ./internal/app/engine/`
- engine untagged suite: `go test -count=1 -covermode=count -coverpkg=./internal/strategyarbiter,./internal/app/engine ./internal/app/engine/`
- Per-test attribution set: every `Test*` function in `internal/strategyarbiter`, each run alone under the seam tag. This is the whole package, not a sample.
- Measured entry: executed only from `Arbitrate` in the arbiter tagged suite and the engine tagged suite; never in either untagged suite.

Exact AST return positions: 251:4, 257:4, 261:4, 275:3, 277:2.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | range | 242:2 | arm not entered (arbiter untagged suite); arm entered 21x (arbiter tagged suite); arm entered 12x (engine tagged suite); arm not entered (engine untagged suite); entered by `TestALaneMatchingTwoScoreRowsIsUnknown`, `TestALaneOutsideTheSealedEligibleSetIsRefused`, `TestALaneWithNoFamilyScoreRowIsUnknown`, `TestAProposalBuiltFromOtherEvidenceIsRefused`, `TestAScoreAboveTheApprovedCeilingIsRefused`, `TestAScoreRowNamingAnUnapprovedFamilyIsUnknown`, `TestASingletonProposalWithoutApprovedScoreAuthorityIsRefused`, `TestAStaleProposalClosesTheWholeScope`, `TestATieAtTheTopIsRefusedRatherThanBrokenArbitrarily`, `TestATieBelowTheTopStillLeavesAUniqueWinner`, `TestThreeFamiliesOnOneSymbolYieldTheSingleHighestScore` |
| B2 | range | 247:2 | arm not entered (arbiter untagged suite); arm entered 22x (arbiter tagged suite); arm entered 12x (engine tagged suite); arm not entered (engine untagged suite); entered by `TestALaneMatchingTwoScoreRowsIsUnknown`, `TestALaneOutsideTheSealedEligibleSetIsRefused`, `TestALaneWithNoFamilyScoreRowIsUnknown`, `TestAProposalBuiltFromOtherEvidenceIsRefused`, `TestAScoreAboveTheApprovedCeilingIsRefused`, `TestAScoreRowNamingAnUnapprovedFamilyIsUnknown`, `TestASingletonProposalWithoutApprovedScoreAuthorityIsRefused`, `TestAStaleProposalClosesTheWholeScope`, `TestATieAtTheTopIsRefusedRatherThanBrokenArbitrarily`, `TestATieBelowTheTopStillLeavesAUniqueWinner`, `TestThreeFamiliesOnOneSymbolYieldTheSingleHighestScore` |
| B3 | if | 250:3 | arm not entered (arbiter untagged suite); arm entered 1x (arbiter tagged suite); arm not entered (engine tagged suite); arm not entered (engine untagged suite); entered by `TestALaneOutsideTheSealedEligibleSetIsRefused` |
| B4 | if | 256:3 | arm not entered (arbiter untagged suite); arm entered 3x (arbiter tagged suite); arm not entered (engine tagged suite); arm not entered (engine untagged suite); entered by `TestAProposalBuiltFromOtherEvidenceIsRefused` |
| B5 | if | 260:3 | arm not entered (arbiter untagged suite); arm entered 4x (arbiter tagged suite); arm not entered (engine tagged suite); arm not entered (engine untagged suite); entered by `TestALaneMatchingTwoScoreRowsIsUnknown`, `TestALaneWithNoFamilyScoreRowIsUnknown`, `TestAScoreAboveTheApprovedCeilingIsRefused`, `TestAScoreRowNamingAnUnapprovedFamilyIsUnknown` |
| B6 | switch | 263:3 | arm not entered (arbiter untagged suite); arm entered 10x (arbiter tagged suite); arm entered 10x (engine tagged suite); arm not entered (engine untagged suite); entered by `TestALaneOutsideTheSealedEligibleSetIsRefused`, `TestAProposalBuiltFromOtherEvidenceIsRefused`, `TestAScoreAboveTheApprovedCeilingIsRefused`, `TestASingletonProposalWithoutApprovedScoreAuthorityIsRefused`, `TestAStaleProposalClosesTheWholeScope`, `TestATieAtTheTopIsRefusedRatherThanBrokenArbitrarily`, `TestATieBelowTheTopStillLeavesAUniqueWinner`, `TestThreeFamiliesOnOneSymbolYieldTheSingleHighestScore` |
| B7 | case | 264:3 | arm not entered (arbiter untagged suite); arm entered 10x (arbiter tagged suite); arm entered 10x (engine tagged suite); arm not entered (engine untagged suite); entered by `TestALaneOutsideTheSealedEligibleSetIsRefused`, `TestAProposalBuiltFromOtherEvidenceIsRefused`, `TestAScoreAboveTheApprovedCeilingIsRefused`, `TestASingletonProposalWithoutApprovedScoreAuthorityIsRefused`, `TestAStaleProposalClosesTheWholeScope`, `TestATieAtTheTopIsRefusedRatherThanBrokenArbitrarily`, `TestATieBelowTheTopStillLeavesAUniqueWinner`, `TestThreeFamiliesOnOneSymbolYieldTheSingleHighestScore` |
| B8 | case | 266:3 | arm not entered (arbiter untagged suite); arm entered 1x (arbiter tagged suite); arm entered 1x (engine tagged suite); arm not entered (engine untagged suite); entered by `TestATieAtTheTopIsRefusedRatherThanBrokenArbitrarily` |
| B9 | if | 274:2 | arm not entered (arbiter untagged suite); arm entered 1x (arbiter tagged suite); arm entered 1x (engine tagged suite); arm not entered (engine untagged suite); entered by `TestATieAtTheTopIsRefusedRatherThanBrokenArbitrarily` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `make` | 241:14 |
| `len` | 241:61 |
| `refuse` | 251:11 |
| `refuse` | 257:11 |
| `familyScore` | 259:26 |
| `refuse` | 261:11 |
| `refuse` | 275:10 |

## State mutations and fallbacks

- AST assignments: 8. Defers: 0. Goroutine statements: 0.

## Safety conclusion

선택은 최대 하나다. 동률·부적격·모르는 가족은 모두 닫는 방향이며, 어느 경로도 제안을 새로 만들지 않는다.
