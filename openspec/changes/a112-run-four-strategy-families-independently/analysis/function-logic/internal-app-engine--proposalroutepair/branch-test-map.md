# Branch Test Map: `proposalRoutePair`

- Source: `internal/app/engine/strategy_proposal_authority_test.go`; file SHA-256
  `4acb8506cc32d2cc5fd4eda1a5366152ba7dcf92e704d078ceced5fe268513ea`. AST branch positions are authoritative.
- Go 커버리지는 `_test.go` 를 계측하지 않으므로 이 함수의 행은 커버리지 카운트가 아니라
  그 분기를 실행한 것으로 관측된 run 을 적는다.
- 관측 run: `go test -count=1 -tags tossos_testseams -run '^TestStrategyProposalAuthority' ./internal/app/engine/` (통과).

| # | mutation | result | killed by |
|---|---|---|---|
| M-F1 | 픽스처에서 `WithArbitrationScoresForTest` 한 줄을 되돌린다(보정 없는 권한) | KILLED | `TestStrategyProposalAuthorityLoadsKRUSConcurrently`, `TestStrategyProposalAuthorityKeepsMarketFailureLocal` |

| M-F2 | 픽스처의 경로 다이제스트를 계보에서 읽지 않고 임의 문자열로 되돌린다 | KILLED | `TestStrategyProposalAuthorityLoadsKRUSConcurrently`, `TestThreeFamiliesOnOneSymbolNowSelectTheHighestScoreInsteadOfClosingTheMarket` |

M-F1 과 M-F2 는 이 lot 을 만들면서 실제로 관측됐다. 배선 직후 픽스처를 고치기 전 상태에서 두 테스트가
`Reason:ARBITRATION_REFUSED ArbitrationRefusal:ARBITRATION_UNCALIBRATED` 로 실패했다.

| Branch | Anchor | Measured disposition |
|---|---|---|
| B1 | if at 89:3 | 두 테스트가 KR/US 양쪽을 만들므로 매 실행 진입한다. |
| B2 | if at 93:3 | 관측된 진입 없음 — `NewOwnerKey` 는 픽스처 입력에서 실패하지 않는다. |
| B3 | if at 100:3 | 관측된 진입 없음 — `ProductionRouteAuthorityForTest` 는 픽스처 입력에서 실패하지 않는다. |

A row states what was measured, not what is intended. An arm recorded as not entered is a coverage gap, not a pass.
