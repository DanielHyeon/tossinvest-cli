# Branch Test Map: `newConsoleBroker`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | (무분기 happy path) `runConsole`이 정확히 하나를 만들고 모든 읽기 seam이 그것을 받는다 | `TestOpeningEveryConsoleReadScreenResolvesTheAccountOnce` (런타임 1회 + `runConsole` 소스 파싱) | yes (seam별 resolver로 계좌 해석 2회) | yes (1회) |
