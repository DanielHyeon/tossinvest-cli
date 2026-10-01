# Branch Test Map: `strategyProjectionFromAssembly`

- Source SHA-256: `4302edefe72942bd1f4f7f4aa51b3c03e26ef97c13c3d8d52a0b6be94a2eef9f`; AST branch locations are authoritative.
- Revision: **modified (a112 7.3, 2026-10-01).** 분기 구조 불변(10). 순회가 index 를 받고, 순회 머리에 조정자 자식 대입 한 줄(`snapshot.Coordinators[index] = strategyCoordinatorProjection(...)`)을 더했다 — 시장 레코드 갈래보다 앞이라 실패 갈래(B2)에서도 조정자가 보인다.
- 편집 전 번들: `analysis/measurements/lot-7.3/pre-edit/internal-app-engine--strategyprojectionfromassembly/`. 변이 원장 `analysis/measurements/lot-7.3/mutation-7.3.tsv`.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | range at 109:2 — KR · US 순회 — **(편집)** 머리에서 조정자 자식을 채움(R4: 승인 범위 전부) | `a112_lane_coordinator_projection_test.go` `TestTheCoordinatorChildShowsEveryAdmittedOwnerScope` | yes — 편집 전 컴파일 실패, 변이 P07 · P09 CAUGHT(`analysis/measurements/lot-7.3/mutation-7.3.tsv`) | yes |
| B2 | if at 119:3 — worker 미승격 → 시장 실패(조정자는 이미 채워짐) | `a112_lane_coordinator_projection_test.go` `TestTheCoordinatorChildShowsEveryAdmittedOwnerScope`(두 worker 미승격 — 조정자가 보임) | yes — 변이 P09(조정자 대입 제거) CAUGHT | yes |
| B3 | switch at 121:4 — 실패 사유 고르기 | 분기 불변 — 편집 전 번들 서술(`pre-edit/`) 그대로 | no — 이 로트가 바꾸지 않음 | yes |
| B4 | case at 122:4 — 활성화 부재 | 분기 불변 — 편집 전 번들 서술(`pre-edit/`) 그대로 | no — 이 로트가 바꾸지 않음 | yes |
| B5 | case at 124:4 — 근거 stale | 분기 불변 — 편집 전 번들 서술(`pre-edit/`) 그대로 | no — 이 로트가 바꾸지 않음 | yes |
| B6 | case at 126:4 — 보호 미배선 | 분기 불변 — 편집 전 번들 서술(`pre-edit/`) 그대로 | no — 이 로트가 바꾸지 않음 | yes |
| B7 | range at 139:3 — 주문 경로와 같은 handoff 목록 순회 | `a112_owner_scope_trading_test.go` `TestAnActivatedTwoScopeMarketPromotesItsWorkerPerScope` | no — 이 로트가 바꾸지 않음 | yes |
| B8 | if at 140:4 — 조정자 순서의 첫 승인 · 유효 범위(시장 레코드 — 불변) | `a112_owner_scope_trading_test.go` `TestAnActivatedTwoScopeMarketPromotesItsWorkerPerScope` | no — 이 로트가 바꾸지 않음 | yes |
| B9 | if at 145:3 — 승인 범위 없음 → EvidenceStale | 진입 0 — 편집 전에도 진입 0(5.2.2.2 번들 기록 그대로) | no — 이 로트가 바꾸지 않음 | n/a |
| B10 | if at 153:3 — 레인 근거 digest 부재 → 후보 근거 | 분기 불변 — 편집 전 번들 서술(`pre-edit/`) 그대로 | no — 이 로트가 바꾸지 않음 | yes |
