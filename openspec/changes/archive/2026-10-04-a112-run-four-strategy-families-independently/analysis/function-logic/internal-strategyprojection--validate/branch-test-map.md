# Branch Test Map: `Validate`

- Source SHA-256: `f192e4f2f934f8bb3e165a47f2ecb7dda91874826c299aff0aa073f8082cbd01`; AST branch locations are authoritative.
- Revision: **modified (a112 7.3, 2026-10-01).** 편집 전 5 분기 → 7: 시장 검사 뒤에 B6(lanes[8] 검사) · B7(coordinators[2] 검사)을 더했다. 앞 다섯은 불변.
- 편집 전 번들: `analysis/measurements/lot-7.3/pre-edit/internal-strategyprojection--validate/`. 변이 원장 `analysis/measurements/lot-7.3/mutation-7.3.tsv`.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 271:2 — envelope 3항 | `TestValidationRejectsMissingDuplicateScopeAndInventedReadiness` | no — 이 로트가 바꾸지 않음 | yes |
| B2 | if at 274:2 — runtime identity | `TestValidationRejectsPartialOrNoncanonicalRuntimeIdentity` | no — 이 로트가 바꾸지 않음 | yes |
| B3 | range at 277:2 — KR · US | `TestDormantSnapshotContainsExactPairedHonestMarkets` | no — 이 로트가 바꾸지 않음 | yes |
| B4 | if at 279:3 — 시장 부재/교차 | `TestValidationRejectsMissingDuplicateScopeAndInventedReadiness` | no — 이 로트가 바꾸지 않음 | yes |
| B5 | if at 282:3 — 시장 레코드 판정 | `TestValidationRejectsMissingDuplicateScopeAndInventedReadiness` | no — 이 로트가 바꾸지 않음 | yes |
| B6 | if at 287:2 — **(새)** 레인 자식: 개수 8 · 고정 순서 · 열쇠 · enum · 거절 코드 ⇔ REFUSED(판정 (A)) · 미관측 무사실 · 관측 사슬(물결 ⇔ 트리거, 투입만 시작, 연 사이클만 결과) | `a112_lane_coordinator_children_test.go` `TestValidateRefusesMalformedLaneAndCoordinatorChildren`(레인 26 행 — 거절 코드 짝 4 행 포함) | yes — 편집 전 컴파일 실패, 변이 P10 · P11 · P13 · P25 CAUGHT(`analysis/measurements/lot-7.3/mutation-7.3.tsv`) | yes |
| B7 | if at 290:2 — **(새)** 조정자 자식: 개수 2 · KR,US · 미관측 무사실 · 사유 · 중재 코드 · 정렬 유일 gated · null 아닌 목록 | `a112_lane_coordinator_children_test.go` `TestValidateRefusesMalformedLaneAndCoordinatorChildren`(조정자 16 행 — 선택 범위 계보 2 행 포함) | yes — 편집 전 컴파일 실패, 변이 P20 CAUGHT | yes |
