# Branch Test Map: `TestPollRejectsTrackedReadFromAnotherCanonicalScope`

Source: `internal/filldetect/detect_test.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/filldetect/detect_test.go:475` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B2 | if at `internal/filldetect/detect_test.go:479` | Focused canonical/temporal ownership regressions plus full and race suites. |

Coverage includes external exclusion, identifier reuse, partial-scope rejection, v19 migration, confirmed temporal ownership, reservation recovery, and operator-release refusal/success.
