# Branch Test Map: `Journal.AppendStrategyExecutionLink`

| Branch | Scenario | Test | RED | GREEN |
|---|---|---|---|---|
| B1 | invalid account/attempt/kind/ref | existing lineage input tests | existing | existing |
| B2 | account/attempt mismatch | account-scope reverse lookup test | existing | existing |
| B3 | first append and reverse lookup | `TestStrategyLineageRestartReverseLookupAndAccountScope` | existing | existing |
| B4 | exact replay after clock advances | `TestStrategyExecutionLinkReplayPreservesFirstTimestamp` | second timestamp collided | pass |
| B5 | other-attempt unique collision | divergent execution tests | existing | existing |
