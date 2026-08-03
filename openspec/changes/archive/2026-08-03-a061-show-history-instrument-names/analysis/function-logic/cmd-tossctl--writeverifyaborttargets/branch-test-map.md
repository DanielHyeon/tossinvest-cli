# Branch Test Map: `writeVerifyAbortTargets`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | no live target is stated rather than hidden | existing abort empty-record tests | existing | existing |
| B2 | every live target is rendered | existing abort list tests, `TestA061AbortReloadsOutstandingTargetsAfterExclusiveAdmission` | existing | yes |
| B3 | held target includes the awaited verdict | verifylive held-chain tests | existing | existing |
