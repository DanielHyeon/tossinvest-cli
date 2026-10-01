# Branch Test Map: `strategyLaneRuntime.evaluate`

- Source SHA-256: `95611377145456c4906da5ebb95eca97dd025df0c1fb255ebc443c1ce12d5d36`; AST branch locations are authoritative.
- Revision: **modified (a112 7.5, 2026-10-01).** 편집 전 6 분기 → 8: 레인 순회(B3)가 레인마다 goroutine 하나를 띄우고(go 문 1 · defer 2 — join.Done · recover) join 한 뒤, B6 · B7(레인 goroutine 의 panic 을 시장 주기 goroutine 에서 다시 던짐)을 더했다. 관측은 레인 순서 색인으로 모음. 나머지 분기 불변.
- 편집 전 번들: `analysis/measurements/lot-7.5/pre-edit/internal-app-engine--strategylaneruntime.evaluate/`. 변이 원장 `analysis/measurements/lot-7.5/mutation-7.5.tsv`.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 202:2 — nil 런타임 | 진입 0(편집 전 번들 기록 그대로) | no — 이 로트가 바꾸지 않음 | n/a |
| B2 | if at 208:2 — 복구 실패 → 오류 | 진입 0(편집 전 번들 기록 그대로) | no — 이 로트가 바꾸지 않음 | n/a |
| B3 | range at 228:2 — **(편집)** 레인마다 goroutine 하나 · join(시장 주기 안) — 멈춘 레인이 이웃을 세우지 않음, 시장 지연 = 최댓값 | `a112_lane_latency_testseam_test.go` `TestAHungLaneDoesNotDelayItsPeersInTheSameWave`(경과 = 마감 시한 정확히 1 회, 이웃은 가상 시각 0 에 반환) | yes — 편집 전 컴파일 실패(`red-7.5.log` — seam 부재), 변이 R01(순차 되돌림) · R04(join 제거) CAUGHT(`analysis/measurements/lot-7.5/mutation-7.5.tsv`) | yes |
| B4 | range at 230:3 — 레인 입력 찾기 | `a112_lane_fanout_test.go` `TestEightLanesShareOneAuthorityWaveAndEachProposalReachesOneLane` | no — 이 로트가 바꾸지 않음 | yes |
| B5 | if at 231:4 — 자기 제안 소유 — 제안 하나 → 레인 하나 | `a112_lane_fanout_test.go` `TestEightLanesShareOneAuthorityWaveAndEachProposalReachesOneLane` | no — 분기 불변, 변이 R08(모든 레인에 줌) CAUGHT(`analysis/measurements/lot-7.5/mutation-7.5.tsv`) | yes |
| B6 | range at 244:2 — **(새)** 레인 goroutine 들의 panic 값 순회 | `a112_lane_latency_testseam_test.go` `TestAPanicOutsideALaneStepStillReachesTheMarketCycle` | yes — 편집 전 컴파일 실패, 변이 R02 · R03 CAUGHT(`analysis/measurements/lot-7.5/mutation-7.5.tsv`) | yes |
| B7 | if at 245:3 — **(새)** panic 이 있으면 시장 주기 goroutine 에서 다시 던짐(순차 때와 같은 회복 경로) | `a112_lane_latency_testseam_test.go` `TestAPanicOutsideALaneStepStillReachesTheMarketCycle` | yes — 변이 R02(삼킴) CAUGHT | yes |
| B8 | if at 252:2 — 잠금 기록 실패 → 오류 | `TestALaneLatchThatCannotBeRecordedIsCountedNotEscalated` | no — 이 로트가 바꾸지 않음 | yes |
