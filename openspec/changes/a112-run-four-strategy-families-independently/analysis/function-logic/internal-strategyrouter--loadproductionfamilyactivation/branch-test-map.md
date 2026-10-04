# Branch Test Map: `LoadProductionFamilyActivation`

- Source SHA-256: `57123f814187d9201ec6999d6375adc649b579c89beb6c4731f36cc51243b005`; AST branch locations are authoritative.
- Revision: **modified (a112 8.5-R, 2026-10-01).** 분기 불변(10). B5(읽기 결함)의 안쪽 오류를 `%w` → `%v` 로(8.5 응답 로트 ④ — 보이스 2 P2-1): 편집 전(f473d815) `errors.Is` 사슬로 복원 — 읽기 결함이 공유 읽기 함수의 sentinel(ErrProductionRouteUnavailable)까지 만족하던 둘째 신원 제거. B10 행 정정(보이스 3 P2-5).
- 편집 전 번들: `analysis/measurements/lot-8.5-R/pre-edit/internal-strategyrouter--loadproductionfamilyactivation/`. 변이 원장 `analysis/measurements/lot-8.5-R/mutation-8.5-R.tsv`.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 434:2 — 핀이 비었으면 미선언 → `%w(Undeclared): manifest_digest pin is empty` — **맨 앞 순서 불변**(엔진 판별이 errors.Is 로 의존) | `a112_activation_error_fields_test.go` `TestEveryActivationRefusalNamesItsFieldAndKeepsItsSentinel`(undeclared pin) · `TestOnlyAnEmptyPinMeansTheActivationWasNeverDeclared` | yes — 변이 B1(sentinel 탈락) CAUGHT(`analysis/measurements/lot-8.8.4-B/mutation-8.8.4-B.tsv`) | yes |
| B2 | if at 437:2 — ctx 가 nil → `%w: context is nil` | `a112_activation_error_fields_test.go` `TestTheDocumentPathNamesItsRefusalToo` | yes — 편집 전 메시지에 필드명 없음(`red-8.8.4-B.log`) | yes |
| B3 | if at 440:2 — ctx 취소 → ctx.Err() 그대로 | `TestACancelledContextPromotesNothing` | no — 갈래 불변 | yes |
| B4 | if at 449:2 — 설정 결속(복합 — 분기 하나) → `%w: config binding: <어긋난 필드 전부>` | `a112_activation_error_fields_test.go` `TestEveryActivationRefusalNamesItsFieldAndKeepsItsSentinel`(config binding: … 아홉 단일 + 세 필드 동시) | yes — 변이 B2 · B3 · B4 · B5 CAUGHT(`analysis/measurements/lot-8.8.4-B/mutation-8.8.4-B.tsv`) | yes — market 항은 아래 불변식(가림) 참조 |
| B5 | if at 468:2 — 매니페스트 파일 읽기 결함 → `%w: manifest file …: %v(읽기 함수 오류 문장)` — 불일치와 다른 종류, 사슬에 활성화 sentinel 하나 | `a112_activation_error_fields_test.go` `TestEveryActivationRefusalNamesItsFieldAndKeepsItsSentinel`(manifest file missing — 모든 모양에서 `errors.Is(err, ErrProductionRouteUnavailable)==false`) | yes — `red-8.5-R.log` — 편집 전 둘째 `%w` 로 ErrProductionRouteUnavailable 도 만족 | yes |
| B6 | if at 471:2 — **(새)** 핀이 파일 바이트와 다름 → `%w: manifest_digest: …` | `a112_activation_error_fields_test.go` `TestEveryActivationRefusalNamesItsFieldAndKeepsItsSentinel`(pin does not match the file bytes) · `TestBytesThatChangedAfterTheDeploymentPinnedThemPromoteNothing` | yes — 편집 전 메시지에 필드명 없음(`red-8.8.4-B.log`) | yes |
| B7 | if at 476:2 — 해석 거절 → 해석기가 붙인 이유 그대로(sentinel 포함) | `a112_activation_error_fields_test.go` `TestEveryActivationRefusalNamesItsFieldAndKeepsItsSentinel`(bytes not canonical · trailing data · unknown field) | yes — 변이 B8(맨 sentinel) CAUGHT(`analysis/measurements/lot-8.8.4-B/mutation-8.8.4-B.tsv`) | yes |
| B8 | if at 482:2 — 폐기 → `%w(Revoked): revoked=true` | `a112_activation_error_fields_test.go` `TestEveryActivationRefusalNamesItsFieldAndKeepsItsSentinel`(revoked) | yes — 편집 전 메시지에 필드명 없음(`red-8.8.4-B.log`) | yes |
| B9 | if at 486:2 — 검증 거절 → 검증기 오류 그대로 | `a112_activation_error_fields_test.go` `TestEveryActivationRefusalNamesItsFieldAndKeepsItsSentinel`(body binding · lifetime · descriptor · expired) | no — 갈래 불변 | yes |
| B10 | if at 489:2 — 끝의 ctx 취소 재확인 | 도달 불가(결정적 입력 없음 — 진입 확인 :440 뒤 ctx 를 보지 않는 동기 읽기 · 검증뿐) — census/검토로 닫음. 인용하던 `TestACancelledContextPromotesNothing` 은 앞의 확인(B3)만 지난다(보이스 3 커버리지: 이 몸통 count=0) | no — 갈래 불변 | 도달 불가 — census/검토 |
