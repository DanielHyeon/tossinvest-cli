# Branch Test Map: `StrategyRuntimeSnapshotFunc`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 부재(nil 또는 부재를 말하는 wrapper)는 스냅샷을 만들지 않는다 | `TestTheStreamHelperRefusesAnUnconfiguredWrapper` | no — a109 편집과 같은 커밋에서 도입한 핀 | yes |
| B2 | 붙었는데 못 읽으면 오류 | `TestStrategyRuntimeRESTDormantHealthAndSSEFullSnapshotParity` | no | yes |
| B3 | 계약 위반 스냅샷은 오류 | 기존 계약 테스트 | no | yes |
