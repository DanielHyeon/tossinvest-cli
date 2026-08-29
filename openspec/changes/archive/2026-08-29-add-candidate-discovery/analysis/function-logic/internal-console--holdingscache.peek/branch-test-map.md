# Branch Test Map: `holdingsCache.peek`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 브로커 seam 미배선 상태에서 개요를 연다 | `brokerReadable`의 `seam_unwired` 경로(`TestAnUnwiredGateLimitsSeamOnlyDarkensItsOwnPanel`과 같은 패널 격리 형태) | — | yes |
| B2 | (무분기 본문) 한 번도 채워지지 않은 캐시를 읽어도 0콜이고 `Held`가 비어 있다 | `TestThereIsAReadOfTheBrokerCacheThatNeitherRefreshesNorInventsAReason` | yes — `peek` 이전에는 `get(…, true, …)`뿐이었고 그 경로는 `Held=true`를 실었다 | yes |
| B3 | (무분기 본문) TTL을 넘긴 캐시를 읽어도 갱신하지 않고 나이를 그대로 보고한다 | `TestPeekServesWhatTheLastRefreshFoundAndStillCallsNothing`(90초 경과 후 호출 수 1 유지) | — | yes |
