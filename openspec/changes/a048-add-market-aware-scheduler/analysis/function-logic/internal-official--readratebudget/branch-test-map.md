# Branch Test Map: `readRateBudget`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | no rate headers remains unreported | `TestAResponseWithNoRateHeadersIsNotZeroRemaining` | existing | yes |
| B2 | absent reset remains `ResetAbsent` | `TestTheReportedBudgetIsKept` | existing | yes |
| B3 | delta below threshold derives from observed-at without overflow | `TestTheResetHeaderIsReadBothWaysAndSaysWhich`, parser boundary table | scheduler duplicated unsafe arithmetic | yes |
| B4 | exact threshold is epoch, not delta | parser boundary table | scheduler accepted raw-kind mismatch | yes |
| B5 | earlier than -1m, later than +24h, extreme integer, and epoch implausibility are unparsed | implausible/parser boundary tests | scheduler could accept manually constructed impossible resets | yes |
