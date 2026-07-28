# Branch Test Map: `TestTheRecordDoesNotCallAUSRequestAKROne`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | US 실행이 성공한다 | 자체 실행 | yes (고정 문자열 시절 B2에서 실패) | yes |
| B2 | US 배치 detail이 KR을 말하지 않는다 | 자체 실행 | yes | yes |
| B3 | US 배치 detail이 US를 말한다 | 자체 실행 | yes | yes |
| B4 | US 정정 detail이 KR을 말하지 않는다 | 자체 실행 | yes | yes |
| B5 | US 정정 detail이 US를 말한다 | 자체 실행 | yes | yes |
| B6 | US 정정 detail이 없는 수량을 주장하지 않는다 | 자체 실행 | yes | yes |
