# Branch Test Map: `Decide`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1-B4 | OFF, missing activation, scope mismatch and missing calendar close before entry | `TestDecisionStatesAreTypedAndOrderedFailClosed` | existing | yes |
| B5 | fresh but newly versioned calendar cannot reuse old approval | `TestDecisionRejectsCalendarVersionChangedAfterActivation` | yes, `ENTRY_ALLOWED` | yes |
| B6-B7 | stale/late/skewed or wrong-day calendar waits | calendar freshness/malformed tests + decision matrix | existing | yes |
| B8-B10 | holiday, pre-open and post-close wait with next transition | `TestHolidayWaitsWithNextSession`, decision matrix | existing | yes |
| B11-B12 | missing/exhausted provenance defers | scheduler budget tests + decision matrix | existing | yes |
| success | exact activation/calendar and discretionary grant allow entry | `TestDecisionStatesAreTypedAndOrderedFailClosed/allowed` | existing | yes |
| B1 | desired OFF | decision matrix/disabled | existing | yes |
| B2 | activation absent or stale | activation tests | existing | yes |
| B3 | market outside desired scope | decision matrix | existing | yes |
| B4 | calendar absent | decision matrix | existing | yes |
| B6 | calendar market/freshness invalid | calendar validity tests | existing | yes |
| B7 | exchange-local day mismatch | malformed calendar tests | existing | yes |
| B8 | holiday | holiday test | existing | yes |
| B9 | before regular open | decision matrix | existing | yes |
| B10 | at/after regular close | decision/calendar boundary suite | existing | yes |
| B11 | coordinator absent | decision matrix | existing | yes |
| B12 | coordinator denies | budget reserve tests | existing | yes |
