## 1. Evidence and RED

- [x] 1.1 Capture CodeGraph definition/caller/impact evidence for config writes, settings handlers, `runConsole`, and `startEngine`
- [x] 1.2 Create Function Logic Maps and Branch Test Maps before changing existing high-risk functions
- [x] 1.3 Add RED config tests for default/missing false, surgical autostart save, invalid JSON refusal, and concurrent key preservation
- [x] 1.4 Add RED console tests for authenticated ON/OFF, one start call, refusal display, CSRF refusal, and no implicit stop
- [x] 1.5 Add RED startup tests for ON/OFF/read-error/existing-engine outcomes without spawning real processes

## 2. Config and Console

- [x] 2.1 Add `engine.autostart` to Go config types, defaults, and JSON schema
- [x] 2.2 Implement locked surgical autostart load/save and audit recording
- [x] 2.3 Add the closed `AutostartSettings` console seam, settings view, POST route, and explanatory UI
- [x] 2.4 On successful ON save call the existing `StartEngine` seam once and surface its exact result; OFF never calls stop
- [x] 2.5 Update route allowlists/static guards and prove no order or credential route was added

## 3. Startup and Deployment

- [x] 3.1 Add a testable fail-closed console-start autostart helper that reuses `startEngine`
- [x] 3.2 Wire startup outcome into the engine note without making console availability depend on engine success
- [x] 3.3 Set `TOSSOS_DATA_DIR=/var/lib/tossos/data` in Compose and document exact host data mapping
- [x] 3.4 Update deployment/rollback instructions for Docker boot restart, local HTTPS, and later VPN address conversion

## 4. Verification and Deployment

- [x] 4.1 Run focused RED/GREEN/REFACTOR tests, race tests, full `make test`, and `make vet`
- [x] 4.2 Validate all OpenSpec specs and complete Function Logic Map, SDD sync/check, and `make gate CHANGE=enable-engine-autostart-menu`
- [x] 4.3 Perform independent code/security review and resolve high-severity findings
- [x] 4.4 Confirm autostart remains OFF, replace the old HTTP console with the hardened Docker HTTPS console, and verify health/login without starting the engine
- [x] 4.5 Verify Docker and container restart policies are enabled at boot, record deployment evidence, and document the VPN-only remote-access blocker
