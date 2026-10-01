# Branch Test Map: `Lane.Run`

- Source: `internal/strategyworker/lane.go` (211-219); file SHA-256 `b869921551c2058b80fa03d2bc9c601f3ea306234d918ee41cde019312d529ab`. AST branch positions are authoritative.
- Rows carry measured counts from Go coverage profiles, count mode.
- Per-test attribution set: 두 바이너리의 테스트 **전체**(태그 72 · 무태그 54).

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 215:2 | arm entered 4x (strategyworker tagged suite); arm entered 3x (strategyworker untagged suite); `TestALatchedLaneEmitsNothingEvenWhenItIsEffective`, `TestALatchedLaneReportsTheLatchRatherThanDormancy`, `TestAProductionLaneBornFromADurableRecordIsLatched`, `TestARestoredLaneStillCannotUnlatchItself` |

RED→GREEN: 잠금 우선 순서는 **반증으로** 잼. 잠금 검사를 지운 변이(E5)는
`TestALatchedLaneStopsItsFamilyAndItsPeersKeepTrading` 과 위 네 시험에서 CAUGHT.

A row states what was measured, not what is intended. An arm recorded as not entered is a coverage gap, not a pass.
