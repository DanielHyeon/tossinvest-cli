# Branch Test Map: `collectOverflowing (시험 도우미)`

- Source SHA-256: `60d5adbcf550c3ff422046d010bfe037481c7fa39e5b1561afb5c1d19a1f2115`; AST branch locations are authoritative.
- Revision: **modified (a112 repin-1e25b3a3, 2026-10-01).** 편집 전 5 분기 → 6(B6 새로, 끝에 덧붙음). 8.5 응답 로트 ⑦: 가변 인자 configure 로 수집 직전 적재기를 바꿀 수 있게(관문 아래 실행).
- 편집 전 번들: `analysis/measurements/repin-1e25b3a3/pre-edit/internal-app-engine--collectoverflowing/`. 변이 원장 `analysis/measurements/lot-8.5-R/mutation-8.5-R.tsv`.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | range at 90:2 — 종목마다 경로 항목 생성 | 이 시험 자신(시험 코드) — `go test -tags tossos_testseams ./internal/app/engine` GREEN(`lot-8.5-R/verify-8.5-R.log`) | no — 시험 코드, 갈래 불변 | yes |
| B2 | if at 92:3 — 소유자 열쇠 생성 실패 → Fatal | 이 시험 자신(시험 코드) — `go test -tags tossos_testseams ./internal/app/engine` GREEN(`lot-8.5-R/verify-8.5-R.log`) | no — 시험 코드, 갈래 불변 | 도달 불가(고정 입력) |
| B3 | if at 98:3 — 후보 경로 픽스처 실패 → Fatal | 이 시험 자신(시험 코드) — `go test -tags tossos_testseams ./internal/app/engine` GREEN(`lot-8.5-R/verify-8.5-R.log`) | no — 시험 코드, 갈래 불변 | 도달 불가(고정 입력) |
| B4 | range at 115:3 — 제안 적재 스텁의 대상 순회 | 이 시험 자신(시험 코드) — `go test -tags tossos_testseams ./internal/app/engine` GREEN(`lot-8.5-R/verify-8.5-R.log`) | no — 시험 코드, 갈래 불변 | yes |
| B5 | if at 119:4 — 수락 결과 픽스처 실패 → Fatal | 이 시험 자신(시험 코드) — `go test -tags tossos_testseams ./internal/app/engine` GREEN(`lot-8.5-R/verify-8.5-R.log`) | no — 시험 코드, 갈래 불변 | 도달 불가(고정 입력) |
| B6 | range at 126:2 — **(새)** configure 적용 | 이 시험 자신(시험 코드) — `go test -tags tossos_testseams ./internal/app/engine` GREEN(`lot-8.5-R/verify-8.5-R.log`) | yes — 8.5 응답 로트에서 새 갈래(편집 전 없음) | yes |
