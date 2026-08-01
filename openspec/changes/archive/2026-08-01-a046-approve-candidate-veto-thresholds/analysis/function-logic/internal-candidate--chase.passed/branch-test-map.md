# Branch Test Map: `Chase.Passed`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | all D3 codes are consulted through an immutable copy | `TestRemovingAVetoCodeCannotRemoveItsVeto` | exported ordering was writable | copy mutation cannot change iteration |
| B2 | raised or unmeasured code exists | zero/absent/dropped-input veto tests | existing fail-closed coverage | false |
