# Branch Test Map: `httpAPIReader.Snapshot`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | engine 조회 실패는 집계 오류 | 기존 `httpapi_reader_test.go` | no | yes |
| B2 | positions 실패는 집계 오류 | 기존 | no | yes |
| B3 | orders 실패는 집계 오류 | 기존 | no | yes |
| B4 | candidates 실패는 집계 오류 | 기존 | no | yes |
| B5 | performance 실패는 집계 오류 | 기존 | no | yes |
| B6 | settings 실패는 집계 오류 | 기존 | no | yes |
| B7 | optimization 실패는 집계 오류 | 기존 | no | yes |
| B8 | reader 가 **부재**면 dormant (wrapper 라도) | `TestAnAbsentStrategyReaderStaysDormantRatherThanUnavailable` · `TestTheDaemonAttachesWhenTheEngineComesUpLater` | yes (a109 §2.3 RED: 부재가 영원히 고착) | yes |
| B9 | 붙었는데 못 읽으면 unavailable, 집계는 산다 | `TestALiveStrategyProjectionThatDiesLeavesTheAggregateStanding` | no (a108 핀) | yes |
| B10 | 읽었으면 그 값 | `TestALiveStrategyProjectionThatDiesLeavesTheAggregateStanding` | no | yes |
