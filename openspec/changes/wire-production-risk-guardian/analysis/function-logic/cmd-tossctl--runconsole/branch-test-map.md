# Branch Test Map: `runConsole`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | nil context falls back to background | existing console command tests | baseline | pass |
| B2 | KR verify path error propagates | existing resolver tests | baseline | pass |
| B3 | US verify path error propagates | existing resolver tests | baseline | pass |
| B4 | soak path error propagates | existing resolver tests | baseline | pass |
| B5 | attestation path error propagates | existing resolver tests | baseline | pass |
| B6 | journal path unavailable is warning/nonfatal | existing console startup tests | baseline | pass |
| B7 | engine directory resolves | `TestRunConsoleWiresFixedSystemUpdaterAndBothActivityGuards` | updater absent | pass |
| B8 | engine directory failure leaves engine seams nil | existing isolated-path tests | baseline | pass |
| B9 | current binary path unavailable disables updater | localupdate/command tests | updater absent | pass |
| B10 | current path available proceeds to updater validation | fixed wiring test | updater absent | pass |
| B11 | invalid current executable disables updater | localupdate refusal tests | updater absent | pass |
| B12 | valid current executable injects fixed updater | fixed wiring test | updater absent | pass |
| B13 | engine directory yields lock closure | fixed wiring test | no update flock | pass |
| B14 | `enginelock.Acquire` error propagates to handler | system-update refusal tests | no real exclusion | pass |
