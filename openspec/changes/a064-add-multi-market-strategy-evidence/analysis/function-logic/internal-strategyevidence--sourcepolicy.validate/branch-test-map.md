# Branch Test Map: `SourcePolicy.validate`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | an authority outside the four known ones makes zero calls | `TestEveryPolicyFieldHasAZeroCallRefusal` | existing coverage | PASS |
| B2 | an unverified contract makes zero calls | `TestEveryPolicyFieldHasAZeroCallRefusal` | existing coverage | PASS |
| B3 | KRX stays typed-unavailable while its contract is unfrozen, read from the frozen contract file | `TestKRXStaysUnavailableWhileItsContractIsNotFrozen` | no RED: the mechanism held; the file that declares it was read by nothing (issues.md I11) | PASS |
| B4 | the seven required strings are walked, not sampled | `TestEveryPolicyFieldHasAZeroCallRefusal` | existing coverage | PASS |
| B5 | each required string, blanked one at a time, makes zero calls | `TestEveryPolicyFieldHasAZeroCallRefusal` | existing coverage | PASS |
| B6 | each positivity/ordering bound, violated one at a time, makes zero calls | `TestEveryPolicyFieldHasAZeroCallRefusal` | seal re-stamped per case, so the field check is the only thing that can refuse; deleting `RetryableStatuses`/`RetryAfterPolicy` from this arm fails exactly those two subtests (measured under `-overlay`) | PASS |
| B7 | page size 0/101 and each contract cap, exceeded one at a time, make zero calls | `TestEveryPolicyFieldHasAZeroCallRefusal` | nine subtests failed before this arm existed | PASS |
