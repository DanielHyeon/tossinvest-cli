# Branch Test Map: `admit`

- Source SHA-256: `3c82793300b39454ccac2ff41fe97c0b76ed2190534c5a76c0f9f7abd59652e5`; AST branch locations are authoritative.
- Revision: **modified (a112 5.2.2.2 리뷰 수리, 2026-10-01).** 리뷰 수리: 80ae96a5 의 B4(수집 오류가 범위 거절 타입이면 결과에 싣기 — errors.As)를 지우고 B3 에서 **수집 오류 그대로**를 `cause` 에 싣는다 — 범위 거절 타입과 결함 원인의 신원이 모두 dispatch 로 건너간다(문구 불변). 7 분기 → 6.
- 편집 전 번들: `analysis/measurements/lot-5.2.2.2-fix/pre-edit/internal-app-engine--strategyfirstlegadmissionbridge.admit/`. 변이 원장 `analysis/measurements/lot-5.2.2.2-fix/mutation-5.2.2.2-fix.tsv`.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 74:2 — 결과 검증 거절 | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B2 | if at 77:2 — bridge · loader · Guardian 부재 | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B3 | if at 81:2 — 1차 레그 권한 수집 실패 → `AuthorityCollectionFailed`, 수집 오류를 `cause` 로 운반 | `a112_owner_scope_trading_test.go` `TestACorruptLedgerRowStopsTheCycleWithItsCause` · `TestAScopeWithoutItsOwnAccountAuthorityIsRefusedAloneAndRecorded` | yes — `analysis/measurements/lot-5.2.2.2-fix/red-fix.log`(편집 전 원인 신원 소실) · 변이 Y19 CAUGHT | yes |
| B4 | if at 86:2 — 권한 불일치 | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B5 | if at 90:2 — Guardian precheck 실패 | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B6 | if at 94:2 — 원자 admission 실패(원장 `BUCKET_USAGE_STALE` 은 cause 없이 문구로 — 타입 없음) | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
