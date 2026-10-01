# Branch Test Map: `bindProductionRiskInputs`

- Source SHA-256: `38de0b7d846b0a1af1b01bc24eb94adc673ca953a3f125f3130adbc3bb4f58c2`; AST branch locations are authoritative.
- Revision: **modified (a112 5.2.2.2 리뷰 수리, 2026-10-01).** 리뷰 수리(J4 = (A)): B4(서명 정책에 그 종목의 섹터 매핑 없음)의 오류를 `ErrProductionRiskScopeRefused` 로 감쌌다 — 범위 국소 신원. 분기 · 판정 불변.
- 편집 전 번들: `analysis/measurements/lot-5.2.2.2-fix/pre-edit/internal-riskbucket--bindproductionriskinputs/`. 변이 원장 `analysis/measurements/lot-5.2.2.2-fix/mutation-5.2.2.2-fix.tsv`.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 285:2 — 봉인된 결과 불일치(무결성 — 결함) | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B2 | if at 292:2 — 지원하지 않는 horizon | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B3 | if at 296:2 — 전략 위험 매핑 없음(결함 — 범위 국소로 넓히지 않음) | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B4 | if at 300:2 — 서명 정책에 그 종목 섹터 매핑 없음 → **`ErrProductionRiskScopeRefused`**(범위 국소) | `a112_owner_scope_trading_test.go` `TestARiskScopeOutsideTheSignedPolicyIsRefusedAloneInEitherOrder` | yes — `analysis/measurements/lot-5.2.2.2-fix/red-fix.log` · 변이 Y01(sentinel 제거) CAUGHT | yes |
| B5 | if at 308:2 — 최악 체결가 권한 무효 | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B6 | if at 314:2 — FX 결속 실패 | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B7 | if at 321:2 — 정책 창 불일치 | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B8 | if at 339:2 — 예약 정책 무효 | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B9 | range at 344:2 — 차원 순회 | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B10 | if at 345:3 — 한도 해석 실패 | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
