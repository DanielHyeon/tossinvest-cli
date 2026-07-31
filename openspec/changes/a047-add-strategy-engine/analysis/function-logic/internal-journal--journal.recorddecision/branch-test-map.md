# Branch Test Map: `Journal.RecordDecision`

| Branch | Scenario | Test | RED observed | GREEN observed |
| --- | --- | --- | --- | --- |
| B1 | missing/invalid identity, account, nonce, class, preimage or window refuses | `TestRecordDecisionRefusals`, reduction/limits tests | baseline | baseline |
| B2 | duplicate nonce/ID or cancelled DB write returns failure | nonce and durability tests | baseline | baseline |
| B3 | canonical decision round-trips and binds attempt | `TestRecordAndLookupDecision`, `TestPrepareRecordsTheDecisionBinding`, `TestDecisionBindingIsImmutableAcrossTransitions` | baseline | baseline |
| A047 | decision/attempt retains lane, candidate and manifest digest across restart | a047 provenance journal tests (to add in RED) | pending | no |
