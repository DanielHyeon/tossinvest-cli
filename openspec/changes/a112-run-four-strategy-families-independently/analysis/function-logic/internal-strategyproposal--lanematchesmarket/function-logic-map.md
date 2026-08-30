# Function Logic Map: `laneMatchesMarket`

- Source: `internal/strategyproposal/production.go` (657-665)
- Function: `laneMatchesMarket` in package `strategyproposal`
- Signature: `laneMatchesMarket(params=4, results=1)`
- File SHA-256: `b6e54b502e5092745426f8f4a37e4a02777d525a2099aa90de9f7379ee4a2c18`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file.
- AST evidence: `ast.json` — AST branches 2.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

Maps a lane id to its market, including both breakout lanes.

The signature above is the exhaustive input/result record; this map does not infer state the AST does not show.

## Branches and early returns

- Measurement regime: Go coverage profiles, count mode.
- untagged proposal suite: `go test -count=1 -covermode=count -coverpkg=./internal/strategyproposal ./internal/strategyproposal/`
- tagged proposal suite: `go test -count=1 -tags tossos_testseams -covermode=count -coverpkg=./internal/strategyproposal,./internal/strategyflow,./internal/strategyrouter,./internal/app/engine ./internal/strategyproposal/`
- tagged engine suite: `go test -count=1 -tags tossos_testseams -covermode=count -coverpkg=./internal/strategyproposal,./internal/strategyflow,./internal/strategyrouter,./internal/app/engine ./internal/app/engine/`
- untagged engine suite: `go test -count=1 -covermode=count -coverpkg=./internal/strategyproposal,./internal/strategyflow,./internal/strategyrouter,./internal/app/engine ./internal/app/engine/`
- Measured entry: the function body was not executed (untagged proposal suite); executed 24x (tagged proposal suite); not executed (tagged engine suite); not executed (untagged engine suite).

Exact AST return positions: 659:3, 662:3, 664:2.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 658:2 | arm not entered (untagged proposal suite); arm entered 21x (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); entered by `TestABreakoutLaneWithNoEvidenceYetIsAbsenceNotFault`, `TestAHealthyBatchCarriesNoFault`, `TestAProposalLostAfterAdmissionIsRecordedAsATypedFault`, `TestAScopeTheCurrentCandidateDoesNotMatchIsAbsenceNotFault`, `TestAnEvaluationRefusalIsAbsenceNotFault`, `TestAnUnusableRouteAuthorityIsAFault`, `TestLaneMatchesMarketCoversTheBreakoutFamily`, `TestProductionProposalAuthorityLoadsPairedSignedKRUSSnapshots`, `TestProductionProposalAuthorityRecognizesExactPairedSixLaneMatrix`, `TestValidScopesAcceptsSeveralFamiliesForOneSymbolAndStillRejectsDuplicateLanes` |
| B2 | if | 661:2 | arm not entered (untagged proposal suite); arm entered 9x (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); entered by `TestLaneMatchesMarketCoversTheBreakoutFamily`, `TestProductionProposalAuthorityFailureIsMarketLocal`, `TestProductionProposalAuthorityLoadsPairedSignedKRUSSnapshots`, `TestProductionProposalAuthorityRecognizesExactPairedSixLaneMatrix` |

## Calls and live bindings

| Callee expression | Position |
|---|---|

## State mutations and fallbacks

- AST assignments: 0. Defers: 0. Goroutine statements: 0.

## Safety conclusion

Unchanged by this lot; the bundle is refreshed because the file hash moved.
