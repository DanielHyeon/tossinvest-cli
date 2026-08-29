# Branch Test Map: `TestThePopularityRankingRefusesAMarketItCannotSee`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | US 스캔 거부 | 자체 실행 | yes (컴파일) | yes |
| B2 | 거부가 호출보다 앞 | 자체 실행 | — (기존 동작) | yes |

이 경로가 `rememberRead`보다 앞이라는 것도 참이다 — US 읽기는 KR 기억을 건드리지 않는다.
그 사실 자체는 `TestOneSourceServingTwoMarketsDoesNotAnswerAboutTheWrongList`가 공식
어댑터 쪽에서 잡고, WTS 쪽에서는 **잡히지 않는다**(이 소스가 한 시장만 섬기므로).
