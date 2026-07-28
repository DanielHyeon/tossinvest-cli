# Branch Test Map: `candidateTallyAlarms`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 모순이 있으면 목록이 차고 없으면 빈다 | `TestTheScanOutputSaysSoWhenTheTallyContradictsItself`(찬다) · `TestTheOrdinaryNoThresholdScanRaisesNoAlarm`(임계가 하나도 없는 매일의 상태에서 비어 있다) | yes | yes |

음성 대조군(`TestTheOrdinaryNoThresholdScanRaisesNoAlarm`)이 이 경보가 가질 가치가 있는지를
결정한다 — 매일의 정상 상태에서 발화하면 출시일부터 소음이다.
