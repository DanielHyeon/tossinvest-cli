# Branch Test Map: `consoleBroker.resolve`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 두 번째 이후 호출은 캐시 — 3회 새로고침에 구축 1회 | `TestTheOrdersSeamResolvesTheAccountOnceAndBuildsNoSecondClient` | yes (캐시 없으면 built=3) | yes |
| B2 | factory 실패는 캐시되지 않고 그대로 올라간다 | 동일 테스트의 factory 대체 경로 + 대시보드 미측정 렌더 | yes | yes |
| (else) | 두 화면을 동시에 열어도 구축 1회 — 락이 직렬화한다 | `TestOpeningEveryConsoleReadScreenResolvesTheAccountOnce` (동시 개시 + `-race`) | yes (seam별 resolver로 2회) | yes (1회) |
