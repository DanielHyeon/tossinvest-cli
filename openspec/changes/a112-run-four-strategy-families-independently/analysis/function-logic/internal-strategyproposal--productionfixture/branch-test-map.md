# Branch Test Map: `productionFixture`

- Source: `internal/strategyproposal/production_test.go` (103-106); file SHA-256 `abfb3e4b1b06d32000fa3b8c4d8ee1361d71b16dd864172e850361a9c63e8969`. AST branch positions are authoritative.
- 이 함수는 `_test.go` 안에 있다. Go 커버리지는 테스트 파일 자체를 계측하지 않으므로
  이 함수에는 커버리지 블록이 없다(태그 proposal 프로파일에서 `production_test.go` 블록 0 개로 확인).
  그래서 아래 행의 근거는 커버리지 수가 아니라 **이 함수를 부르는 테스트의 목록**이다.

분기가 없다. 행은 하나이고, 그 행이 재는 것은 이 도우미를 실제로 부르는 시험이 있는가다.

| Branch | Anchor | Measured disposition |
|---|---|---|
| B1 | branchless happy path at 103:1 | 커버리지 계측 대상 아님(테스트 파일); 부르는 시험은 `TestProductionProposalAuthorityLoadsPairedSignedKRUSSnapshots`, `TestProductionProposalAuthorityFailureIsMarketLocal`, `TestAHealthyBatchCarriesNoFault`, `TestAnUnusableRouteAuthorityIsAFault` — 네 시험 전부 이 lot 의 전체 스위트에서 PASS |

A row states what was measured, not what is intended. An arm recorded as not entered is a coverage gap, not a pass.
