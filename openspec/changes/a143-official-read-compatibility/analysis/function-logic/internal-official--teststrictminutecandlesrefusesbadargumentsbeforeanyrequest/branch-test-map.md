# Branch Test Map: `TestStrictMinuteCandlesRefusesBadArgumentsBeforeAnyRequest`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | iterate invalid market/symbol/count/before cases | TestStrictMinuteCandlesRefusesBadArgumentsBeforeAnyRequest | not targeted; these forms were already invalid | named test PASS |
| B2 | err == nil would fail the test | TestStrictMinuteCandlesRefusesBadArgumentsBeforeAnyRequest | not targeted | assertion evaluated; fatal path not taken |
| B3 | nonzero request count would fail the test | TestStrictMinuteCandlesRefusesBadArgumentsBeforeAnyRequest | not targeted | assertion evaluated; fatal path not taken |

Typed refusal is separately asserted by the strictMinuteRefusal helper between B2 and B3; it is not B2's condition. No instrumented branch-coverage claim.
