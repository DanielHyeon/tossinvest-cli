# Branch Test Map: `consoleOpenAPISeam.Check`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | file mode checks pending generation | `TestConsoleOpenAPIPendingGenerationBlocksPreflightUntilSetupCompletes` | yes | passed |
| B2 | marker read failure remains stopped | `TestConsoleOpenAPIMarkerFailuresRemainFailClosed/marker_read` | yes | passed |
| B3 | pending marker blocks ordinary restart | `TestConsoleOpenAPIPendingGenerationBlocksPreflightUntilSetupCompletes` | yes | passed |
| B4 | credential read failure is unavailable | `TestConsoleOpenAPIPreflightClassifiesMissingReadyAndTransient` | existing | passed |
| B5 | credential source is missing | `TestConsoleOpenAPIPreflightClassifiesMissingReadyAndTransient` | existing | passed |
| B6 | environment mode allocates isolated validation cache | `TestConsoleOpenAPIEnvironmentPreflightUsesFreshTokenAndInvalidatesNormalCache` | yes | passed |
| B7 | isolated allocation failure stays stopped | injected dependency branch coverage | existing | passed |
| B8 | isolated cache is cleaned | `TestConsoleOpenAPIEnvironmentPreflightUsesFreshTokenAndInvalidatesNormalCache` | yes | passed |
| B9 | isolated cleanup failure stays stopped | injected dependency branch coverage | existing | passed |
| B10 | official validation execution failure stays stopped | transient preflight tests | existing | passed |
| B11 | ready environment pair invalidates normal cache | `TestConsoleOpenAPIEnvironmentPreflightUsesFreshTokenAndInvalidatesNormalCache` | yes | passed |
| B12 | normal-cache invalidation failure stays stopped | injected dependency branch coverage | existing | passed |
