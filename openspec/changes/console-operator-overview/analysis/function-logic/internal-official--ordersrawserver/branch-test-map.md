# Branch Test Map: `ordersRawServer`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 커서가 있으면 인용 문자열, 없으면 JSON `null` | `TestTheRawOrderReadReportsThatThePageWasTruncated`(있음) / 나머지 3건(없음) | yes (헬퍼 부재로 컴파일 실패) | yes |
| B2 | 경로별 응답 분기 | 호출 테스트 4건 | yes | yes |
| B3 | 토큰 발급 | 동상 | yes | yes |
| B4 | 미체결 + 체결 두 건 페이지 | 동상 | yes | yes |
| B5 | 예상 밖 경로는 404 | 동상 | yes | yes |
