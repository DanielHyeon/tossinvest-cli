# Branch Test Map: `collectStrategyFirstLegAuthority`

- Source SHA-256: `d0d6281292dafcc979edce741a3a2bf98ed348f023267d8198d2436c71ec7291`; AST branch locations are authoritative.
- Revision: **modified (a112 5.2.2.2, 2026-10-01).** 편집 전 11 분기 → 13: 편집 전 B4(시장 단위 개수 관문 `len(proposal.entries) != 1`)를 **지우고**, identity 대조 뒤에 범위별 위험 권한 부재(B5) · 범위별 계좌 권한 부재(B6) · 계보 시장 통화 미지(B7)를 더했다. 편집 전 B5(identity)는 B4, B6~B11 은 B8~B13(조건의 권한 출처만 범위 번들로).
- 편집 전 번들: `analysis/measurements/lot-5.2.2.2/pre-edit/internal-app-engine--productionstrategyfirstlegauthorityloader.collectstrategyfirstlegauthority/`. 변이 원장 `analysis/measurements/lot-5.2.2.2/mutation-5.2.2.2.tsv`.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 256:2 — loader · ctx · 시계 · 원장 · Guardian 부재 → 발급 불가 | 분기 불변 — 편집 전 번들 서술(`pre-edit/`) 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B2 | if at 262:2 — 시장 권한(위험 · 환율 · 계좌 · 일정) 준비 미완 → `paired production authority is incomplete for market`(개수 조건은 6.2 에서 이미 뗌) | 분기 불변 — 편집 전 번들 서술(`pre-edit/`) 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B3 | if at 269:2 — 소유자 범위 선택 실패(0 또는 복수) → `…owner scope is not uniquely authorized by the assembly` — **범위 거절 타입 아님**(위조 의심 → 주기 멈춤) | `a112_first_leg_owner_scope_seal_test.go` `TestTheFirstLegSealRefusesEveryForgeryAxis`(범위 하나 · 둘 · 조정자 순서 세 쌍 × 미선택 범위 · 타 시장) | no — 분기 불변; 두 범위 쌍 재실행은 이 로트 | yes |
| B4 | if at 273:2 — 봉인된 identity 대조(편집 전 B5) — 불일치는 **타입 없는 오류**(범위 거절 아님) | `a112_first_leg_owner_scope_seal_test.go` `TestTheFirstLegSealRefusesEveryForgeryAxis`(세 쌍 × 같은 범위 패자 · 게이트된 레인 · 조건 재작성) · `a112_owner_scope_trading_test.go` `TestAForgedScopeStopsTheCycleBeforeTheNextValidScope` | yes — 변이 X07(identity 불일치를 범위 거절로) CAUGHT 14(`analysis/measurements/lot-5.2.2.2/mutation-5.2.2.2.tsv`) | yes |
| B5 | if at 281:2 — **(새)** 그 범위의 준비된 위험 권한 없음 → `*strategyScopeRefusal`(그 범위만 거절 · 봉투 폴백 없음, J3) | `a112_owner_scope_handoff_test.go` `TestTwoOwnerScopesTradeOnlyWhereEachHasItsOwnAuthority`(두 순서) · `a112_first_leg_owner_scope_seal_test.go` `TestTheFirstLegSealSelectsByScopeInATwoScopePair` | yes — `analysis/measurements/lot-5.2.2.2/red-5.2.2.2.log`(편집 전 개수 관문 거절) · 변이 X08(시장 번들 폴백) · X22(타입 없는 오류) CAUGHT | yes |
| B6 | if at 285:2 — **(새)** 그 범위의 준비된 계좌 권한 없음 → `*strategyScopeRefusal` | `a112_owner_scope_trading_test.go` `TestAScopeWithoutItsOwnAccountAuthorityIsRefusedAloneAndRecorded` | yes — 변이 X09(시장 권한 폴백) · X21(타입 없는 오류) CAUGHT | yes |
| B7 | if at 290:2 — **(새)** 계보 시장 통화 미지(KR/US 밖) → 발급 불가 — 통화는 봉투가 아니라 `Lineage.Market` 에서 유도(A#3) | `a112_owner_scope_trading_test.go` `TestTheFirstLegCurrencyComesFromTheLineageNotTheEnvelope`(유도 경로; 미지 시장 갈래는 시험 seam 으로 못 만듦 — 진입 0) | yes — 변이 X10(봉투 통화) CAUGHT | 유도 경로 yes · 거절 갈래 진입 0 |
| B8 | if at 294:2 — 위험 권한 범위 불일치(편집 전 B6) — 이제 **그 범위의** 번들로 대조 | `a112_owner_scope_trading_test.go` `TestAnActivatedTwoScopeMarketIssuesOneFirstLegPerScope`(범위마다 자기 번들로 통과) | no — 조건 불변, 출처만 범위 번들 | yes |
| B9 | if at 299:2 — 포지션 캠페인 CAS 변경(편집 전 B7) | 분기 불변 — 편집 전 번들 서술(`pre-edit/`) 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B10 | range at 306:2 — 위험 버킷 항목 순회(편집 전 B8) — 범위 번들 | 분기 불변 — 편집 전 번들 서술(`pre-edit/`) 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B11 | if at 317:2 — 가격 단위 무효(편집 전 B9) | 분기 불변 — 편집 전 번들 서술(`pre-edit/`) 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B12 | if at 326:3 — 노출 스냅숏 만료(collect 클로저, 편집 전 B10) — 범위 계좌 권한의 FreshUntil | 분기 불변 — 편집 전 번들 서술(`pre-edit/`) 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B13 | if at 330:3 — 예약 버전 읽기 실패(collect 클로저, 편집 전 B11) | 분기 불변 — 편집 전 번들 서술(`pre-edit/`) 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
