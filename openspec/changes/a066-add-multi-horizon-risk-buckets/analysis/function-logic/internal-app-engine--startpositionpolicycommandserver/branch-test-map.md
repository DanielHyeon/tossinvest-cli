# Branch Test Map: `StartPositionPolicyCommandServer`

Post-edit, from `ast.json` (50–172, 17 branches). B1–B15 are unchanged by a066 and keep the tests of their owners: the position-policy transport tests, a079, a109 and a114. The table names the existing tests for those rows and the new tests for this lot's rows.

| Branch | Line | Test | Kind |
|---|---|---|---|
| B1 | 52 | commands == nil — `position_policy_transport_test.go` and the a109/a114 endpoint suites (existing; unchanged by a066) | behavioural (owners' suites) |
| B2 | 56 | engine directory blank — `position_policy_transport_test.go` and the a109/a114 endpoint suites (existing; unchanged by a066) | behavioural (owners' suites) |
| B3 | 59 | engine directory not private — `position_policy_transport_test.go` and the a109/a114 endpoint suites (existing; unchanged by a066) | behavioural (owners' suites) |
| B4 | 64 | control directory create error — `position_policy_transport_test.go` and the a109/a114 endpoint suites (existing; unchanged by a066) | behavioural (owners' suites) |
| B5 | 68 | control directory created (else) — `position_policy_transport_test.go` and the a109/a114 endpoint suites (existing; unchanged by a066) | behavioural (owners' suites) |
| B6 | 65 | control directory exists (not an error) — `position_policy_transport_test.go` and the a109/a114 endpoint suites (existing; unchanged by a066) | behavioural (owners' suites) |
| B7 | 71 | control directory validation fails — `position_policy_transport_test.go` and the a109/a114 endpoint suites (existing; unchanged by a066) | behavioural (owners' suites) |
| B8 | 72 | created by this call → remove — `position_policy_transport_test.go` and the a109/a114 endpoint suites (existing; unchanged by a066) | behavioural (owners' suites) |
| B9 | 78 | cleanup helper: created by this call — `position_policy_transport_test.go` and the a109/a114 endpoint suites (existing; unchanged by a066) | behavioural (owners' suites) |
| B10 | 107 | loopback listen fails — `position_policy_transport_test.go` and the a109/a114 endpoint suites (existing; unchanged by a066) | behavioural (owners' suites) |
| B11 | 112 | token generation fails — `position_policy_transport_test.go` and the a109/a114 endpoint suites (existing; unchanged by a066) | behavioural (owners' suites) |
| B12 | 123 | /v1/health non-GET — `position_policy_transport_test.go` and the a109/a114 endpoint suites (existing; unchanged by a066) | behavioural (owners' suites) |
| B13 | 130 | /v1/positions non-GET — `position_policy_transport_test.go` and the a109/a114 endpoint suites (existing; unchanged by a066) | behavioural (owners' suites) |
| B14 | 135 | List error — `position_policy_transport_test.go` and the a109/a114 endpoint suites (existing; unchanged by a066) | behavioural (owners' suites) |
| B15 | 148 | `a079_quarantine_transport_test.go` (existing) | behavioural |
| B16 | 153 | taken: `TestA066EntryLockReleaseThroughTheEngineEndpoint`, `TestA066LatchReleaseCarriesTheBindingIntoTheEngine`, `TestEngineEntryLockReleaseGoesThroughTheRunningEngine` (CLI, production deps); not taken: `TestA066AnEngineWithoutTheCapabilityOffersNoRelease` | behavioural |
| B17 | 163 | a109 descriptor publication tests (existing; was B16) | behavioural |

GREEN: `go test ./internal/app/engine -run TestA066` 7 PASS, `./cmd/tossctl` relaxation tests 6 PASS (2026-09-29).
