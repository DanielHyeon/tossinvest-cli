# Branch Test Map: `Console.livePositions`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 원장 미배선·부재·스키마 불일치 | `TestThePositionsScreenRendersWithEitherSourceMissing`, `TestThePositionsScreenNamesBothSchemaDirections` | — | yes |
| B2 | `AccountRefs` 실패 — 사유를 적고 계속 | `TestAHoldingIsNotCalledUnmanagedWhenTheJournalCouldNotBeRead` | — | yes |
| B3 | 계좌가 여럿이면 계좌 열이 뜬다 | `positionsView.Multi()` 경로의 portfolio_test 다계좌 케이스 | — | yes |
| B4 | 포지션 읽기 실패 — 부분 답 유지 + 사유 | `TestAHoldingIsNotCalledUnmanagedWhenTheJournalCouldNotBeRead` | — | yes |
| B5 | 0행 계좌는 계좌 목록에 들어가지 않는다 | `TestAnEmptyJournalSaysSoRatherThanLookingLikeAMissingOne`, `TestTheDashboardMasksTheAccountNumber` | — | yes |
