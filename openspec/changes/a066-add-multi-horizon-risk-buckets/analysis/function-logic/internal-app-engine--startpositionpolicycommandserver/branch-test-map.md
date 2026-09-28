# Branch Test Map: `StartPositionPolicyCommandServer`

Post-edit, from `ast.json` (50–172, 17 branches). B1–B15 are unchanged by a066 and keep the tests of their owners: the position-policy transport tests, a079, a109 and a114. The table names the existing tests for those rows and the new tests for this lot's rows.

| Branch | Line | Test | Kind |
|---|---|---|---|
| B1–B11 | 52–112 | `position_policy_transport_test.go`, a109 staging/descriptor tests (existing) | behavioural (owners' suites) |
| B12–B14 | 123–135 | `position_policy_transport_test.go` (existing) | behavioural |
| B15 | 148 | `a079_quarantine_transport_test.go` (existing) | behavioural |
| B16 | 153 | taken: `TestA066EntryLockReleaseThroughTheEngineEndpoint`, `TestA066LatchReleaseCarriesTheBindingIntoTheEngine`, `TestEngineEntryLockReleaseGoesThroughTheRunningEngine` (CLI, production deps); not taken: `TestA066AnEngineWithoutTheCapabilityOffersNoRelease` | behavioural |
| B17 | 163 | a109 descriptor publication tests (existing; was B16) | behavioural |

GREEN: `go test ./internal/app/engine -run TestA066` 7 PASS, `./cmd/tossctl` relaxation tests 6 PASS (2026-09-29).
