# Branch Test Map: `laneMatchesMarket`

- Source: `internal/strategyproposal/production.go` (657-665); file SHA-256 `b6e54b502e5092745426f8f4a37e4a02777d525a2099aa90de9f7379ee4a2c18`. AST branch positions are authoritative.
- Rows carry measured counts from Go coverage profiles, count mode.
- untagged proposal suite: `go test -count=1 -covermode=count -coverpkg=./internal/strategyproposal ./internal/strategyproposal/`
- tagged proposal suite: `go test -count=1 -tags tossos_testseams -covermode=count -coverpkg=./internal/strategyproposal,./internal/strategyflow,./internal/strategyrouter,./internal/app/engine ./internal/strategyproposal/`
- tagged engine suite: `go test -count=1 -tags tossos_testseams -covermode=count -coverpkg=./internal/strategyproposal,./internal/strategyflow,./internal/strategyrouter,./internal/app/engine ./internal/app/engine/`
- untagged engine suite: `go test -count=1 -covermode=count -coverpkg=./internal/strategyproposal,./internal/strategyflow,./internal/strategyrouter,./internal/app/engine ./internal/app/engine/`

| Branch | Anchor | Measured disposition |
|---|---|---|
| B1 | if at 658:2 | arm not entered (untagged proposal suite); arm entered 21x (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); entered by `TestABreakoutLaneWithNoEvidenceYetIsAbsenceNotFault`, `TestAHealthyBatchCarriesNoFault`, `TestAProposalLostAfterAdmissionIsRecordedAsATypedFault`, `TestAScopeTheCurrentCandidateDoesNotMatchIsAbsenceNotFault`, `TestAnEvaluationRefusalIsAbsenceNotFault`, `TestAnUnusableRouteAuthorityIsAFault`, `TestLaneMatchesMarketCoversTheBreakoutFamily`, `TestProductionProposalAuthorityLoadsPairedSignedKRUSSnapshots`, `TestProductionProposalAuthorityRecognizesExactPairedSixLaneMatrix`, `TestValidScopesAcceptsSeveralFamiliesForOneSymbolAndStillRejectsDuplicateLanes` |
| B2 | if at 661:2 | arm not entered (untagged proposal suite); arm entered 9x (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); entered by `TestLaneMatchesMarketCoversTheBreakoutFamily`, `TestProductionProposalAuthorityFailureIsMarketLocal`, `TestProductionProposalAuthorityLoadsPairedSignedKRUSSnapshots`, `TestProductionProposalAuthorityRecognizesExactPairedSixLaneMatrix` |

A row states what was measured, not what is intended. An arm recorded as not entered is a coverage gap, not a pass.
