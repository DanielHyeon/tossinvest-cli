# Branch Test Map: `holdingsCache.snapshotLocked`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 채워진 캐시는 나이를 보고하고, 빈 캐시는 `Present=false`로 남는다 | `TestPeekServesWhatTheLastRefreshFoundAndStillCallsNothing`, `TestAColdBrokerCacheSaysNotYetReadAndLinksToTheScreenThatFillsIt` | — | yes |
| B2 | 시계가 판독 시각보다 이른 경우 나이를 0으로 고정 | 직접 테스트 없음 — 음수 나이 방어이며 신설 시 함께 옮겨온 줄이다 | — | n/a |
