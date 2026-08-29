# Branch Test Map: `observationDetail`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 단계 기록이 없으면 실패시킨다 | `TestTheRecordDoesNotCallAUSRequestAKROne`, `TestTheKRRecordStillSaysKR` | yes (헬퍼 부재로 컴파일 실패) | yes |
| B2 | 관측을 순회한다 | 동상 | yes | yes |
| B3 | 키가 맞는 관측의 detail을 돌려준다 | 동상(`order.place.ok`, `order.amend.ok`) | yes | yes |
