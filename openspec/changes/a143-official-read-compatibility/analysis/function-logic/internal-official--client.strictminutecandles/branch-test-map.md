# Branch Test Map: `Client.StrictMinuteCandles`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B3 | `0101N0` accepts with exact query and `Page.Symbol` | `TestStrictMinuteCandlesAcceptsDomesticAlphanumericAndPreservesRequestSymbol` | pending RED | pending GREEN |
| B4 | numeric `005930` stays accepted | `TestStrictMinuteCandlesAcceptsTheKoreanMarket` | unchanged branch; not RED target | pending GREEN |
| B5 | invalid KR forms refuse before request | `TestStrictMinuteCandlesRefusesBadArgumentsBeforeAnyRequest` | pending RED | pending GREEN |
| B6 | US policy remains unchanged | `TestStrictMinuteCandlesRefusesBadArgumentsBeforeAnyRequest` | unchanged branch; not RED target | pending focused run |
| B7 | `0101N0` retains raw newest-first close label/cursor | `TestStrictMinuteCandlesAcceptsDomesticAlphanumericAndPreservesRequestSymbol` | pending RED | pending GREEN |
| I1 | officialbars converts close label once | `TestPollConvertsTheBrokerCloseLabelIntoAnOpenInstant` | unchanged branch; not RED target | pending focused run |
| B1 | nil client refuses | `TestStrictMinuteCandlesRefusesANilClient` | unchanged branch; not RED target | unchanged |
| B2 | invalid market refuses | `TestStrictMinuteCandlesRefusesBadArgumentsBeforeAnyRequest` | unchanged branch; not RED target | focused pass |
| B8 | outer observer chains | `TestStrictMinuteCandlesUsesTheLastSuccessfulAttemptAndChainsTheOuterObserver` | unchanged branch; not RED target | unchanged |
| B9 | transport failure propagates | `TestStrictMinuteCandlesPropagatesTransportClassification` | unchanged branch; not RED target | unchanged |
| B10 | final-attempt failure refuses | `TestStrictMinuteCandlesUsesTheLastSuccessfulAttemptAndChainsTheOuterObserver` | unchanged branch; not RED target | unchanged |
| B11 | strict decoder failure refuses | `TestStrictMinuteCandlesRefusesMalformedBodies` | unchanged branch; not RED target | unchanged |
