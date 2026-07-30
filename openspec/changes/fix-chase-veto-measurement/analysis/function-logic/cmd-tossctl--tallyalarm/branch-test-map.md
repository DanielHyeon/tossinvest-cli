# Branch Test Map: `tallyAlarm`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | kind별로 다른 문장 | `TestEveryAnomalyKindStatesItsOwnArithmetic` | yes | yes |
| B2 | 임계 없는 통과 | `TestTheScanOutputSaysSoWhenTheTallyContradictsItself` · `TestAPassBesideAnAbsentThresholdIsADirectContradiction`(상류 판정) | yes | yes |
| B3 | 버킷 합이 전체를 넘음 | `TestTheScanOutputSaysSoWhenTheTallyContradictsItself`(near_high 12) · `TestAnUnmeasuredVetoCountedAsAPassIsCaughtByTheArithmetic`(상류) | yes | yes |
