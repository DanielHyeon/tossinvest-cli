# Branch Test Map: `validateProductionFamilyActivation`

- Source SHA-256: `bac16c04b38d49c479a381fb326d7dd066e5525997322619d8ceb496e3ddb0d8`; AST branch locations are authoritative.
- Revision: **modified (a112 8.5-R, 2026-10-01).** 분기 불변(8). 서술자 거절 셋(B5 · B6 · B7)의 메시지가 lane_id 원문 대신 위치 `descriptors[i]` 와 필드명만 싣는다(8.5 응답 로트 ③ — codex r2 P2: 매니페스트의 임의 문자열 · 개행이 오류 문장으로 새지 않음). 순회가 색인을 받는다(`for index, descriptor := range`).
- 편집 전 번들: `analysis/measurements/lot-8.5-R/pre-edit/internal-strategyrouter--validateproductionfamilyactivation/`. 변이 원장 `analysis/measurements/lot-8.5-R/mutation-8.5-R.tsv`.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 544:2 — 몸통 결속(복합) → `%w: body binding: <어긋난 필드 전부>` | `a112_activation_error_fields_test.go` `TestEveryActivationRefusalNamesItsFieldAndKeepsItsSentinel`(body binding: … 열하나 단일 + 세 필드 동시) · `TestAnActivationWhoseBindingsDoNotMatchPromotesNothing` | yes — 변이 B2 · B3 CAUGHT(`analysis/measurements/lot-8.8.4-B/mutation-8.8.4-B.tsv`) | yes |
| B2 | if at 564:2 — 수명(복합) → `%w: lifetime: <어긋난 항목 전부>` | `a112_activation_error_fields_test.go` `TestEveryActivationRefusalNamesItsFieldAndKeepsItsSentinel`(lifetime: … 넷 + 둘 동시 + 8.5 응답 로트 ⑧ 셋 — issued_at · expires_at 비정규 · 발급 = 만료 · 만료 < 발급) · `TestAnActivationOutsideItsApprovedLifetimePromotesNothing` | no — 갈래 불변(새 모양은 기존 항의 핀 — 변이 T565 · T566 · T569/m569) | yes |
| B3 | if at 576:2 — 만료 → familyActivationRemaining 의 Expired 오류 그대로 | `a112_activation_error_fields_test.go` `TestEveryActivationRefusalNamesItsFieldAndKeepsItsSentinel`(expired) | no — 갈래 불변 | yes |
| B4 | range at 596:2 — 서술자 순회 | `TestAVerifiedFourFamilyActivationPromotesExactlyTheLanesItNames` | no — 갈래 불변 | yes |
| B5 | if at 599:3 — 서술자 필드(복합) → `%w: descriptors[i]: <어긋난 필드>`(lane_id 원문 없음) | `a112_activation_error_fields_test.go` `TestEveryActivationRefusalNamesItsFieldAndKeepsItsSentinel`(descriptor: unknown lane · horizon drift · 개행 lane id · effective 열거 밖 — 8.5 보이스 3 P1-1 · 보이스 1 m602) | yes — `red-8.5-R.log` — 편집 전 `descriptors[lane_id=<원문>]`(개행 포함) | yes |
| B6 | if at 611:3 — effective ON 인데 desired ON 아님 → `descriptors[i]: effective ON without desired ON` | `a112_activation_error_fields_test.go` `TestEveryActivationRefusalNamesItsFieldAndKeepsItsSentinel`(descriptor: effective without desired) | yes — `red-8.5-R.log` | yes |
| B7 | if at 615:3 — 중복 레인 → `%w: descriptors[i]: duplicate lane_id`(원문 없음) | `a112_activation_error_fields_test.go` `TestEveryActivationRefusalNamesItsFieldAndKeepsItsSentinel`(descriptor: duplicate lane) | yes — `red-8.5-R.log` | yes |
| B8 | if at 620:2 — 네 레인이 아님 → `%w: descriptors: N of M lanes` | `a112_activation_error_fields_test.go` `TestEveryActivationRefusalNamesItsFieldAndKeepsItsSentinel`(descriptor: three of four) · `TestAnActivationWhoseDescriptorSetIsNotExactlyTheFourPromotesNothing` | yes — 편집 전 메시지에 필드명 없음(`red-8.8.4-B.log`) | yes |
