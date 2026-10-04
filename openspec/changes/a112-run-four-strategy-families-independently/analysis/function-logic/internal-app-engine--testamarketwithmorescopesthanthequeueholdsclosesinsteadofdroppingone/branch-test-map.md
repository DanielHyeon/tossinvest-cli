# Branch Test Map: `TestAMarketWithMoreScopesThanTheQueueHoldsClosesInsteadOfDroppingOne (시험)`

- Source SHA-256: `8581d7275bae005df65081299b204c1eeecb6f2ad6234aa4b00748538f04ad73`; AST branch locations are authoritative.
- Revision: **modified (a112 repin-1e25b3a3, 2026-10-01).** 편집 전 5 분기 → 7(B6 · B7 새로, 끝에 덧붙음 — 앞 다섯은 번호 그대로). 8.5 응답 로트 ⑦: 검증된 관문 아래에서 같은 넘침을 다시 돌려 QUEUE_OVERFLOW 닫힘이 판정 활성화를 싣는지 단언(보이스 3 P2-1 — 변이 E7).
- 편집 전 번들: `analysis/measurements/repin-1e25b3a3/pre-edit/internal-app-engine--testamarketwithmorescopesthanthequeueholdsclosesinsteadofdroppingone/`. 변이 원장 `analysis/measurements/lot-8.5-R/mutation-8.5-R.tsv`.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | range at 47:2 — 종목 Capacity+1 개 생성 | 이 시험 자신(시험 코드) — `go test -tags tossos_testseams ./internal/app/engine` GREEN(`lot-8.5-R/verify-8.5-R.log`) | no — 시험 코드, 갈래 불변 | yes |
| B2 | if at 51:2 — 미선언 실행의 사유 ≠ QUEUE_OVERFLOW → Fatal | 이 시험 자신(시험 코드) — `go test -tags tossos_testseams ./internal/app/engine` GREEN(`lot-8.5-R/verify-8.5-R.log`) | no — 시험 코드, 갈래 불변 | yes |
| B3 | if at 54:2 — 닫힌 시장이 열림 · 항목 있음 → Fatal | 이 시험 자신(시험 코드) — `go test -tags tossos_testseams ./internal/app/engine` GREEN(`lot-8.5-R/verify-8.5-R.log`) | no — 시험 코드, 갈래 불변 | yes |
| B4 | if at 57:2 — 버린 수 0 → Fatal(조용한 유실 금지) | 이 시험 자신(시험 코드) — `go test -tags tossos_testseams ./internal/app/engine` GREEN(`lot-8.5-R/verify-8.5-R.log`) | no — 시험 코드, 갈래 불변 | yes |
| B5 | if at 61:2 — 넘침이 중재 코드를 빌림 → Fatal | 이 시험 자신(시험 코드) — `go test -tags tossos_testseams ./internal/app/engine` GREEN(`lot-8.5-R/verify-8.5-R.log`) | no — 시험 코드, 갈래 불변 | yes |
| B6 | if at 73:2 — **(새)** 관문 아래 실행의 사유 ≠ QUEUE_OVERFLOW → Fatal | 이 시험 자신(시험 코드) — `go test -tags tossos_testseams ./internal/app/engine` GREEN(`lot-8.5-R/verify-8.5-R.log`) | yes — 8.5 응답 로트에서 새 갈래(편집 전 없음) | yes |
| B7 | if at 76:2 — **(새)** QUEUE_OVERFLOW 닫힘이 판정 활성화를 안 실음 → Fatal | 이 시험 자신(시험 코드) — `go test -tags tossos_testseams ./internal/app/engine` GREEN(`lot-8.5-R/verify-8.5-R.log`) | yes — 8.5 응답 로트에서 새 갈래(편집 전 없음); 변이 E7 CAUGHT(`analysis/measurements/lot-8.5-R/mutation-8.5-R.tsv`) | yes |
