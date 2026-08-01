# Branch Test Map: `runConsole`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | nil command context falls back to background context | existing console command tests | existing | yes |
| B2 | remote access option validation fails before server assembly | existing remote option tests | existing | yes |
| B3 | KR verify record resolution failure returns | existing resolver characterization | existing | yes |
| B4 | US verify record resolution failure returns | existing resolver characterization | existing | yes |
| B5 | soak record resolution failure returns | existing resolver characterization | existing | yes |
| B6 | attestation resolution failure returns | existing resolver characterization | existing | yes |
| B7 | OpenAPI seam construction failure returns | existing OpenAPI console tests | existing | yes |
| B8 | journal path failure degrades the panel without aborting | existing console journal tests | existing | yes |
| B9 | engine directory resolution succeeds and wires marker path | existing engine marker tests | existing | yes |
| B10 | engine directory resolution failure leaves status unwired | existing engine marker tests | existing | yes |
| B11 | container mode disables executable update seams | existing system-update wiring tests | existing | yes |
| B12 | executable path resolution failure leaves updater unwired | existing system-update wiring tests | existing | yes |
| B13 | update cache resolution fails but fixed-sibling updater may remain | existing system-update wiring tests | existing | yes |
| B14 | updater construction failure reports and leaves install unwired | existing system-update wiring tests | existing | yes |
| B15 | updater construction succeeds after cache failure | existing system-update wiring tests | existing | yes |
| B16 | normal assembly exposes fixed sibling updater/downloader only when verified | existing system-update wiring tests | existing | yes |
| B17 | partial updater result preserves safe local inspection seam | existing system-update wiring tests | existing | yes |
| B18 | signed release downloader failure is advisory | existing signed release tests | existing | yes |
| B19 | signed release downloader success wires fetch-only seam | existing signed release tests | existing | yes |
| B20 | engine directory wires real update exclusion lock | existing update-lock tests | existing | yes |
| B21 | engine boot seam supplies a load function | existing engine autostart tests | existing | yes |
| B22 | scheduler-independent engine autostart refusal is printed | existing engine autostart tests | existing | yes |
| B23 | container finish path stops the engine with bounded semantics | existing `finishConsole` tests | existing | yes |
| B24 | non-container finish path preserves serve result | existing `finishConsole` tests | existing | yes |
| B25 | engine-stop result preserves original serve error precedence | existing `finishConsole` tests | existing | yes |
| B26 | missing state renders without network; selected KR state reads calendar through shared official client but never writes/activates | `TestConsoleMarketScheduleSeamReadsClosedDefaults`, `TestConsoleMarketScheduleSeamDoesNotActivateApprovedDesiredState`, `TestOnlyConsoleGoReachesTheConsolePackage` | provenance/importer RED | yes |
