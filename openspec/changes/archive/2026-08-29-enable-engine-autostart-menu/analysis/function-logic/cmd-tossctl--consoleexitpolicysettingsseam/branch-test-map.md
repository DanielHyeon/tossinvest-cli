# Branch Test Map: `consoleExitPolicySettingsSeam`

- Source: `cmd/tossctl/console.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if` path at line 360 and its complement/boundary | `TestExitPolicySeamSavesAuditsAndPreservesUnrelatedConfig`; console characterization tests | yes | yes |
