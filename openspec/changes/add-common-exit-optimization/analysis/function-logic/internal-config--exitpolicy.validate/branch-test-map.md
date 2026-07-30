# Branch Test Map: `ExitPolicy.validate`

- Source: `internal/config/engine.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if` path at line 61 and its complement/boundary | `TestMissingCommonExitPolicyPreservesLegacyRatchetSelection`; `TestSaveCommonExitPolicyChangesOnlyItsValueBlock`; config engine tests | yes | yes |
| B2 | `if` path at line 64 and its complement/boundary | `TestMissingCommonExitPolicyPreservesLegacyRatchetSelection`; `TestSaveCommonExitPolicyChangesOnlyItsValueBlock`; config engine tests | yes | yes |
