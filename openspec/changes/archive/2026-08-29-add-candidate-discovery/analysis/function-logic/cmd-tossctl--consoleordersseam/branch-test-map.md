# Branch Test Map: `consoleOrdersSeam`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | (무분기 happy path) 반환값이 정확히 `Orders` 한 메서드만 선언하고, `console.Options`가 이 필드를 받는다 | `TestTheConsoleIsHandedOneCapabilityAndNotABroker` (cmd/tossctl) + `TestEveryCapabilityTheConsoleReceivesIsEnumeratedAndDeclaresNothingButReads` (internal/console) | yes (Orders 필드 부재로 실패) | yes |
| B1 (배선) | 인자가 포지션 화면과 **같은** resolver다 | `TestOpeningEveryConsoleReadScreenResolvesTheAccountOnce` | yes (seam별 resolver로 계좌 해석 2회) | yes (1회) |

세 콜의 실제 배선은 `lazyOrders.Orders`의 Branch Test Map이 소유한다.
