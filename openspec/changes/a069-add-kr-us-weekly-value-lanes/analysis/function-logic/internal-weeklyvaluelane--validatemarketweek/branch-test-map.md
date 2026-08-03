# Branch Test Map: `ValidateMarketWeek`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | forged stable identity rejected | TestMarketWeekDerivesStableIdentityExactlyAndBoundsIt | yes | yes |
| B2 | stale calendar rejected at trusted evaluatedAt | TestReservationRejectsStaleCalendarAtCommandEvaluation | yes | yes |
| B3 | KR/US timezone and ISO week valid | TestOfficialIANAWeekIdentityIsStableAcrossCalendarCorrectionHolidayAndDST | existing | existing |
| B4 | oversized identity rejected | TestMarketWeekDerivesStableIdentityExactlyAndBoundsIt | yes | yes |
| B5 | official/provider/timezone mismatch refused | TestOfficialIANAWeekIdentityIsStableAcrossCalendarCorrectionHolidayAndDST | existing | existing |
| B6 | IANA location load fail closed | TestOfficialIANAWeekIdentityIsStableAcrossCalendarCorrectionHolidayAndDST | existing | existing |
| B7 | session parse/non-Monday refused | TestMarketWeekDerivesStableIdentityExactlyAndBoundsIt | yes | yes |
| B8 | ISO stable identity mismatch refused | TestMarketWeekDerivesStableIdentityExactlyAndBoundsIt | yes | yes |
| B9 | exact KR identity accepted | TestOfficialIANAWeekIdentityIsStableAcrossCalendarCorrectionHolidayAndDST | existing | existing |
| B10 | exact US DST identity accepted | TestMarketWeekDerivesStableIdentityExactlyAndBoundsIt | yes | yes |
