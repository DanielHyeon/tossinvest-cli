# Branch Test Map: `strategyLaneRuntime.evaluate`

- Source SHA-256: `4a7fd7fedb3237720070c6c4c6ef03030fa30c67e181fdb0a86053a8418390a6`; AST branch locations are authoritative.
- Revision: **modified (a112 7.3, 2026-10-01).** 분기 불변(6). `record(observations)` → `record(market, observations)` 한 줄(시장별 물결 번호 — 판정 Q1=(B)). 순서(복구 → 돌기 → 기록 → 잠금 남기기) 불변.
- 편집 전 번들: `analysis/measurements/lot-7.3/pre-edit/internal-app-engine--strategylaneruntime.evaluate/`. 변이 원장 `analysis/measurements/lot-7.3/mutation-7.3.tsv`.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 202:2 — nil 런타임 | 진입 0(편집 전 번들 기록 그대로) | no — 이 로트가 바꾸지 않음 | n/a |
| B2 | if at 208:2 — 복구 실패 → 오류 | 진입 0(편집 전 번들 기록 그대로) | no — 이 로트가 바꾸지 않음 | n/a |
| B3 | range at 213:2 — 이 시장 네 레인 순회 | `a112_lane_coordinator_projection_test.go` `TestTheCycleGenerationIsTheMarketWaveInWhichTheLaneWasLastObserved` | no — 이 로트가 바꾸지 않음 | yes |
| B4 | range at 215:3 — 레인 입력 찾기 | `TestAPromotedLaneAdmitsItsFamilyWhileAnUnpromotedOneStopsIt` | no — 이 로트가 바꾸지 않음 | yes |
| B5 | if at 216:4 — 자기 제안 소유 | `TestAPromotedLaneAdmitsItsFamilyWhileAnUnpromotedOneStopsIt` | no — 이 로트가 바꾸지 않음 | yes |
| B6 | if at 226:2 — 잠금 기록 실패 → 오류 | `TestALaneLatchThatCannotBeRecordedIsCountedNotEscalated` | no — 이 로트가 바꾸지 않음 | yes |
