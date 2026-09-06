# Branch Test Map: `TestIncompleteOrUnverifiedPolicyMakesZeroCalls`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | four broken policies are each exercised | `TestIncompleteOrUnverifiedPolicyMakesZeroCalls` | existing coverage | PASS |
| B2 | each returns its expected sentinel | `TestIncompleteOrUnverifiedPolicyMakesZeroCalls` | existing coverage | PASS |
| B3 | none of them reaches the transport | `TestIncompleteOrUnverifiedPolicyMakesZeroCalls` | existing coverage | PASS |
