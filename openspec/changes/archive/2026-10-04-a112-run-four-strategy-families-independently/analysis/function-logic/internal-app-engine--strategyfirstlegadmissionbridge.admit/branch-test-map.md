# Branch Test Map: `admit`

- Source SHA-256: `254b4a6abb0d95febd036b0f437829c391e61000fa71b9b9abac1241ee14444c`; AST branch locations are authoritative.
- Revision: **modified (a112 6.2, 2026-10-01).** a112 6.2(Manager 판정 (A)): Guardian precheck 거절 갈래(B5) 안에 B6(`*execgw.QFinalRefusal` 의 코드가 `BUCKET_CAP_EXHAUSTED` — 타입 · 코드) · B7(범위 키)을 더해, 그 범위의 버킷 고갈만 범위 거절 타입으로 싣는다. 6 분기 → 8, 편집 전 B6(원자 admission 실패)은 B8.
- 편집 전 번들: `analysis/measurements/lot-6.2/pre-edit/internal-app-engine--strategyfirstlegadmissionbridge.admit/`. 변이 원장 `analysis/measurements/lot-6.2/mutation-6.2.tsv`.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 76:2 — 결과 검증 거절 | 분기 불변 — 편집 전 번들(`lot-6.2/pre-edit/`) | no — 이 로트가 바꾸지 않음 | 편집 전 측정 |
| B2 | if at 79:2 — bridge · loader · Guardian 부재 | 분기 불변 — 편집 전 번들(`lot-6.2/pre-edit/`) | no — 이 로트가 바꾸지 않음 | 편집 전 측정 |
| B3 | if at 83:2 — 1차 레그 권한 수집 실패 → cause 운반(5.2.2.2) | 분기 불변 — 편집 전 번들(`lot-6.2/pre-edit/`) | no — 이 로트가 바꾸지 않음 | 편집 전 측정 |
| B4 | if at 88:2 — 권한 불일치 | 분기 불변 — 편집 전 번들(`lot-6.2/pre-edit/`) | no — 이 로트가 바꾸지 않음 | 편집 전 측정 |
| B5 | if at 92:2 — Guardian precheck 거절 → `AuthorityMismatch`(문구 = 거절) | `a112_qfinal_family_test.go` `TestAPrecheckRefusalOtherThanBucketExhaustionStillStopsTheCycle`(SYMBOL_NOT_ALLOWED — 범위 거절 아님) · `TestAnExistingGuardianCapRefusalIsNotAScopeRefusal`(EXISTING_GUARDIAN_CAP — 같은 q_final 타입 다른 코드) | no — 조건 불변 | yes |
| B6 | if at 98:3 — **(새)** 거절이 `*execgw.QFinalRefusal` 이고 코드가 `BUCKET_CAP_EXHAUSTED` → 그 범위의 정책 결과(결함 아님) | `a112_qfinal_family_test.go` `TestAnExhaustedFamilyBucketRefusesOnlyThatFamily`(두 순서 — 조정자 순서에서 continuation 발급) · `TestASharedDimensionExhaustionRefusesEveryScopeAndIssuesNothing`(공유 차원 — 두 범위 각자 거절 · 발급 0) | yes — `analysis/measurements/lot-6.2/red-6.2.log`(편집 전 조정자 순서 발급 0 · 타입 없음) · 변이 U01(분류 제거) · U02(모든 precheck 로 확대) · U03(다른 q_final 코드로 확대) · U05(원인 누락) CAUGHT(`analysis/measurements/lot-6.2/mutation-6.2.tsv`) | yes |
| B7 | if at 99:4 — **(새)** 범위 키 정규화 → 범위 거절 타입 `strategyScopeRefusal{cause: QFinalRefusal}`(정규화 실패면 결함 그대로 — 봉인이 보장해 도달 불가) | `a112_qfinal_family_test.go` `TestAnExhaustedFamilyBucketRefusesOnlyThatFamily` | yes — U01 · U05 | 거짓 갈래 진입 0 |
| B8 | if at 106:2 — 원자 admission(원장 트랜잭션) 실패 — **범위 거절 아님**(같은 파도 둘째 범위 BUCKET_USAGE_STALE · owner 충돌 = 설계 ④ 보호) | `a112_owner_scope_trading_test.go` `TestAnActivatedTwoScopeMarketIssuesOneFirstLegPerScope`(STALE 타입 없음) · `a112_qfinal_family_test.go` `TestTwoFamiliesRacingForOneOwnerScopeLeaveOneTransaction`(owner conflict) | yes — 변이 U04(STALE 을 범위로 재분류) CAUGHT | yes |
