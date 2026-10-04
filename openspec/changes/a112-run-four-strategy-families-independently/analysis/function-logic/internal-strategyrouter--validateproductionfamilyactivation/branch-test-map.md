# Branch Test Map: `validateProductionFamilyActivation`

- Source SHA-256: `7842842bc1b1bf9d2be542624169502fcc846a4e60080841887f5072f4fb6507`; AST branch locations are authoritative.
- Revision: **modified (a112 8.8.4-B, 2026-10-01).** 분기 불변(8, difflib 정렬 — B1 · B2 · B5 의 조건만 다시 씀). B1(몸통 결속) · B2(수명) · B5(서술자 필드)는 같은 분기 하나로 `len(failedFields(...)) != 0` — 어긋난 필드 전부. B5 의 표 대조 셋은 `known &&` 로 묶어 모르는 레인이 표 필드까지 탓하지 않게 했다(판정은 앞 판 `!known || …` 과 같다). B6 · B7 · B8 은 이유를 `%w` 로.
- 편집 전 번들: `analysis/measurements/lot-8.8.4-B/pre-edit/internal-strategyrouter--validateproductionfamilyactivation/`. 변이 원장 `analysis/measurements/lot-8.8.4-B/mutation-8.8.4-B.tsv`.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 543:2 — 몸통 결속(복합) → `%w: body binding: <어긋난 필드 전부>` | `a112_activation_error_fields_test.go` `TestEveryActivationRefusalNamesItsFieldAndKeepsItsSentinel`(body binding: … 열하나 단일 + 세 필드 동시) · `TestAnActivationWhoseBindingsDoNotMatchPromotesNothing` | yes — 변이 B2 · B3 CAUGHT(`analysis/measurements/lot-8.8.4-B/mutation-8.8.4-B.tsv`) | yes |
| B2 | if at 563:2 — 수명(복합) → `%w: lifetime: <어긋난 항목 전부>` | `a112_activation_error_fields_test.go` `TestEveryActivationRefusalNamesItsFieldAndKeepsItsSentinel`(lifetime: … 넷 + 둘 동시) · `TestAnActivationOutsideItsApprovedLifetimePromotesNothing` | yes — 편집 전 메시지에 필드명 없음(`red-8.8.4-B.log`) | yes |
| B3 | if at 575:2 — 만료 → familyActivationRemaining 의 Expired 오류 그대로 | `a112_activation_error_fields_test.go` `TestEveryActivationRefusalNamesItsFieldAndKeepsItsSentinel`(expired) | no — 갈래 불변 | yes |
| B4 | range at 593:2 — 서술자 순회 | `TestAVerifiedFourFamilyActivationPromotesExactlyTheLanesItNames` | no — 갈래 불변 | yes |
| B5 | if at 596:3 — 서술자 필드(복합) → `%w: descriptors[lane_id=…]: <어긋난 필드>` | `a112_activation_error_fields_test.go` `TestEveryActivationRefusalNamesItsFieldAndKeepsItsSentinel`(descriptor: unknown lane · horizon drift) | yes — 변이 B9(known 가드 탈락) CAUGHT(`analysis/measurements/lot-8.8.4-B/mutation-8.8.4-B.tsv`) | yes |
| B6 | if at 608:3 — effective ON 인데 desired ON 아님 → 이유 | `a112_activation_error_fields_test.go` `TestEveryActivationRefusalNamesItsFieldAndKeepsItsSentinel`(descriptor: effective without desired) | yes — 편집 전 메시지에 필드명 없음(`red-8.8.4-B.log`) | yes |
| B7 | if at 612:3 — 중복 레인 → `%w: descriptors: duplicate lane_id …` | `a112_activation_error_fields_test.go` `TestEveryActivationRefusalNamesItsFieldAndKeepsItsSentinel`(descriptor: duplicate lane) | yes — 편집 전 메시지에 필드명 없음(`red-8.8.4-B.log`) | yes |
| B8 | if at 617:2 — 네 레인이 아님 → `%w: descriptors: N of M lanes` | `a112_activation_error_fields_test.go` `TestEveryActivationRefusalNamesItsFieldAndKeepsItsSentinel`(descriptor: three of four) · `TestAnActivationWhoseDescriptorSetIsNotExactlyTheFourPromotesNothing` | yes — 편집 전 메시지에 필드명 없음(`red-8.8.4-B.log`) | yes |
