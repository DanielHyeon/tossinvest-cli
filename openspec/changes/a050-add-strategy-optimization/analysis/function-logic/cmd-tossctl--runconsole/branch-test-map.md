# Branch Test Map: `runConsole`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | nil context falls back safely | console command suite | historical behavior | yes |
| B2 | remote options invalid | console remote tests | historical behavior | yes |
| B3 | verify record invalid | console command suite | historical behavior | yes |
| B4 | US verify record invalid | console command suite | historical behavior | yes |
| B5 | soak record invalid | console command suite | historical behavior | yes |
| B6 | attestation path invalid | console command suite | historical behavior | yes |
| B7 | OpenAPI seam invalid | console OpenAPI tests | historical behavior | yes |
| B8 | journal path unavailable | console journal tests | historical behavior | yes |
| B9 | journal path available | `TestConsoleOptimizationCommanderUsesSeparatePrivateControlStore` | new seam absent | yes |
| B10 | optimization store open fails | fail-closed warning contract | new seam absent | yes |
| B11 | optimization store opens | `TestConsoleOptimizationCommanderUsesSeparatePrivateControlStore` | new seam absent | yes |
| I1 | performance DB opens and one read source is wired to both console dashboard and optimization evidence | `TestConsolePerformanceCapabilitiesOpenOneProfileDatabaseForBothReadSeams`, `TestRunConsoleWiresAndClosesPerformanceWithoutJournalCollection` | `Options.Performance` absent and evidence hard-coded nil | yes |
| I2 | performance DB open fails with no partial read capability | `TestConsolePerformanceCapabilitiesFailWithoutPartialReadAuthority` | no production performance open existed | yes |
| I3 | opened performance DB is closed with the console lifecycle | `TestConsolePerformanceCapabilitiesOpenOneProfileDatabaseForBothReadSeams`, `TestRunConsoleWiresAndClosesPerformanceWithoutJournalCollection` | no production performance handle existed | yes |
| B35 | position-policy descriptor stat failure remains a warning/read-only fallback | existing position-policy wiring tests | historical behavior | yes |
| B36 | console server/shutdown result preserves existing error precedence | existing finish-console tests | historical behavior | yes |
| B12 | engine marker resolves | engine runtime tests | historical behavior | yes |
| B13 | engine marker unavailable | engine runtime tests | historical behavior | yes |
| B14 | container disables self-update | system-update wiring tests | historical behavior | yes |
| B15 | native self path unavailable | system-update wiring tests | historical behavior | yes |
| B16 | native self path available | `TestRunConsoleWiresFixedSystemUpdaterAndBothActivityGuards` | historical behavior | yes |
| B17 | cache path unavailable | system-update wiring tests | historical behavior | yes |
| B18 | local updater construction fails | system-update wiring tests | historical behavior | yes |
| B19 | signed updater path available | `TestRunConsoleWiresFixedSystemUpdaterAndBothActivityGuards` | historical behavior | yes |
| B20 | local updater succeeds | system-update wiring tests | historical behavior | yes |
| B21 | local updater fails | system-update wiring tests | historical behavior | yes |
| B22 | assembled updater exists | `TestAssembleConsoleSystemUpdateUsesRealTossctlBinary` | historical behavior | yes |
| B23 | downloader assembly fails | system-update wiring tests | historical behavior | yes |
| B24 | downloader assembly succeeds | system-update wiring tests | historical behavior | yes |
| B25 | engine directory enables lock seam | system-update wiring tests | historical behavior | yes |
| B26 | lock acquisition fails | system-update wiring tests | historical behavior | yes |
| B27 | engine boot seam exists | engine autostart tests | historical behavior | yes |
| B28 | autostart produces operator note | engine autostart tests | historical behavior | yes |
| B29 | engine directory enables policy RPC lookup | position-policy wiring tests | historical behavior | yes |
| B30 | policy RPC descriptor exists | position-policy wiring tests | historical behavior | yes |
| B31 | descriptor missing benignly | position-policy wiring tests | historical behavior | yes |
| B32 | policy RPC dial fails | position-policy wiring tests | historical behavior | yes |
| B33 | policy RPC dial succeeds | position-policy wiring tests | historical behavior | yes |
| B34 | descriptor stat fails unexpectedly | position-policy wiring tests | historical behavior | yes |
