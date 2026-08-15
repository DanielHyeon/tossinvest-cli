# Branch Test Map: `TestProductionAPIExportsNoAuthorityMintingFunction`

Source: `internal/protectionlifecycle/external_api_test.go` (11-30).

| Branch | Scenario | Test |
|---|---|---|
| B1 | directory read is required | `TestProductionAPIExportsNoAuthorityMintingFunction` |
| B2 | all package files iterate | `TestProductionAPIExportsNoAuthorityMintingFunction` |
| B3 | only production Go files pass | `TestProductionAPIExportsNoAuthorityMintingFunction` |
| B4 | parser failure fails test | `TestProductionAPIExportsNoAuthorityMintingFunction` |
| B5 | every scope object is checked | `TestProductionAPIExportsNoAuthorityMintingFunction` |
| B6 | exported function is forbidden | `TestProductionAPIExportsNoAuthorityMintingFunction` |

Planned A100 review assertion: no lifecycle export can mint or restore worker/entry authority.
