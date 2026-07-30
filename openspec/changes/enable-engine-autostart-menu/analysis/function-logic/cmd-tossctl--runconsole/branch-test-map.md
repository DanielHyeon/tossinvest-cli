# Branch Test Map: `runConsole`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | nil command context receives a background context | existing console command tests | yes | yes |
| B2 | remote option validation fails closed | `TestRemoteAccessTokenFileMustBePrivateAndLong` | yes | yes |
| B3 | KR verify record path fails | existing console path tests | yes | yes |
| B4 | US verify record path fails | existing console path tests | yes | yes |
| B5 | soak record path fails | existing console path tests | yes | yes |
| B6 | attestation path fails | existing console path tests | yes | yes |
| B7 | journal path is unavailable but console remains inspectable | existing journal-path degradation test | yes | yes |
| B8 | engine journal directory resolves | existing engine marker wiring test | yes | yes |
| B9 | engine journal directory fails and status degrades visibly | existing engine marker wiring test | yes | yes |
| B10 | container disables in-place binary update | hardened-container smoke test | yes | yes |
| B11 | native binary path resolution fails | existing system-update assembly tests | yes | yes |
| B12 | native binary path resolution error branch | existing system-update assembly tests | yes | yes |
| B13 | native binary path resolves | existing system-update assembly tests | yes | yes |
| B14 | update cache path fails | existing system-update assembly tests | yes | yes |
| B15 | update cache path resolves | existing system-update assembly tests | yes | yes |
| B16 | fallback local updater construction fails | existing system-update assembly tests | yes | yes |
| B17 | fallback local updater is wired | existing system-update assembly tests | yes | yes |
| B18 | signed updater exists | existing system-update assembly tests | yes | yes |
| B19 | signed downloader construction fails | existing system-update assembly tests | yes | yes |
| B20 | signed downloader is wired | existing system-update assembly tests | yes | yes |
| B21 | engine directory enables the update/engine exclusion lock | existing update serialization tests | yes | yes |
| B22 | exclusion lock acquisition failure propagates | existing update serialization tests | yes | yes |
| B23 | engine autostart settings seam is present | `TestConfiguredEngineAutostartOffDoesNotStart`, seam audit test | yes | yes |
| B24 | non-empty startup outcome is surfaced | autostart read-failure/refusal/start tests | yes | yes |
