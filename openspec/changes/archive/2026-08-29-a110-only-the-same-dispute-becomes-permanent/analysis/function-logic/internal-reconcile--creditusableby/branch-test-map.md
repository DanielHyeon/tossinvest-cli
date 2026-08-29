# Branch Test Map: `creditUsableBy`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | malformed/missing credit stamp is unusable | a083 credit time-order tests | preserve | yes |
| B2 | malformed/missing observation stamp is unusable | a083 credit time-order tests | preserve | yes |
| Return | later disputed observation refutes; equal/unrelated stays | `TestA110AuthorityOutageStillRefutesUsableAdjustmentCredit` plus a083 release suite | yes (M28) | yes |
