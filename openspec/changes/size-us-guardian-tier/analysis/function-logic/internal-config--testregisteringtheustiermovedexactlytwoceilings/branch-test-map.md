# Branch Test Map: `TestRegisteringTheUSTierMovedExactlyTwoCeilings`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | USD 상한을 읽는다 | 자기 자신 | no | yes |
| B2 | 주문 상한이 500이 아니다 | 자기 자신 | yes (300) | yes |
| B3 | 노출 상한이 1,500이 아니다 | 자기 자신 | yes (1000) | yes |
| B4 | 수량 상한이 움직였다 | 자기 자신 | no | yes |
| B5 | 비율 상한이 움직였다 | 자기 자신 | no | yes |
| B6 | 일일 손실 상한이 움직였다 | 자기 자신 | no | yes |
| B7 | KRW 상한을 읽는다 | 자기 자신 | no | yes |
| B8 | KRW 상한이 움직였다 | 자기 자신 | no | yes |
