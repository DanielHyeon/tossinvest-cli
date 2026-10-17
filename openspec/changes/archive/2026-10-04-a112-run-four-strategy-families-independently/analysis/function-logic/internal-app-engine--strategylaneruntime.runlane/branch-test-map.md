# Branch Test Map: `strategyLaneRuntime.runLane`

- Source SHA-256: `0935960f05276aa2bf972734943b879a5c56b3525f2083aeed9f5b2e3855c620`; AST branch locations are authoritative.
- Revision: **modified (a112 7.5, 2026-10-01).** 분기 불변(2). `RunBounded` 에 넘기는 값이 `strategyFamilyLaneStep(lane, promotion)` → `runtime.laneStepFor(lane, promotion)` 한 곳 — 생산 정의(`strategy_lane_step.go`, `!tossos_testseams`)는 strategyFamilyLaneStep 한 줄이고 seam 은 태그 빌드(`strategy_lane_step_testseam.go`)에만 있다. 첫 구현은 무태그 함수 필드였고 5.1.2.1 핀이 잡았다(「핀이 잡은 자기 이탈」 — review).
- 편집 전 번들: `analysis/measurements/lot-7.5/pre-edit/internal-app-engine--strategylaneruntime.runlane/`. 변이 원장 `analysis/measurements/lot-7.5/mutation-7.5.tsv`.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 313:2 — 투입 거절(DISABLED · FULL) → 건강만 싣고 반환 | `a112_lane_coordinator_projection_test.go` `TestEightLatchedLanesAreProjectedInProductionOrderWithTheirFirstFailure` | no — 이 로트가 바꾸지 않음 | yes |
| B2 | if at 329:2 — 유계 사이클 오류 → 실패 문장 | `a112_lane_latency_testseam_test.go` `TestAHungLaneDoesNotDelayItsPeersInTheSameWave`(멈춘 레인의 마감 시한 오류) | no — 분기 불변(편집 전 진입 0 → 이 로트의 시험이 처음 진입) | yes |
