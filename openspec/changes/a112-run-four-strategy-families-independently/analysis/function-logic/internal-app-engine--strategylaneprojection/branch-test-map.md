# Branch Test Map: `strategyLaneProjection`

- Source SHA-256: `04dcd7ed4366c18c5ac8b8c0cc490b944f5287dee4db481c35cfec07e6173d70`; AST branch locations are authoritative.
- Revision: **modified (a112 7.5, 2026-10-01).** 분기 불변(5). 레인 상태를 접근자별 잠금 아홉 번 대신 `lane.Status()` 한 번(한 잠금)으로 읽는다 — 찢긴 행 제거(D2).
- 편집 전 번들: `analysis/measurements/lot-7.5/pre-edit/internal-app-engine--strategylaneprojection/`. 변이 원장 `analysis/measurements/lot-7.5/mutation-7.5.tsv`.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 61:2 — 미관측 레인 → 관측 사실 없이 반환 | `a112_lane_coordinator_projection_test.go` `TestAProcessWithoutLanesProjectsTheEightUnobservedDefaults` 계열 | no — 이 로트가 바꾸지 않음 | yes |
| B2 | if at 68:2 — 투입이 들어간 물결 → 시작 | `a112_lane_coordinator_projection_test.go` `TestTheCycleGenerationIsTheMarketWaveInWhichTheLaneWasLastObserved` | no — 이 로트가 바꾸지 않음 | yes |
| B3 | if at 71:3 — 연 사이클 → 결과 · 비정상 | `a112_lane_coordinator_projection_test.go` `TestTheCycleGenerationIsTheMarketWaveInWhichTheLaneWasLastObserved` | no — 이 로트가 바꾸지 않음 | yes |
| B4 | if at 73:4 — 결과 있음 | `a112_lane_coordinator_projection_test.go` `TestTheCycleGenerationIsTheMarketWaveInWhichTheLaneWasLastObserved` | no — 이 로트가 바꾸지 않음 | yes |
| B5 | if at 79:4 — REFUSED 결과만 거절 코드 | `a112_lane_coordinator_projection_test.go` `TestTheLaneDesiredAndEffectiveAreTheActivationTheWaveRanWith` | no — 이 로트가 바꾸지 않음 | yes |
