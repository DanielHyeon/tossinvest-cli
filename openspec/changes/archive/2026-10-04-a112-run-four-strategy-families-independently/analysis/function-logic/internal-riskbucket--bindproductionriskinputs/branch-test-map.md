# Branch Test Map: `bindProductionRiskInputs`

- Source SHA-256: `9a74db4abd523da823e02e716258d428d3f385647c3d18121e744f9fac45119e`; AST branch locations are authoritative.
- Revision: **modified (a112 5.2.2.2 리뷰 수리, 2026-10-01).** 리뷰 수리(J4 = (A)): B4(서명 정책에 그 종목의 섹터 매핑 없음)의 오류를 `ErrProductionRiskScopeRefused` 로 감쌌다 — 범위 국소 신원. 분기 · 판정 불변.
- 편집 전 번들: `analysis/measurements/lot-5.2.2.2-fix/pre-edit/internal-riskbucket--bindproductionriskinputs/`. 변이 원장 `analysis/measurements/lot-5.2.2.2-fix/mutation-5.2.2.2-fix.tsv`.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 323:2 — 봉인된 결과 불일치(무결성 — 결함) | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B2 | if at 330:2 — 지원하지 않는 horizon | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B3 | if at 334:2 — 전략 위험 매핑 없음(결함 — 범위 국소로 넓히지 않음) | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B4 | if at 338:2 — 서명 정책에 그 종목 섹터 매핑 없음 → **`ErrProductionRiskScopeRefused`**(범위 국소) | `a112_owner_scope_trading_test.go` `TestARiskScopeOutsideTheSignedPolicyIsRefusedAloneInEitherOrder` | yes — `analysis/measurements/lot-5.2.2.2-fix/red-fix.log` · 변이 Y01(sentinel 제거) CAUGHT | yes |
| B5 | if at 346:2 — 최악 체결가 권한 무효 | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B6 | if at 352:2 — FX 결속 실패 | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B7 | if at 359:2 — 정책 창 불일치 | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B8 | if at 377:2 — 예약 정책 무효 | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B9 | range at 382:2 — 차원 순회 | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B10 | if at 383:3 — 한도 해석 실패 | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
