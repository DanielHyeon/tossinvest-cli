# Branch Test Map: `mergeEngine`

- Source: `internal/config/engine.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if` path at line 299 and its complement/boundary | `TestMissingCommonExitPolicyPreservesLegacyRatchetSelection`; `TestSaveCommonExitPolicyChangesOnlyItsValueBlock`; config engine tests | yes | yes |
| B2 | `if` path at line 303 and its complement/boundary | `TestMissingCommonExitPolicyPreservesLegacyRatchetSelection`; `TestSaveCommonExitPolicyChangesOnlyItsValueBlock`; config engine tests | yes | yes |
| B3 | `if` path at line 308 and its complement/boundary | `TestMissingCommonExitPolicyPreservesLegacyRatchetSelection`; `TestSaveCommonExitPolicyChangesOnlyItsValueBlock`; config engine tests | yes | yes |
| B4 | `if` path at line 312 and its complement/boundary | `TestMissingCommonExitPolicyPreservesLegacyRatchetSelection`; `TestSaveCommonExitPolicyChangesOnlyItsValueBlock`; config engine tests | yes | yes |
