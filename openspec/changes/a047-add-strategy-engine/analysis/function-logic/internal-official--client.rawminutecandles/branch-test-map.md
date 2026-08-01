# Branch Test Map: `Client.RawMinuteCandles`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | unsupported market rejected pre-network | `TestRawMinuteCandlesRejectsUnsupportedMarketAndAdjustedStrategyReadRemainsExplicit` | yes | yes |
| B2 | fixed one-minute query and lossless DTO | `TestRawMinuteCandlesPreservesOfficialDecimalAndTimestampStrings` | yes | yes |
| B3 | page provenance retained | same preservation test | yes | yes |
| B4 | official transport failure | covered by shared `Client.get` tests; no new branch behavior | existing | existing |
| B5 | optional count query | preservation query assertion | existing | yes |
| B6 | optional cursor query | official read contract; empty cursor branch in preservation test | existing | yes |
