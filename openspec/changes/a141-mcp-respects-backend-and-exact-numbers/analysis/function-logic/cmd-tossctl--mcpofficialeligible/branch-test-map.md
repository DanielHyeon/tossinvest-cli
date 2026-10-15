# Branch Test Map: `mcpOfficialEligible`

AST has no branch IDs. The checker requires a synthetic B1 happy-path row for branchless functions (check_analysis.py); B1 below labels the return expression and is not an AST branch.

| Condition | Scenario | Test | Regression evidence | GREEN observed |
|---|---|---|---|---|
| B1 | nonnil/nil credentials; enabled/disabled; WTS/auto/openapi | TestMCPOfficialEligibilityMatchesCLI | Enabled and WTS removal mutants detected | Terra focused PASS |
| caller use | actual effective dependency publication | TestMCPCmdWTSDoesNotWireOfficialDependencies; TestMCPCmdDisabledAndEnabledBackendWiring | WTS removal mutant detected by actual auth status | Terra guarded PASS |
