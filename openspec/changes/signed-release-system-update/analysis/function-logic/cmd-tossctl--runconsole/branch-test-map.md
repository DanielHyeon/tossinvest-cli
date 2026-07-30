# Branch Test Map: `runConsole`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | Nil Cobra context is replaced with a cancellable background context | existing console startup tests | baseline | pass |
| B2 | KR verification-record resolution fails | existing console path tests | baseline | pass |
| B3 | US verification-record resolution fails | existing console path tests | baseline | pass |
| B4 | Soak-record resolution fails | existing console path tests | baseline | pass |
| B5 | Attestation-path resolution fails | existing console path tests | baseline | pass |
| B6 | Journal path failure degrades only the journal dashboard | existing unwired-journal test | baseline | pass |
| B7 | Engine directory resolves and binds the marker/lock paths | `TestRunConsoleWiresFixedSystemUpdaterAndBothActivityGuards` | baseline | pass |
| B8 | Engine directory failure leaves engine status and update exclusion unwired | existing unwired-engine test | baseline | pass |
| B9 | Self executable cannot be resolved, so all update capabilities remain nil | existing system-update wiring failure test | baseline | pass |
| B10 | Self executable resolves and update assembly begins | `TestAssembleConsoleSystemUpdateUsesRealTossctlBinary` | no signed assembler | pass |
| B11 | Sigstore cache path fails; fixed local inspect/install is retained when constructible | degraded assembly test | download and local updater coupled | pass |
| B12 | Sigstore cache resolves; production assembler is used | `TestAssembleConsoleSystemUpdateUsesRealTossctlBinary` | no production assembly seam | pass |
| B13 | Cache failure plus local updater failure disables the update section | existing updater-construction failure test | baseline | pass |
| B14 | Cache failure plus valid local updater preserves candidate review/install | degraded assembly test | local fallback absent | pass |
| B15 | Production assembler returned an updater and it is bound as inspector/stager | real CLI assembly regression | stager absent | pass |
| B16 | Verifier/downloader construction failed; local candidate operations remain | degraded case in `TestAssembleConsoleSystemUpdateUsesRealTossctlBinary` | all update functions disabled | pass |
| B17 | Verifier/downloader construction succeeded and download is enabled | real CLI assembly regression | download absent | pass |
| B18 | Resolved engine directory installs the real cross-process flock callback | `TestRunConsoleWiresFixedSystemUpdaterAndBothActivityGuards` | no real exclusion | pass |
| B19 | Flock contention/error propagates from the callback | updater/engine exclusion tests | update could overlap engine | pass |
