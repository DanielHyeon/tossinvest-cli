# Branch Test Map: `strategyLaneRuntime.record`

- Source SHA-256: `95611377145456c4906da5ebb95eca97dd025df0c1fb255ebc443c1ce12d5d36`; AST branch locations are authoritative.
- Revision: **modified (a112 7.3, 2026-10-01).** 편집 전 2 분기 → 4: 시장 인자를 받고, 물결 맵 지연 생성(B2) · 포화 상한 아래 증가(B3)를 더했다. 관측마다 물결 번호를 찍는다(B4 몸통).
- 편집 전 번들: `analysis/measurements/lot-7.3/pre-edit/internal-app-engine--strategylaneruntime.record/`. 변이 원장 `analysis/measurements/lot-7.3/mutation-7.3.tsv`.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 323:2 — nil 런타임 또는 관측 0 → 물결이 아님(번호 안 올림) | 진입 0 — evaluate 는 레인 넷을 늘 넘긴다(시장마다 넷) | no — 편집 전과 같은 조건 | n/a |
| B2 | if at 328:2 — **(새)** 물결 맵 없음 → 생성(첫 물결) | `a112_lane_coordinator_projection_test.go` `TestTheCycleGenerationIsTheMarketWaveInWhichTheLaneWasLastObserved` | yes — 편집 전 컴파일 실패 | yes |
| B3 | if at 331:2 — **(새)** 포화 상한 아래면 +1 | `a112_lane_coordinator_projection_test.go` `TestTheCycleGenerationIsTheMarketWaveInWhichTheLaneWasLastObserved`(KR 2 · US 1) | yes — 변이 P04(증가 제거) · P05(공유 계수기) CAUGHT(`analysis/measurements/lot-7.3/mutation-7.3.tsv`) | yes |
| B4 | range at 334:2 — 관측마다 물결 번호를 찍어 덮어쓰기 | `a112_lane_coordinator_projection_test.go` `TestTheCycleGenerationIsTheMarketWaveInWhichTheLaneWasLastObserved` | yes — 편집 전 컴파일 실패 | yes |
