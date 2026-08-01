# Branch Test Map: `PositionPolicyCommandService.issueCapability`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | instance-bound issuance succeeds or fails closed | capability instance tests | yes | yes |
| B2 | iterate outstanding grants | expiry/capacity contract | yes | yes |
| B3 | expired grants are pruned before bounded admission | expiry/capacity contract | yes | yes |
| B4 | capacity failure issues no authority | bounded capability contract | yes | yes |
| B5 | entropy failure issues no authority | entropy contract | yes | yes |
| B6 | only RELEASE/READOPT receive mandatory delay | `TestPositionPolicyDangerousCapabilityRequiresServerSideConfirmation` | yes | yes |
