# Branch Test Map: `sameExitOperationalLine`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | Branch-free happy path: all operational/recovery/display-affecting fields must match; observed price and provenance identities intentionally vary on a heartbeat | `TestA111OperationalEqualityChecksEveryD1FieldUsingEvaluatorGeneratedDonors` | intentional A111 RED before production change | asserted by focused A111 suite |
