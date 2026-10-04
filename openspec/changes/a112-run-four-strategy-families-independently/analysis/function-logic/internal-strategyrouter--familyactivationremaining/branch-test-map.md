# Branch Test Map: `familyActivationRemaining`

- Source SHA-256: `7842842bc1b1bf9d2be542624169502fcc846a4e60080841887f5072f4fb6507`; AST branch locations are authoritative.
- Revision: **modified (a112 8.8.4-B, 2026-10-01).** 분기 불변(1). 만료 반환에 `expires_at` 과 두 시각을 `%w` 로 붙였다(ErrProductionFamilyActivationExpired 보존).
- 편집 전 번들: `analysis/measurements/lot-8.8.4-B/pre-edit/internal-strategyrouter--familyactivationremaining/`. 변이 원장 `analysis/measurements/lot-8.8.4-B/mutation-8.8.4-B.tsv`.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 385:2 — 만료(지금이 expires_at 이상) → `%w: expires_at … is not after …` | `a112_activation_error_fields_test.go` `TestEveryActivationRefusalNamesItsFieldAndKeepsItsSentinel`(expired) · `TestTheLeaseCeilingOnlyEverShrinks` · `TestLoadingAndLeasingJudgeExpiryAtTheSameInstant` | yes — 편집 전 필드명 없음; 변이 B10(sentinel 탈락) CAUGHT(`analysis/measurements/lot-8.8.4-B/mutation-8.8.4-B.tsv`) | yes |
