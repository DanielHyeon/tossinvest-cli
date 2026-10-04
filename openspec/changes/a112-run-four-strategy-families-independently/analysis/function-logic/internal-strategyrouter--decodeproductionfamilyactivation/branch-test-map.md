# Branch Test Map: `decodeProductionFamilyActivation`

- Source SHA-256: `bac16c04b38d49c479a381fb326d7dd066e5525997322619d8ceb496e3ddb0d8`; AST branch locations are authoritative.
- Revision: **modified (a112 8.8.4-B, 2026-10-01).** 분기 불변(4). 네 거절이 이유(크기 · json · 뒤 데이터 · 정규 직렬화 아님)를 `%w` 로 싣는다 — json 오류는 원래 오류도 사슬에 남긴다. 호출자(Load B7)는 이 오류를 그대로 돌려준다.
- Revision: **modified (a112 0.5 응답 로트, 2026-10-05 — 리뷰 유지#4 = 보안#2).** B2 의 둘째 `%w` → `%v`: json 오류는 문장으로만 남고 사슬의 신원은 sentinel 하나(shadow 사본 `decodeProductionFamilyShadow` 와 같은 규칙). 분기 · 좌표 불변(한 글자). 편집 전 번들 `analysis/measurements/lot-0.5-response/pre-edit/internal-strategyrouter--decodeproductionfamilyactivation/`, RED `red-decode-identity.log`, 변이 D01(`mutation-0.5-R-run1.tsv`) CAUGHT.
- 편집 전 번들(8.8.4-B): `analysis/measurements/lot-8.8.4-B/pre-edit/internal-strategyrouter--decodeproductionfamilyactivation/`. 변이 원장 `analysis/measurements/lot-8.8.4-B/mutation-8.8.4-B.tsv`.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 510:2 — 크기 0 또는 상한 초과 → `%w: manifest size …` | 읽기 함수가 크기를 먼저 막아 Load 경로로는 도달 불가 — 정규 직렬화 등식이 같은 몫 | no — 갈래 불변 | 도달 불가(Load 경로) |
| B2 | if at 517:2 — json 해석 실패(모르는 필드 포함) → `%w: manifest json: %v`(0.5 응답 로트 전 `%w`) | `a112_activation_error_fields_test.go` `TestEveryActivationRefusalNamesItsFieldAndKeepsItsSentinel`(an unknown field) · `a112_activation_decode_identity_test.go` `TestActivationDecodeFaultCarriesOnlyTheUnavailableSentinel`(사슬 단일 신원) | yes — 편집 전 메시지에 필드명 없음(`red-8.8.4-B.log`) · 0.5: 편집 전 `io.ErrUnexpectedEOF` · `*json.SyntaxError` · `*json.UnmarshalTypeError` 가 둘째 신원(`lot-0.5-response/red-decode-identity.log`) | yes |
| B3 | if at 520:2 — 문서 뒤 데이터 → `%w: trailing data …` | `a112_activation_error_fields_test.go` `TestEveryActivationRefusalNamesItsFieldAndKeepsItsSentinel`(trailing data after the document) | yes — 편집 전 메시지에 필드명 없음(`red-8.8.4-B.log`) | yes |
| B4 | if at 524:2 — 정규 직렬화와 다름 → `%w: manifest bytes are not …canonical…` | `a112_activation_error_fields_test.go` `TestEveryActivationRefusalNamesItsFieldAndKeepsItsSentinel`(bytes not canonical) · `TestAnActivationWhoseBytesAreNotCanonicalPromotesNothing` | yes — 편집 전 메시지에 필드명 없음(`red-8.8.4-B.log`) | yes |
