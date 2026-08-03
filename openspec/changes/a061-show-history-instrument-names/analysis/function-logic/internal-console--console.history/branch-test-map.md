# Branch Test Map: `Console.history`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | missing/unreadable journal is explicit | `TestTheHistoryScreenIsHonestWhenNothingHasClosed` | existing green | existing green |
| B2 | account enumeration failure is not empty success | journal read-only failure coverage | existing green | existing green |
| B3 | every account is projected independently | account-view/history multi-account coverage | existing green | existing green |
| B4 | trip query failure does not render mixed partial results | journal read-only failure coverage | existing green | existing green |
| B5 | every frozen trip is retained | `TestTheHistoryScreenShowsTheFrozenRoundTrip` | existing green | existing green |
| B6 | event query failure does not render mixed partial results | journal read-only failure coverage | existing green | existing green |
| B7 | bounded event window states truncation | `TestTheHistoryScreenStatesItsOwnLimit` | existing green | existing green |
| B8 | every bounded exit event is retained | `TestTheHistoryScreenStatesItsOwnLimit` | existing green | existing green |
| B9 | account ownership appears only for accounts with rows | history multi-account coverage | existing green | existing green |
| B10 | completed trips attach market-keyed optional names | `TestA061HistoryShowsCodeAndNameForTripsAndEvents` | yes | yes |
| B11 | exit events attach the same optional names | `TestA061HistoryShowsCodeAndNameForTripsAndEvents` | yes | yes |
