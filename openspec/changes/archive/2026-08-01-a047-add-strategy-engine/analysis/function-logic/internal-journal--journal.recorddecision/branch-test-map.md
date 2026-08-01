# Branch Test Map: `Journal.RecordDecision`

| Branch | Scenario | Test | RED observed | GREEN observed |
| --- | --- | --- | --- | --- |
| B1 | missing/invalid identity, account, nonce, class, preimage or window refuses | `TestRecordDecisionRefusals`, reduction/limits tests | baseline | baseline |
| B2 | duplicate nonce/ID or cancelled DB write returns failure | nonce and durability tests | baseline | baseline |
| Scenario | canonical decision build and insert both succeed | `TestRecordAndLookupDecision`; attempt binding belongs to prepare/transition functions, not `RecordDecision` | baseline | baseline |
| A047 | decision/attempt retains lane, candidate and manifest digest across restart | strategy lineage atomic/restart tests | missing | pass |
