# Branch Test Map: `router.read`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 자원 이름으로 갈린다 | `TestStrategyRuntimeRESTUsesSharedProjectionAndStrictGuards` | no | yes |
| B2 | engine 자원 | 기존 REST 계약 테스트 (`router_test.go` 계열) | no | yes |
| B3 | positions 자원 | 기존 | no | yes |
| B4 | positions 목록 nil → 빈 배열 | 기존 | no | yes |
| B5 | orders 자원 | 기존 | no | yes |
| B6 | orders 목록 nil → 빈 배열 | 기존 | no | yes |
| B7 | candidates 자원 | 기존 | no | yes |
| B8 | candidates 목록 nil → 빈 배열 | 기존 | no | yes |
| B9 | performance 자원 | 기존 | no | yes |
| B10 | settings 자원 | 기존 | no | yes |
| B11 | settings 목록 nil → 빈 배열 | 기존 | no | yes |
| B12 | optimization 자원 | 기존 | no | yes |
| B13 | strategy-runtime 자원 | `TestStrategyRuntimeRESTUsesSharedProjectionAndStrictGuards` | no | yes |
| B14 | **부재는 200 + dormant** (nil 이든 부재를 말하는 wrapper 든) | `TestTheRESTRouteStaysDormantForAnUnconfiguredWrapper` | no — a109 편집과 같은 커밋에서 도입한 핀 | yes |
| B15 | 붙었는데 못 읽으면 오류 | `TestStrategyRuntimeRESTDormantHealthAndSSEFullSnapshotParity` | no | yes |
| B16 | 계약 위반 스냅샷은 오류 | 기존 | no | yes |
| B17 | 모르는 자원은 오류 | 기존 | no | yes |
