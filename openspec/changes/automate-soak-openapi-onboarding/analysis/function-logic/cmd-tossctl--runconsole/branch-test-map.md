# Branch Test Map: `runConsole`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | nil context | existing command construction tests | existing | pending |
| B2 | invalid remote options | existing remote option tests | existing | pending |
| B3 | KR verify record resolution error | existing resolver tests | existing | pending |
| B4 | US verify record resolution error | existing resolver tests | existing | pending |
| B5 | soak record resolution error | existing resolver tests | existing | pending |
| B6 | attestation resolution error | existing resolver tests | existing | pending |
| B7 | journal path unavailable | existing console path tests | existing | pending |
| B8 | engine directory available | existing console path tests | existing | pending |
| B9 | engine directory unavailable | existing console path tests | existing | pending |
| B10 | container disables updater | existing update tests | existing | pending |
| B11 | local self-path branch | existing update tests | existing | pending |
| B12 | local self-path error | existing update tests | existing | pending |
| B13 | local updater assembly | existing update tests | existing | pending |
| B14 | update cache path error | existing update tests | existing | pending |
| B15 | update cache path ready | existing update tests | existing | pending |
| B16 | fallback updater construction error | existing update tests | existing | pending |
| B17 | fallback updater ready | existing update tests | existing | pending |
| B18 | assembled updater non-nil | existing update tests | existing | pending |
| B19 | downloader assembly error | existing update tests | existing | pending |
| B20 | downloader assembly success | existing update tests | existing | pending |
| B21 | engine directory wires lock | existing update lock tests | existing | pending |
| B22 | engine lock acquisition error | existing update lock tests | existing | pending |
| B23 | engine autostart seam present | existing autostart tests | existing | pending |
| B24 | engine autostart note nonblank | existing autostart tests | existing | pending |
| B25 | engine autostart note nonblank after Open API seam assembly | existing autostart tests | existing | pending |
| new preflight | source + official result maps to console state | `TestConsoleOpenAPIPreflightClassification` | pending | pending |
| new setup | isolated validation then save/cache/audit | `TestConsoleOpenAPISetupOrderAndIsolation` | pending | pending |
| new environment | env rejection cannot save file | `TestConsoleOpenAPIEnvironmentCredentialsAreNotWebReplaceable` | pending | pending |
