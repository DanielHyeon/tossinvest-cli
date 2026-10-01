# Branch Test Map: `LoadProductionRiskSnapshotAuthority`

- Source SHA-256: `38de0b7d846b0a1af1b01bc24eb94adc673ca953a3f125f3130adbc3bb4f58c2`; AST branch locations are authoritative.
- Revision: **modified (a112 5.2.2.2 리뷰 수리, 2026-10-01).** 리뷰 수리(J4 = (A)): B6 · B7 의 감싸기를 `%w: %v` → `%w: %w` 로 — 원인의 신원(범위 국소 sentinel · 원장 결함)을 사슬에 보존. 문구 · 분기 · 판정 불변.
- 편집 전 번들: `analysis/measurements/lot-5.2.2.2-fix/pre-edit/internal-riskbucket--loadproductionrisksnapshotauthority/`. 변이 원장 `analysis/measurements/lot-5.2.2.2-fix/mutation-5.2.2.2-fix.tsv`.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 145:2 — ctx · 관측 시각 부재 | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B2 | if at 148:2 — ctx 종료 | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B3 | if at 154:2 — 구성 · 소유자 · 경로 · digest 형식 | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B4 | if at 160:2 — 매니페스트 파일 · digest 불일치 | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B5 | if at 164:2 — 매니페스트 해석 · 서명 검증 실패 | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B6 | if at 169:2 — bind 실패 → `ErrProductionRiskSnapshotUnavailable` + **원인 %w**(범위 국소 sentinel 보존) | `a112_owner_scope_trading_test.go` `TestARiskScopeOutsideTheSignedPolicyIsRefusedAloneInEitherOrder` | yes — `analysis/measurements/lot-5.2.2.2-fix/red-fix.log` · 변이 Y05(%v) CAUGHT(`analysis/measurements/lot-5.2.2.2-fix/mutation-5.2.2.2-fix.tsv`) | yes |
| B7 | if at 173:2 — 원장 항목 적재 실패 → `ErrProductionRiskSnapshotUnavailable` + **원인 %w**(latch sentinel · 원장 결함 원인 보존) | `a112_owner_scope_trading_test.go` `TestACorruptLedgerRowStopsTheCycleWithItsCause` · `TestAScopeLatchIsRefusedAloneButALatchReadFaultStops` | yes — 변이 Y04(결함에도 sentinel) CAUGHT | yes |
