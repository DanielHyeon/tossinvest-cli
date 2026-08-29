# Branch Test Map: `parseRatio`

- Source: `internal/exitpolicy/decimal.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if` path at line 82 and its complement/boundary | `TestCommonPolicyRegistryHasExactlyTheThreeApprovedProfiles`; ladder and decimal boundary tests | yes | yes |
| B2 | `if` path at line 86 and its complement/boundary | `TestCommonPolicyRegistryHasExactlyTheThreeApprovedProfiles`; ladder and decimal boundary tests | yes | yes |
| B3 | `if` path at line 91 and its complement/boundary | `TestCommonPolicyRegistryHasExactlyTheThreeApprovedProfiles`; ladder and decimal boundary tests | yes | yes |
