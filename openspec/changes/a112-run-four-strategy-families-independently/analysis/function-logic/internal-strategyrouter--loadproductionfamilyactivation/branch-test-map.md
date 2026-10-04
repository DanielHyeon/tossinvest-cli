# Branch Test Map: `LoadProductionFamilyActivation`

- Source SHA-256: `7842842bc1b1bf9d2be542624169502fcc846a4e60080841887f5072f4fb6507`; AST branch locations are authoritative.
- Revision: **modified (a112 8.8.4-B, 2026-10-01).** 편집 전 9 분기 → 10(재번호 `lot-8.8.4-B/renumber.txt`, difflib 정렬): 편집 전 B5(`err != nil || digest != pin`)가 B5(읽기 결함 — 읽기 함수의 오류를 `%w` 사슬에 보존) · B6(새 — 핀 불일치, `manifest_digest`)으로 갈렸다(Manager 판정 — 결함과 불일치는 다른 종류). 편집 전 B6~B9 → B7~B10. B4(설정 결속)는 같은 분기 하나로 조건을 `len(failedFields(...)) != 0` 으로 바꿔 어긋난 필드 전부를 싣는다. 나머지 맨 sentinel 반환은 이유를 `%w` 로 붙였다.
- 편집 전 번들: `analysis/measurements/lot-8.8.4-B/pre-edit/internal-strategyrouter--loadproductionfamilyactivation/`. 변이 원장 `analysis/measurements/lot-8.8.4-B/mutation-8.8.4-B.tsv`.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 434:2 — 핀이 비었으면 미선언 → `%w(Undeclared): manifest_digest pin is empty` — **맨 앞 순서 불변**(엔진 판별이 errors.Is 로 의존) | `a112_activation_error_fields_test.go` `TestEveryActivationRefusalNamesItsFieldAndKeepsItsSentinel`(undeclared pin) · `TestOnlyAnEmptyPinMeansTheActivationWasNeverDeclared` | yes — 변이 B1(sentinel 탈락) CAUGHT(`analysis/measurements/lot-8.8.4-B/mutation-8.8.4-B.tsv`) | yes |
| B2 | if at 437:2 — ctx 가 nil → `%w: context is nil` | `a112_activation_error_fields_test.go` `TestTheDocumentPathNamesItsRefusalToo` | yes — 편집 전 메시지에 필드명 없음(`red-8.8.4-B.log`) | yes |
| B3 | if at 440:2 — ctx 취소 → ctx.Err() 그대로 | `TestACancelledContextPromotesNothing` | no — 갈래 불변 | yes |
| B4 | if at 449:2 — 설정 결속(복합 — 분기 하나) → `%w: config binding: <어긋난 필드 전부>` | `a112_activation_error_fields_test.go` `TestEveryActivationRefusalNamesItsFieldAndKeepsItsSentinel`(config binding: … 아홉 단일 + 세 필드 동시) | yes — 변이 B2 · B3 · B4 · B5 CAUGHT(`analysis/measurements/lot-8.8.4-B/mutation-8.8.4-B.tsv`) | yes |
| B5 | if at 467:2 — 매니페스트 파일 읽기 결함 → `%w: manifest file …: %w(읽기 함수 오류)` — 불일치와 다른 종류 | `a112_activation_error_fields_test.go` `TestEveryActivationRefusalNamesItsFieldAndKeepsItsSentinel`(manifest file missing) | yes — 편집 전에는 불일치와 한 갈래; 변이 B6(%v) · B7(불일치로 위장) CAUGHT(`analysis/measurements/lot-8.8.4-B/mutation-8.8.4-B.tsv`) | yes |
| B6 | if at 470:2 — **(새)** 핀이 파일 바이트와 다름 → `%w: manifest_digest: …` | `a112_activation_error_fields_test.go` `TestEveryActivationRefusalNamesItsFieldAndKeepsItsSentinel`(pin does not match the file bytes) · `TestBytesThatChangedAfterTheDeploymentPinnedThemPromoteNothing` | yes — 편집 전 메시지에 필드명 없음(`red-8.8.4-B.log`) | yes |
| B7 | if at 475:2 — 해석 거절 → 해석기가 붙인 이유 그대로(sentinel 포함) | `a112_activation_error_fields_test.go` `TestEveryActivationRefusalNamesItsFieldAndKeepsItsSentinel`(bytes not canonical · trailing data · unknown field) | yes — 변이 B8(맨 sentinel) CAUGHT(`analysis/measurements/lot-8.8.4-B/mutation-8.8.4-B.tsv`) | yes |
| B8 | if at 481:2 — 폐기 → `%w(Revoked): revoked=true` | `a112_activation_error_fields_test.go` `TestEveryActivationRefusalNamesItsFieldAndKeepsItsSentinel`(revoked) | yes — 편집 전 메시지에 필드명 없음(`red-8.8.4-B.log`) | yes |
| B9 | if at 485:2 — 검증 거절 → 검증기 오류 그대로 | `a112_activation_error_fields_test.go` `TestEveryActivationRefusalNamesItsFieldAndKeepsItsSentinel`(body binding · lifetime · descriptor · expired) | no — 갈래 불변 | yes |
| B10 | if at 488:2 — 끝의 ctx 취소 재확인 | `TestACancelledContextPromotesNothing` | no — 갈래 불변 | 도달 — 측정은 기존 BTM |
