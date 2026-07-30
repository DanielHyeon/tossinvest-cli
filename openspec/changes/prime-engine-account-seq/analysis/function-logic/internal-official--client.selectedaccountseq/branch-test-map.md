# Branch Test Map: `Client.SelectedAccountSeq`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 분기 없는 atomic read가 valid primed selection을 반환하고 positive/negative mismatch는 engine에서 거부된다 | `TestActualEngineRecoveryReusesTheStartupAccountSequence`, `TestActualEngineRecoveryAcceptsAMatchingExplicitAccountSequence`, `TestEngineRefusesAnExplicitSequenceThatDoesNotMatchTheFirstRecord` | mismatch accepted in RED | yes |
