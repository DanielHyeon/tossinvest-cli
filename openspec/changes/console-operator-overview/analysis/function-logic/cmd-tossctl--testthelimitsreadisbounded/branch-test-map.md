# Branch Test Map: `TestTheLimitsReadIsBounded`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 소스를 읽는다 | 이 테스트 | no (자기 진단) | yes |
| B2 | 무경계 Load 부재 | 동일 | yes (`context.Background()` 직접 전달 시 FAIL) | yes |
| B3 | WithTimeout 존재 | 동일 | yes (타임아웃 제거 시 FAIL) | yes |
| B4 | 타임아웃 상수가 양수 | 동일 | yes (0으로 바꾸면 FAIL) | yes |
