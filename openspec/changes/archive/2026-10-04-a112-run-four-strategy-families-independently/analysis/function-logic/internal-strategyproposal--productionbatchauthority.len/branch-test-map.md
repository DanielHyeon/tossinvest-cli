# Branch Test Map: `ProductionBatchAuthority.Len`

- Source: `internal/strategyproposal/production.go` (150-150); file SHA-256 `b6e54b502e5092745426f8f4a37e4a02777d525a2099aa90de9f7379ee4a2c18`. AST branch positions are authoritative.
- Rows carry measured counts from Go coverage profiles, count mode.
- tagged proposal suite: `go test -c -tags tossos_testseams -covermode=count -coverpkg=./internal/strategyproposal,./internal/strategyflow,./internal/strategyrouter,./internal/app/engine ./internal/strategyproposal/` 뒤 그 바이너리를 `-test.coverprofile` 로 실행.
- untagged proposal suite: 같은 명령에서 태그만 뺀 것(`-coverpkg=./internal/strategyproposal`).
- tagged engine suite: 같은 `-coverpkg` 로 `./internal/app/engine/` 를 빌드해 실행.
- Per-test attribution set: 태그 proposal 바이너리의 스무 개 `Test*` 를 하나씩 `-test.run` 으로 돌린 프로파일 전부. 표본이 아니라 그 패키지의 전수다.

이 함수는 분기가 없다. 그래서 행은 하나이고, 그 행이 재는 것은 **본문에 들어왔는가** 다.

| Branch | Anchor | Measured disposition |
|---|---|---|
| B1 | branchless happy path at 150:1 | arm entered 6x (tagged proposal suite); arm not entered (untagged proposal suite); arm entered 15x (tagged engine suite); entered by `TestABreakoutLaneWithNoEvidenceYetIsAbsenceNotFault`, `TestAHealthyBatchCarriesNoFault`, `TestAProposalLostAfterAdmissionIsRecordedAsATypedFault`, `TestAScopeTheCurrentCandidateDoesNotMatchIsAbsenceNotFault`, `TestAnEvaluationRefusalIsAbsenceNotFault`, `TestProductionProposalAuthorityFailureIsMarketLocal` |

A row states what was measured, not what is intended. An arm recorded as not entered is a coverage gap, not a pass.
