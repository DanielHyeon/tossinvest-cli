# Branch Test Map: `TestTheOrdersSeamResolvesTheAccountOnceAndBuildsNoSecondClient`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 3회 반복 새로고침 | 이 테스트 | yes | yes |
| B2 | 새로고침 중 에러 없음 | 동일 | yes | yes |
| B3 | 브로커 구축 1회 — 매 호출 구축이면 3 | 동일 | yes (캐시 제거 변이로 built=3) | yes |
| B4 | console.go에 `official.New(` 없음 | 동일 | yes (두 번째 클라이언트 삽입 시 FAIL) | yes |
| B5 | console.go의 주문 seam이 `l.shared.resolve()`를 통과 | 동일 | yes | yes |

이 표의 범위는 seam 하나다. 콘솔 세션 전체(포지션+/orders)의 해석 1회는 `TestOpeningEveryConsoleReadScreenResolvesTheAccountOnce`의 Branch Test Map이 소유한다.
