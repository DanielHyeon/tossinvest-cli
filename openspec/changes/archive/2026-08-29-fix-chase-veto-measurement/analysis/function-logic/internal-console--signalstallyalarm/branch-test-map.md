# Branch Test Map: `signalsTallyAlarm`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | kind별 문장 | `TestTheSignalsScreenSaysSoWhenTheTallyContradictsItself` | yes | yes |
| B2 | 임계 없는 통과 | 동상 | yes | yes |
| B3 | 버킷 합 초과 | 동상 | yes | yes |
| (음성 대조) | 매일의 정상 상태에서는 아무 경보도 없다 | `TestTheOrdinarySignalsScreenRaisesNoAlarm` | yes | yes |
