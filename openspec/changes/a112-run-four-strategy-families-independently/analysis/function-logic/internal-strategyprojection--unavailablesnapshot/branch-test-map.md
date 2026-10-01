# Branch Test Map: `UnavailableSnapshot`

- Source SHA-256: `f192e4f2f934f8bb3e165a47f2ecb7dda91874826c299aff0aa073f8082cbd01`; AST branch locations are authoritative.
- Revision: **modified (a112 7.3, 2026-10-01).** 분기 없음. DormantSnapshot 과 같은 기본 자식 둘을 더했다.
- 편집 전 번들: `analysis/measurements/lot-7.3/pre-edit/internal-strategyprojection--unavailablesnapshot/`. 변이 원장 `analysis/measurements/lot-7.3/mutation-7.3.tsv`.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | happy path (분기 없음) — 엔진에 닿지 못한 스냅숏도 여덟 · 둘을 미관측으로 싣는다(건강 추론 없음) | `a112_lane_coordinator_children_test.go` `TestDormantAndUnavailableSnapshotsCarryEightUnobservedLanesAndTwoCoordinators` | no — 시험 코드 | yes |
