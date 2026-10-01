# Branch Test Map: `loadProductionRiskEntries`

- Source SHA-256: `38de0b7d846b0a1af1b01bc24eb94adc673ca953a3f125f3130adbc3bb4f58c2`; AST branch locations are authoritative.
- Revision: **modified (a112 5.2.2.2 리뷰 수리, 2026-10-01).** 리뷰 수리(J4 = (A)): 편집 전 B6(`err != nil || scopeLatches != 0` → `scope latch present` 하나)을 B6(조회 결함 → `scope latch unreadable: %w`) 과 B7(latch 존재 → `ErrProductionRiskScopeRefused`)로 나눴다 — 둘 다 거절(판정 불변), latch 만 범위 국소 신원. 13 분기 → 14.
- 편집 전 번들: `analysis/measurements/lot-5.2.2.2-fix/pre-edit/internal-riskbucket--loadproductionriskentries/`. 변이 원장 `analysis/measurements/lot-5.2.2.2-fix/mutation-5.2.2.2-fix.tsv`.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 354:2 — 소유자 UID 없음 | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B2 | if at 357:2 — 원장 파일 검증 실패 | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B3 | if at 366:2 — 열기 실패 | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B4 | if at 371:2 — ping 실패 | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B5 | if at 375:2 — 스키마 핀 불일치(R1 — 별도 change) | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B6 | if at 380:2 — **(분리)** latch 조회 결함 → `scope latch unreadable: %w`(결함) | `a112_owner_scope_trading_test.go` `TestAScopeLatchIsRefusedAloneButALatchReadFaultStops`(표 없음) | yes — `analysis/measurements/lot-5.2.2.2-fix/red-fix.log`(편집 전 범위 거절) · 변이 Y03(결함에 sentinel) CAUGHT | yes |
| B7 | if at 384:2 — **(분리)** 그 범위에 scope latch 있음 → `ErrProductionRiskScopeRefused`(범위 국소) | `a112_owner_scope_trading_test.go` `TestAScopeLatchIsRefusedAloneButALatchReadFaultStops`(latch 행) | yes — 변이 Y02(sentinel 제거) CAUGHT | yes |
| B8 | if at 394:2 — 권한 창 불일치 | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B9 | range at 398:2 — 차원 순회 | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B10 | if at 400:3 — 사용량 읽기 실패(원장 결함 — 손상 행 포함) | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B11 | if at 404:3 — latch 된 사용량 | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B12 | if at 414:3 — 정책 출처 구성 실패 | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B13 | if at 424:3 — 스냅숏 출처 구성 실패 | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B14 | if at 433:3 — 권한 항목 구성 실패 | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
