# Branch Test Map: `AggregateClosedKRXFiveMinute`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | legacy raw caller cannot mint proof | replacement-page fail-closed tests plus source refusal | yes | yes |
| B2 | wrong symbol/adjusted/zero page | `TestAggregateClosedKRXFiveMinuteFailsClosed` | yes | yes |
| B3 | exact decimals and official identity preserved | `TestAggregateClosedKRXFiveMinutePreservesExactDecimals` | yes | yes |
