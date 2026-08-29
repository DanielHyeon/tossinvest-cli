# Branch Test Map: `TestTheMeasuredInstrumentFitsWithHeadroom`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | USD 상한을 읽는다 | 자기 자신 | no | yes |
| B2 | 관측 1주가 상한을 넘는다 | 자기 자신 | no (경계 포함으로 통과) | yes |
| B3 | 헤드룸이 50% 미만 | 자기 자신 | yes (0.0%) | yes |
