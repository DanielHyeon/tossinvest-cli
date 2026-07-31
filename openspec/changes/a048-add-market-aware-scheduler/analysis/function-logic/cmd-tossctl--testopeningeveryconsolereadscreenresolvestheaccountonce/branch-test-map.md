# Branch Test Map: `TestOpeningEveryConsoleReadScreenResolvesTheAccountOnce`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | persist KR desired state required to make the calendar screen read | `TestOpeningEveryConsoleReadScreenResolvesTheAccountOnce` | fixture prerequisite | yes |
| B2 | run two serial rounds | same test | existing coverage extended | yes |
| B3 | open each declared read screen in each round | same test | market schedule was absent | yes |
| B4 | any serial screen error fails with screen identity | same test | market schedule was absent | yes |
| B5 | all serial screens together build the broker once | same test | market schedule was absent | yes |
| B6 | start one concurrent goroutine per cold screen | same test | market schedule was absent | yes |
| B7 | inspect every concurrent result | same test | market schedule was absent | yes |
| B8 | any concurrent screen error fails with screen identity | same test | market schedule was absent | yes |
| B9 | all concurrent cold screens build the broker once | same test | market schedule was absent | yes |
| B10 | current `console.go` must parse | same test | existing | yes |
| B11 | scan every function declaration for factory calls | same test | existing | yes |
| B12 | skip non-function declarations | same test | existing | yes |
| B13 | qualify receiver methods in the factory-site inventory | same test | existing | yes |
| B14 | skip AST nodes that are not calls | same test | existing | yes |
| B15 | count direct `verifyBrokerFactory` calls | same test | existing | yes |
| B16 | inspect every discovered factory-build site | same test | existing | yes |
| B17 | reject an unargued or blank factory-build exemption | same test | existing | yes |
| B18 | inspect every allowlisted factory-build site | same test | existing | yes |
| B19 | reject a stale allowlist exemption with no source call | same test | existing | yes |
| B20 | scan declarations for `runConsole` | same test | existing | yes |
| B21 | skip declarations other than `runConsole` | same test | existing | yes |
| B22 | distinguish assignments from seam calls while inspecting `runConsole` | same test | existing | yes |
| B23 | inspect assignment statements for shared resolver construction | same test | existing | yes |
| B24 | inspect every assignment RHS | same test | existing | yes |
| B25 | skip assignment RHS values that are not calls | same test | existing | yes |
| B26 | accept only aligned `newConsoleBroker` assignments | same test | existing | yes |
| B27 | record the resolver holder when LHS is an identifier | same test | existing | yes |
| B28 | inspect call expressions for read-seam wiring | same test | existing | yes |
| B29 | skip calls without an identifier callee | same test | existing | yes |
| B30 | compare each call against every declared read seam | same test | market schedule added | yes |
| B31 | read each seam's configured shared argument, including calendar argument 1 | same test | RED: two-argument calendar seam was reported unwired | yes |
| B32 | require exactly one shared resolver construction | same test | existing | yes |
| B33 | validate every declared seam after AST collection | same test | market schedule added | yes |
| B34 | reject a declared seam not found in `runConsole` | same test | RED exposed calendar guard gap | yes |
| B35 | reject a seam wired to a value other than the shared holder | same test | calendar second argument previously unrecognized | yes |
