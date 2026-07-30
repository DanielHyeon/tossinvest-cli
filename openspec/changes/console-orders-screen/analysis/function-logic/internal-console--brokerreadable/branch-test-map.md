# Branch Test Map: `brokerReadable`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 네 상태가 각자 다른 사유로 갈린다 | `TestTheUnmeasuredReasonsStayApart` | — (동작 무변경) | yes |
| B2 | 브로커 seam 미배선 | 패널 격리 테스트 계열(`TestAnUnwiredGateLimitsSeamOnlyDarkensItsOwnPanel`과 같은 형태) | — (동작 무변경) | yes |
| B3 | 채워진 캐시는 측정됨 | `TestPeekServesWhatTheLastRefreshFoundAndStillCallsNothing` 이후 개요 렌더 | — (동작 무변경) | yes |
| B4 | 실패한 브로커 읽기는 빈 캐시와 다르다 | `TestAFailedBrokerReadIsNotTheSameAsAnEmptyCache` | — (동작 무변경) | yes |
| B5 | 한 번도 읽지 않은 캐시는 `never_fetched`이고, 검증이 돌고 있어도 그렇다 | `TestAColdBrokerCacheSaysNotYetReadAndLinksToTheScreenThatFillsIt`, `TestARunningVerificationIsANoteAndNotTheReasonTheCacheIsCold`, `TestAColdCacheWithNoVerificationSaysNothingAboutOne` | yes — I-5 이전 코드는 여기서 `verify_suspended`를 답했고 리뷰가 그것을 잡았다 | yes |
