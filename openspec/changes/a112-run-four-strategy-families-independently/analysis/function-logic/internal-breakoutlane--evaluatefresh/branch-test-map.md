# Branch Test Map: `evaluateFresh`

- Source SHA-256: `ab03efc63446a63ab7efdd387c42ec85e1396c606c43c6f8e3d46d8575e18c8f`; AST branch locations are authoritative.
- Revision: **modified (a112 2.3, 2026-10-01).** 편집 전 19 분기 → 20: 범위 뒤 봉 루프(B4) 안, 입장 갈래 B6(입장 시 break) **뒤**에 1.2 반사실 기록 B7 을 더했다(Manager 판정 2.3 (b)). 그래서 편집 전 B7~B19 는 B8~B20 이 되었다(위치 번호). 입장 갈래와 그 안의 1.2 · 2.0 · 2.5 기록 줄은 바이트 그대로다.
- 편집 전 번들: `analysis/measurements/lot-2.3/pre-edit/internal-breakoutlane--evaluatefresh/`. 변이 원장 `analysis/measurements/lot-2.3/mutation-2.3.tsv`.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | range at 40:2 — 범위 봉 순회(저항 · 범위 하단) | `TestEvaluatorDerivesOnlyCompleteClosedBarPathToProposal` | no — 갈래 불변 | yes |
| B2 | if at 41:3 — 범위 봉 high 가 저항을 올림 | `TestEvaluatorDerivesOnlyCompleteClosedBarPathToProposal` · `TestGstackRepairFrozenVocabularyAndV1Thresholds` | no — 갈래 불변 | yes |
| B3 | if at 44:3 — 범위 봉 low 가 하단을 내림 | `TestNoAveragingDownLegAfterAFailedSetup`(범위 하단 아래 종가) | no — 갈래 불변 | yes |
| B4 | range at 52:2 — 범위 뒤 봉 순회 | `TestEvaluatorDerivesOnlyCompleteClosedBarPathToProposal` | no — 갈래 불변 | yes |
| B5 | if at 54:3 — high 만 저항 위(첫 touch) | `TestAdversarialFirstTouchMissingRangeAndBadBarCannotPropose` | no — 갈래 불변 | yes |
| B6 | if at 57:3 — 입장 돌파(close buffer · RVOL >= 1.5 · wick) → 반사실 1.2/2.0/2.5 기록 후 break | `TestRVOLAdmissionAndCounterfactualBoundaries` · `TestTheAdmittedBreakoutPathIsUnchangedByTheCounterfactual` | no — 갈래 · 몸 불변, 변이 CF-10 CAUGHT | yes |
| B7 | if at 70:3 — **(새)** 입장 못 한 봉: close buffer · wick 통과 · 1.2 <= RVOL → `RVOLAt1200000` 기록(기록 전용) | `a112_rvol_counterfactual_test.go` `TestTheOnePointTwoCounterfactualRecordsABarThatOnlyTheLowerThresholdWouldAdmit` | yes — 편집 전 1.2 정확 · 1.5 바로 아래 두 경우 실패(`red-2.3.log`); 변이 CF-01~09 CAUGHT(`analysis/measurements/lot-2.3/mutation-2.3.tsv`) | yes |
| B8 | if at 74:2 — 입장 돌파 없음 → 범위 단계 결정 | `TestRVOLAdmissionAndCounterfactualBoundaries`(1.5 바로 아래) · `a112_rvol_counterfactual_test.go` | no — 갈래 불변 | yes |
| B9 | if at 76:3 — 첫 touch 거절 | `TestAdversarialFirstTouchMissingRangeAndBadBarCannotPropose` | no — 갈래 불변 | yes |
| B10 | if at 84:2 — US 시한 | `TestTimeoutExactBoundaryKRAndUS` | no — 갈래 불변 | yes |
| B11 | for at 87:2 — 돌파 뒤 봉 순회 | `TestEvaluatorDerivesOnlyCompleteClosedBarPathToProposal` | no — 갈래 불변 | yes |
| B12 | if at 90:3 — 범위 하단 아래 종가 → INVALIDATED | `TestNoAveragingDownLegAfterAFailedSetup` · `TestTheObservedBreakoutEdgesPlusTheReservedSixAreTheGoldenSet` | no — 갈래 불변 | yes |
| B13 | if at 93:3 — `since > timeout` — **도달 불가**(B16 이 같음에서 먼저 반환, 스위트 커버리지 0 실측) · 동작 동등 방어 | `TestTheBreakoutTransitionProducersAreExactlyTheCensus`(생산 자리 수로 고정 — BK2-08) | no — 갈래 불변 | 도달 불가 — census 만 |
| B14 | if at 96:3 — retest 뒤 reclaim → RECLAIMED · ARMED | `TestEvaluatorDerivesOnlyCompleteClosedBarPathToProposal` | no — 갈래 불변 | yes |
| B15 | if at 101:3 — 거래량 확장 실패 재돌파 → INVALIDATED | `TestGstackRepairFailedReclaimAndRiskRewardBoundary` · `TestNoAveragingDownLegAfterAFailedSetup` | no — 갈래 불변 | yes |
| B16 | if at 104:3 — 시한 정확 → TIMED_OUT | `TestTimeoutExactBoundaryKRAndUS` | no — 갈래 불변 | yes |
| B17 | if at 107:3 — retest 허용폭 안 | `TestGstackRepairRetestQualifiesToleranceEndpoints` | no — 갈래 불변 | yes |
| B18 | if at 111:2 — ARMED 아님 → 진행 단계 결정 | `TestAdversarialCorrectionReplaysPreTerminalAndPreservesProposed`(RETEST_WAIT) | no — 갈래 불변 | yes |
| B19 | if at 114:2 — 호가 거절 → ARMED + typed refusal | `TestEvaluatorInvalidationTimeoutAndQuoteVetoEdges` · `TestQuoteVetoMatchesTheGoldenFormulasAtAndBeyondEachLimit` | no — 갈래 불변 | yes |
| B20 | if at 118:2 — 사이징 거절 → ARMED + typed refusal | `TestANonProtectiveStopNeverProposes` · `TestSizingMatchesTheExactRationalOracle` | no — 갈래 불변 | yes |
