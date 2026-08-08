# Branch Test Map: `resolveAccountRef`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | Nil official reader refuses startup | `TestResolveAccountRefRejectsNilOfficialClient` | defensive baseline | yes |
| B2 | Account-list error refuses startup | `TestStartupRefusesWhenTheAccountCannotBeResolved/broker_error` | baseline | yes |
| B3 | Empty list refuses | existing account-resolution coverage | baseline | yes |
| B4 | Blank first number with a later valid record refuses rather than splitting identities | `TestEngineRefusesAnIncompleteFirstAccountRecordBeforeOpeningTheJournal/blank_account_number` | yes | yes |
| B5 | Zero or negative first sequence refuses even when the number is present | `TestEngineRefusesAnIncompleteFirstAccountRecordBeforeOpeningTheJournal` | yes | yes |
| B6 | Selected sequence is invalid or differs from the first record | `TestEngineRefusesAnExplicitSequenceThatDoesNotMatchTheFirstRecord` | yes | yes |

Success path: `TestActualEngineRecoveryReusesTheStartupAccountSequence` failed on
the duplicate account 429 in RED and passes with one discovery and identical
headers in GREEN.
