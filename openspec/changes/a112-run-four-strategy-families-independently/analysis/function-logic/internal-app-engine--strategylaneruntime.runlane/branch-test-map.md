# Branch Test Map: `strategyLaneRuntime.runLane`

- Source SHA-256: `4a7fd7fedb3237720070c6c4c6ef03030fa30c67e181fdb0a86053a8418390a6`; AST branch locations are authoritative.
- Revision: **modified (a112 7.3, 2026-10-01).** 분기 불변(2). 관측 구성자가 desired/effective(활성화 위임) · 입력 digest 둘을 싣고, 열린 사이클의 거절 코드(`Cycle.Refusal`)를 기록하며(판정 (A)), `lane.Offer()` 를 구성자 밖 다음 줄로 옮겼다 — 레인 호출 순서 · 횟수 불변(Offer 한 번, 투입 시 RunBounded 한 번).
- 편집 전 번들: `analysis/measurements/lot-7.3/pre-edit/internal-app-engine--strategylaneruntime.runlane/`. 변이 원장 `analysis/measurements/lot-7.3/mutation-7.3.tsv`.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 268:2 — 투입 거절(DISABLED · FULL) → 건강만 싣고 반환 | `a112_lane_coordinator_projection_test.go` `TestEightLatchedLanesAreProjectedInProductionOrderWithTheirFirstFailure`(DISABLED) | no — 이 로트가 바꾸지 않음 | yes |
| B2 | if at 281:2 — 유계 사이클 오류 → 실패 문장 | 진입 0 — 생산 Step 은 오류를 내지 않는다(편집 전 번들 기록 그대로) | no — 이 로트가 바꾸지 않음 | n/a |
