# Branch Test Map: `SetA087LowerLimitForTest`

| Branch | Scenario | Evidence |
|---|---|---|
| options nil | 변경 없음 | 손 읽기, 별도 시험 미실행 |
| options non-nil | 활성 판 fixture 주입 | `TestA087RealGatewayLowerLimitAndDormantFallback`, A087 집중 race GREEN |
