# Branch Test Map: `DormantSnapshot`

- Source SHA-256: `f192e4f2f934f8bb3e165a47f2ecb7dda91874826c299aff0aa073f8082cbd01`; AST branch locations are authoritative.
- Revision: **modified (a112 7.3, 2026-10-01).** 분기 없음. 구성자에 기본 자식 둘(`defaultLanes()` · `defaultCoordinators()`)을 더했다.
- 편집 전 번들: `analysis/measurements/lot-7.3/pre-edit/internal-strategyprojection--dormantsnapshot/`. 변이 원장 `analysis/measurements/lot-7.3/mutation-7.3.tsv`.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | happy path (분기 없음) — 여덟 미관측 레인(골든 순서 · OFF/OFF/UNOBSERVED) · 두 미관측 조정자 | `a112_lane_coordinator_children_test.go` `TestDormantAndUnavailableSnapshotsCarryEightUnobservedLanesAndTwoCoordinators` · `TestTheDefaultLaneTableIsTheFrozenGoldenDescriptorList` | no — 시험 코드 | yes |
