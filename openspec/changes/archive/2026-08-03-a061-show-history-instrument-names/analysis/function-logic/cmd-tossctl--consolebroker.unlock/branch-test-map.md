# Branch Test Map: `consoleBroker.unlock`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | successful reads release the gate for the next metadata/account operation | A061 concurrent shared-client and cancellation tests under `-race` | yes | yes |
