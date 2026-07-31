# Branch Test Map: `consoleOpenAPISeam.Save`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | environment/blank input rejected | environment replacement and handler tests | existing | pending |
| B2 | replacement validation is isolated | `TestConsoleOpenAPISetupUsesIsolatedTokenThenSavesInvalidatesAndAudits` | existing | pending |
| B3 | marker creation failure prevents persistence | focused marker helper test | pending | pending |
| B4 | persistence error retains marker and does not start | `TestConsoleOpenAPIMarkerFailuresRemainFailClosed/save_error_retains_marker` | yes | passed |
| B5 | cache/audit failure remains blocked after next restart | `TestConsoleOpenAPIPendingGenerationBlocksPreflightUntilSetupCompletes` | yes | pending |
| B6 | retry completes and releases preflight | same pending-generation recovery test | yes | pending |
| B7 | classified validation is not ready | rejected/transient setup tests | existing | pending |
| B8 | pending marker creation fails | `TestConsoleOpenAPIMarkerFailuresRemainFailClosed/marker_create` | yes | passed |
| B9 | save error cannot clear marker | `TestConsoleOpenAPIMarkerFailuresRemainFailClosed/save_error_retains_marker` | yes | passed |
| B10 | normal token invalidation fails | pending-generation recovery test | yes | pending |
| B11 | audit write fails | pending-generation recovery test | yes | pending |
| B12 | final marker clear fails | `TestConsoleOpenAPIMarkerFailuresRemainFailClosed/final_marker_clear` | yes | passed |
