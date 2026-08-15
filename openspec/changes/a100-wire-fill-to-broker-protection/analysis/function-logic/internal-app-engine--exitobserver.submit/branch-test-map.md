# Branch Test Map: `ExitObserver.submit`

Source: `internal/app/engine/exitloop.go` (1343-1418).

| Branch | Scenario | Test |
|---|---|---|
| B1 | floor error | `TestAFloorThatCannotBeComputedSellsNothing` |
| B2 | zero floor | `TestAZeroFloorSubmitsNothingAndLeavesTheLevelProposable` |
| B3 | issuer refusal | `TestARefusedProposalReleasesTheLevelAndAlerts` |
| B4 | attach failure | `TestTheWholeExitPathEndToEnd` |
| B5 | sell intent failure | `TestARefusedProposalReleasesTheLevelAndAlerts` |
| B6 | outcome switch | `TestTheWholeExitPathEndToEnd` |
| B7 | confirmed | `TestTheWholeExitPathEndToEnd` |
| B8 | in-doubt | `TestAnInDoubtSubmissionKeepsTheProposalArmed` |
| B9 | in-flight | `TestARefusedProposalReleasesTheLevelAndAlerts` |
| B10 | default refusal | `TestARefusedProposalReleasesTheLevelAndAlerts` |
| B11 | detail fallback | `TestARefusedProposalReleasesTheLevelAndAlerts` |
