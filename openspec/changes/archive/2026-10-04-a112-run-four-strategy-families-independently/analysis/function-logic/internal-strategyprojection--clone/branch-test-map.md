# Branch Test Map: `Clone`

- Source SHA-256: `f192e4f2f934f8bb3e165a47f2ecb7dda91874826c299aff0aa073f8082cbd01`; AST branch locations are authoritative.
- Revision: **modified (a112 7.3, 2026-10-01).** 분기 불변(1). 직선 코드: 자식 둘의 깊은 복사(`cloneLanes` · `cloneCoordinators`)를 구성자에 더했다.
- 편집 전 번들: `analysis/measurements/lot-7.3/pre-edit/internal-strategyprojection--clone/`. 변이 원장 `analysis/measurements/lot-7.3/mutation-7.3.tsv`.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | range at 228:2 — 시장 두 개 복사(불변) | `TestCloneCarriesRuntimeIdentityWithoutSharingIt` · `a112_lane_coordinator_children_test.go` `TestCloneDeepCopiesLaneAndCoordinatorChildren` | yes(자식 복사 — 직선) — 편집 전 컴파일 실패, 변이 P12(얕은 복사) CAUGHT(`analysis/measurements/lot-7.3/mutation-7.3.tsv`) | yes |
