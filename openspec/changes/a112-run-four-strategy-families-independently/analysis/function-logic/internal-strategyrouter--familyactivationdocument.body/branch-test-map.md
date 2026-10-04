# Branch Test Map: `FamilyActivationDocument.body`

- Source SHA-256: `bac16c04b38d49c479a381fb326d7dd066e5525997322619d8ceb496e3ddb0d8`; AST branch locations are authoritative.
- Revision: **modified (a112 8.8.4-B, 2026-10-01).** 분기 불변(5). B1 · B3 의 맨 sentinel 반환에 이유(시장 표 없음 · 모르는 가족 이름)를 `%w` 로 붙였다.
- 편집 전 번들: `analysis/measurements/lot-8.8.4-B/pre-edit/internal-strategyrouter--familyactivationdocument.body/`. 변이 원장 `analysis/measurements/lot-8.8.4-B/mutation-8.8.4-B.tsv`.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 228:2 — 서술자 표가 없는 시장 → `%w: market … has no descriptor table` | `a112_activation_error_fields_test.go` `TestTheDocumentPathNamesItsRefusalToo` | yes — 편집 전 메시지에 필드명 없음(`red-8.8.4-B.log`) | yes |
| B2 | range at 232:2 — 켤 가족 순회 | `TestTheCommittedGoldenManifestMatchesItsPinAndPromotesTheFourLanes` · `a112_activation_error_fields_test.go` `TestTheDocumentPathNamesItsRefusalToo` | no — 갈래 불변 | yes |
| B3 | if at 233:3 — 모르는 가족 이름 → `%w: on: unknown family …` | `a112_activation_error_fields_test.go` `TestTheDocumentPathNamesItsRefusalToo` | yes — 편집 전 메시지에 필드명 없음(`red-8.8.4-B.log`) | yes |
| B4 | range at 239:2 — 표 순회로 네 서술자 생성 | `TestTheAuthoringEncoderStillProducesTheCommittedGoldenBytes` | no — 갈래 불변 | yes |
| B5 | if at 241:3 — 켠 가족이면 ON | `TestTheAuthoringEncoderStillProducesTheCommittedGoldenBytes` | no — 갈래 불변 | yes |
