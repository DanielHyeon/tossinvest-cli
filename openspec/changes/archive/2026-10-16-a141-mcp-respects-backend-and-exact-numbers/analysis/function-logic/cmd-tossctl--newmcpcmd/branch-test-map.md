# Branch Test Map: `newMCPCmd`

Final AST branch IDs. A passed package suite is not proof that every unchanged error branch was taken. Focused receipts are in ../../mutation-receipt.md; whole-suite receipts are separate.

| Branch | Scenario | Test / evidence | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | config load error | unchanged; static AST comparison | not claimed | no focused branch-hit claim |
| B2 | invalid backend before construction | unchanged resolver behavior, moved earlier; AST | not claimed | no focused branch-hit claim |
| B3 | path resolution error | unchanged; AST | not claimed | no focused branch-hit claim |
| B4 | credential load error | unchanged; AST | not claimed | no focused branch-hit claim |
| B5 | disabled/WTS deny, enabled auto/openapi allow | TestMCPCmdWTSDoesNotWireOfficialDependencies; TestMCPCmdDisabledAndEnabledBackendWiring; TestMCPOfficialEligibilityMatchesCLI | enabled/WTS mutants detected | Terra focused PASS |
| B6 | unexpected session read error | unchanged; AST | not claimed | no focused branch-hit claim |
| B7 | synthetic WTS session retained | guarded command tests; TestCallOperationWTSDispatches separately verifies local-fixture operation dispatch | WTS gate mutant detected at auth status | Terra focused PASS |
| B8 | neither credential capability | unchanged; AST | not claimed | no focused branch-hit claim |
| B9 | sessionless hybrid embedding | unchanged; AST | not claimed | no focused branch-hit claim |
| B10 | update checker invoked | guarded command subprocess denies external transport | not claimed | Terra guarded PASS |
| B11 | newer-version instruction append | unchanged; AST | not claimed | no focused branch-hit claim |
