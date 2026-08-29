# Branch Test Map: `newConsoleHoldings`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | (무분기 happy path) 반환값이 정확히 `Holdings` 한 메서드만 선언한다 | `TestTheConsoleIsHandedOneCapabilityAndNotABroker` | yes (두 번째 메서드 변이로 확인 가능) | yes |
| B1 (배선) | 인자가 /orders와 **같은** resolver다 | `TestOpeningEveryConsoleReadScreenResolvesTheAccountOnce` | yes (seam별 resolver로 계좌 해석 2회) | yes (1회) |
