# Branch Test Map: `TestAGatedFamilyMustNotShrinkTheMarketIntoTheExactlyOneValve (시험)`

- Source SHA-256: `49511cdebc1ba5506737d375b775aa2a533b358701fb531f33f41051f2400fcf`; AST branch locations are authoritative.
- Revision: **modified (a112 repin-1e25b3a3, 2026-10-01).** 편집 전 8 분기 → 9(B9 새로, 끝에 덧붙음). 8.5 응답 로트 ⑦: FAMILY_GATE_CLOSED 닫힘이 판정 활성화를 싣는지 단언(보이스 3 P2-1 — 변이 E9).
- 편집 전 번들: `analysis/measurements/repin-1e25b3a3/pre-edit/internal-app-engine--testagatedfamilymustnotshrinkthemarketintotheexactlyonevalve/`. 변이 원장 `analysis/measurements/lot-8.5-R/mutation-8.5-R.tsv`.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | range at 468:2 — KR 레인 순회 | 이 시험 자신(시험 코드) — `go test -tags tossos_testseams ./internal/app/engine` GREEN(`lot-8.5-R/verify-8.5-R.log`) | no — 시험 코드, 갈래 불변 | yes |
| B2 | if at 469:3 — 지속형이 아니면 건너뜀 | 이 시험 자신(시험 코드) — `go test -tags tossos_testseams ./internal/app/engine` GREEN(`lot-8.5-R/verify-8.5-R.log`) | no — 시험 코드, 갈래 불변 | yes |
| B3 | if at 472:3 — 비정상 실패로도 안 잠기면 → Fatal | 이 시험 자신(시험 코드) — `go test -tags tossos_testseams ./internal/app/engine` GREEN(`lot-8.5-R/verify-8.5-R.log`) | no — 시험 코드, 갈래 불변 | yes |
| B4 | if at 477:2 — 잠근 레인 수 ≠ 1 → Fatal | 이 시험 자신(시험 코드) — `go test -tags tossos_testseams ./internal/app/engine` GREEN(`lot-8.5-R/verify-8.5-R.log`) | no — 시험 코드, 갈래 불변 | yes |
| B5 | if at 482:2 — 시장이 열림 → Fatal(고장이 시스템을 관대하게) | 이 시험 자신(시험 코드) — `go test -tags tossos_testseams ./internal/app/engine` GREEN(`lot-8.5-R/verify-8.5-R.log`) | no — 시험 코드, 갈래 불변 | yes |
| B6 | if at 487:2 — 사유 ≠ FAMILY_GATE_CLOSED → Fatal | 이 시험 자신(시험 코드) — `go test -tags tossos_testseams ./internal/app/engine` GREEN(`lot-8.5-R/verify-8.5-R.log`) | no — 시험 코드, 갈래 불변 | yes |
| B7 | if at 490:2 — 거절 수 ≠ 경로 수 → Fatal | 이 시험 자신(시험 코드) — `go test -tags tossos_testseams ./internal/app/engine` GREEN(`lot-8.5-R/verify-8.5-R.log`) | no — 시험 코드, 갈래 불변 | yes |
| B8 | if at 494:2 — 닫힌 시장이 dispatch 에 건넴 → Fatal | 이 시험 자신(시험 코드) — `go test -tags tossos_testseams ./internal/app/engine` GREEN(`lot-8.5-R/verify-8.5-R.log`) | no — 시험 코드, 갈래 불변 | yes |
| B9 | if at 499:2 — **(새)** FAMILY_GATE_CLOSED 닫힘이 판정 활성화를 안 실음 → Fatal | 이 시험 자신(시험 코드) — `go test -tags tossos_testseams ./internal/app/engine` GREEN(`lot-8.5-R/verify-8.5-R.log`) | yes — 8.5 응답 로트에서 새 갈래(편집 전 없음); 변이 E9 CAUGHT(`analysis/measurements/lot-8.5-R/mutation-8.5-R.tsv`) | yes |
