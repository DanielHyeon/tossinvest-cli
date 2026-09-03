# Branch Test Map: `FamilyWorker.Run`

- Source: `internal/strategyworker/worker.go` (193-203); file SHA-256 `588c56f0a3100d0f1fa93e8fed4cf303b506ad88823c53e8f8690ebe84504335`. AST branch positions are authoritative.
- Rows carry measured counts from Go coverage profiles, count mode.
- Per-test attribution set: 두 바이너리의 테스트 **전체**(태그 72 · 무태그 54).

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 194:2 | arm entered 9x (strategyworker tagged suite); arm entered 1x (strategyworker untagged suite); `TestALatchedLaneReportsTheLatchRatherThanDormancy`, `TestEveryProductionWorkerIsBornDormantAndEmitsNothing` |
| B2 | if | 197:2 | arm entered 4x (strategyworker tagged suite); arm not entered (strategyworker untagged suite); `TestAWorkerOfTheOtherMarketRefuses`, `TestAWorkerRefusesAProposalFromAnotherLane`, `TestAWorkerRefusesAProposalWhoseSealNoLongerHolds`, `TestAWorkerRefusesWhenTheFamilyCannotBeDerivedFromTheSealedAuthority` |

RED→GREEN: 이 함수의 새 계약(승격이 인자다)은 **반증으로** 잼. `worker.Effective(activation)`
검사를 지운 변이(E12, review.md 의 2026-09-03 절)는 태그·무태그 양쪽에서 CAUGHT.
`TestASignedFourFamilyActivationPromotesExactlyTheLanesItNames` 와
`TestAPromotedLaneAdmitsItsFamilyWhileAnUnpromotedOneStopsIt` 가 두 상태를 함께 세운다.

A row states what was measured, not what is intended. An arm recorded as not entered is a coverage gap, not a pass.
