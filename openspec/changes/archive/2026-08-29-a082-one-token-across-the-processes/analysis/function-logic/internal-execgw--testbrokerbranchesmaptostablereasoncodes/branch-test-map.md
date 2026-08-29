# Branch Test Map: `TestBrokerBranchesMapToStableReasonCodes`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 표를 돈다. 감싼 인증 거부 2행이 새로 들어갔다 | 자기 자신 | **yes** — `isAuthRefusal`을 `==`로 바꾸면 두 새 행이 깨진다 (실측) | yes |
| B2 | 분류에 실패하면 보고한다 | 자기 자신 (손대지 않음) | no | yes |
| B3 | 기대 reason과 다르면 보고한다 | 자기 자신 (손대지 않음) | no | yes |
| B4 | 분류되지 않아야 할 입력이 분류되면 보고한다 | 자기 자신 (손대지 않음) | no | yes |
| B5 | 기존 운영자 분기(거래 인증) 판정 | 자기 자신 (손대지 않음) | no | yes |
| B6 | 기존 운영자 분기(환전 동의) 판정 | 자기 자신 (손대지 않음) | no | yes |
| B7 | 기존 운영자 분기(입금) 판정 | 자기 자신 (손대지 않음) | no | yes |
