# Branch Test Map: `lazyHoldings.Holdings`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 자격증명 없음 → 해석 실패가 화면 위 문장이 되고 기억되지 않는다 | 대시보드 미측정 렌더 테스트 + `TestOpeningEveryConsoleReadScreenResolvesTheAccountOnce`의 factory 경로 | no (기존 실패 계약 유지) | yes |
| (else) | 포지션 화면과 /orders를 모두 열어도 계좌 해석은 1회 | `TestOpeningEveryConsoleReadScreenResolvesTheAccountOnce` | yes (seam별 resolver로 2회 관측) | yes (1회) |
