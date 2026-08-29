# Branch Test Map: `TestTheRoutePathIsReadOutOfEitherQuotationForm`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 세 리터럴 형태 | 이 테스트 자신 | yes — `strings.Trim(value, "\"")`로 되돌리면 raw string 케이스가 문다 | yes |
| B2 | 불일치 보고 | 같은 위 | yes | yes |
