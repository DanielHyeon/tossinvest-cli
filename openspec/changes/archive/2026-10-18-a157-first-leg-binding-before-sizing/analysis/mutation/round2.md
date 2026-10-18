# a157 변이 원장 — round 2 (구현 리뷰 2.3 수리 뒤)

- 대상: 커밋 `b8d6d4cee568b6c05133910810b3cfe9871860c8`(2.3 수리). 사본 `/tmp/claude-1000/-mnt-D-Axipient-workspace-TossOS/1711ae2d-4cc3-48b0-97b1-c0fa1bb7c4da/scratchpad/a157-mut-114302`(저장소 밖, pid 접미). 하네스 · 규율은 round1 과 같음(`analysis/harness/a157_mutate.py` —
  시작 sha 단언 `riskguardian_first_leg.go` d37b2fcb0187…, `a157_daily_loss_kind.go` 6196fb379d12…, `strategy_first_leg_admission.go` 2ef5d3da61a8…, `strategy_entry_supervisor.go` 32f3fe65c0b1… · 변이마다 복원 sha 단언 · 무변이 대조군 GREEN
  27.3s · 13.7s · 대상 함수 본문 sha 변화로 닿음 확인 · 순차 한 판).
- 왜 다시 돌렸나: 구조 시험 강화(P2), 판별 이름 변경(P3-2), 봉인 값 시험 강화(P3-4 · codex P3-2), 엔진 강화 실패 행(P3-3)이 시험 · 생산 이름을 바꿨으므로
  round1 의 31 개를 모두 다시 재고, 리뷰가 지적한 모양 셋(M15b · M15c · M15d)과 봉인 대입 삭제 둘(M18 · M19)을 더함.
- 원문: `round2.json` · `round2.log`.

## 결과: 36/36 CAUGHT · SURVIVED 0 · BUILD-FAIL 0 · 닿지 않은 변이 0

| 변이 | 내용 | 파일 | 닿음 | 판정 | 실패한 시험(하위 시험 제외) |
|---|---|---|---|---|---|
| M1-sizing-first | round1 과 같은 변이(판별 이름 변경 반영) | `riskguardian_first_leg.go` | yes | **CAUGHT** | `TestA157CapZeroAndABindingDefect`, `TestA157OneInputThatTripsBothJudgements`, `TestA157OtherEntryRefusalsAndABindingDefect`, `TestA157ThePrecheckJudgesTheBindingOnceAndBeforeSizing`, `TestA157EngineACapZeroNoLongerHidesABindingDefect` |
| M2-move-B2 | round1 과 같은 변이(판별 이름 변경 반영) | `riskguardian_first_leg.go` | yes | **CAUGHT** | `TestA157CapZeroAndABindingDefect`, `TestA157OtherEntryRefusalsAndABindingDefect`, `TestA157ThePrecheckJudgesTheBindingOnceAndBeforeSizing`, `TestA157EngineACapZeroNoLongerHidesABindingDefect` |
| M3-move-B3 | round1 과 같은 변이(판별 이름 변경 반영) | `riskguardian_first_leg.go` | yes | **CAUGHT** | `TestA157CapZeroAndABindingDefect`, `TestA157ThePrecheckJudgesTheBindingOnceAndBeforeSizing` |
| M4-move-B4 | round1 과 같은 변이(판별 이름 변경 반영) | `riskguardian_first_leg.go` | yes | **CAUGHT** | `TestA157ABindingDefectAloneIsReportedAsItself`, `TestA157CapZeroAndABindingDefect`, `TestA157ThePrecheckJudgesTheBindingOnceAndBeforeSizing` |
| M5-move-B5 | round1 과 같은 변이(판별 이름 변경 반영) | `riskguardian_first_leg.go` | yes | **CAUGHT** | `TestA157CapZeroAndABindingDefect`, `TestA157OneInputThatTripsBothJudgements`, `TestA157ThePrecheckJudgesTheBindingOnceAndBeforeSizing` |
| M6-move-B6 | round1 과 같은 변이(판별 이름 변경 반영) | `riskguardian_first_leg.go` | yes | **CAUGHT** | `TestA157CapZeroAndABindingDefect`, `TestA157ThePrecheckJudgesTheBindingOnceAndBeforeSizing` |
| M7-move-B7 | round1 과 같은 변이(판별 이름 변경 반영) | `riskguardian_first_leg.go` | yes | **CAUGHT** | `TestA157CapZeroAndABindingDefect`, `TestA157ThePrecheckJudgesTheBindingOnceAndBeforeSizing` |
| M8-move-B8 | round1 과 같은 변이(판별 이름 변경 반영) | `riskguardian_first_leg.go` | yes | **CAUGHT** | `TestA157CapZeroAndABindingDefect`, `TestA157OtherEntryRefusalsAndABindingDefect`, `TestA157ThePrecheckJudgesTheBindingOnceAndBeforeSizing`, `TestA157EngineACapZeroNoLongerHidesABindingDefect` |
| M9-drop-B2 | round1 과 같은 변이(판별 이름 변경 반영) | `riskguardian_first_leg.go` | yes | **CAUGHT** | `TestA157ABindingDefectAloneIsReportedAsItself`, `TestA157CapZeroAndABindingDefect`, `TestA157OtherEntryRefusalsAndABindingDefect`, `TestA157EngineACapZeroNoLongerHidesABindingDefect` |
| M9-drop-B3 | round1 과 같은 변이(판별 이름 변경 반영) | `riskguardian_first_leg.go` | yes | **CAUGHT** | `TestA157ABindingDefectAloneIsReportedAsItself`, `TestA157CapZeroAndABindingDefect` |
| M9-drop-B4 | round1 과 같은 변이(판별 이름 변경 반영) | `riskguardian_first_leg.go` | yes | **CAUGHT** | `TestA157ABindingDefectAloneIsReportedAsItself`, `TestA157CapZeroAndABindingDefect` |
| M9-drop-B5 | round1 과 같은 변이(판별 이름 변경 반영) | `riskguardian_first_leg.go` | yes | **CAUGHT** | `TestA157ABindingDefectAloneIsReportedAsItself`, `TestA157CapZeroAndABindingDefect`, `TestA157OneInputThatTripsBothJudgements`, `TestQFinalCampaignFirstLegRejectsTamperWithoutWritesPairedKRUS` |
| M9-drop-B6 | round1 과 같은 변이(판별 이름 변경 반영) | `riskguardian_first_leg.go` | yes | **CAUGHT** | `TestA157ABindingDefectAloneIsReportedAsItself`, `TestA157CapZeroAndABindingDefect`, `TestQFinalCampaignFirstLegRejectsTamperWithoutWritesPairedKRUS` |
| M9-drop-B7 | round1 과 같은 변이(판별 이름 변경 반영) | `riskguardian_first_leg.go` | yes | **CAUGHT** | `TestA157ABindingDefectAloneIsReportedAsItself`, `TestA157CapZeroAndABindingDefect` |
| M9-drop-B8 | round1 과 같은 변이(판별 이름 변경 반영) | `riskguardian_first_leg.go` | yes | **CAUGHT** | `TestA157ABindingDefectAloneIsReportedAsItself`, `TestA157CapZeroAndABindingDefect`, `TestA157OtherEntryRefusalsAndABindingDefect`, `TestQFinalCampaignFirstLegWeeklyBindingPairedKRUS`, `TestA157EngineABindingDefectAloneIsAnAuthorityMismatch`, `TestA157EngineACapZeroNoLongerHidesABindingDefect` |
| M10-drop-nil-guard | round1 과 같은 변이(판별 이름 변경 반영) | `riskguardian_first_leg.go` | yes | **CAUGHT** | `TestA157ANilGuardianRefusesBeforeAnyBindingJudgement`, `panic` |
| M11-B5-qcandidate-self | round1 과 같은 변이(판별 이름 변경 반영) | `riskguardian_first_leg.go` | yes | **CAUGHT** | `TestA157ABindingDefectAloneIsReportedAsItself`, `TestA157CapZeroAndABindingDefect`, `TestA157OneInputThatTripsBothJudgements`, `TestQFinalCampaignFirstLegRejectsTamperWithoutWritesPairedKRUS` |
| M12a-B5-lane-self | round1 과 같은 변이(판별 이름 변경 반영) | `riskguardian_first_leg.go` | yes | **CAUGHT** | `TestA157ABindingDefectAloneIsReportedAsItself`, `TestA157CapZeroAndABindingDefect` |
| M12b-B5-campaign-self | round1 과 같은 변이(판별 이름 변경 반영) | `riskguardian_first_leg.go` | yes | **CAUGHT** | `TestA157ABindingDefectAloneIsReportedAsItself`, `TestA157CapZeroAndABindingDefect` |
| M13-drop-daily-loss-precedence | round1 과 같은 변이(판별 이름 변경 반영) | `riskguardian_first_leg.go` | yes | **CAUGHT** | `TestA157DailyLossIsReportedBeforeABindingDefect`, `TestA157EngineADailyLossLimitStillEscalatesWithABindingDefect`, `TestA157EngineAFailedEscalationIsTheSameJoinedFaultWithABindingDefect`, `TestA157EngineAnUnmeasurableDayStillWinsOverABindingDefect` |
| M13b-invert-daily-loss-test | round1 과 같은 변이(판별 이름 변경 반영) | `riskguardian_first_leg.go` | yes | **CAUGHT** | `TestA157CapZeroAndABindingDefect`, `TestA157DailyLossIsReportedBeforeABindingDefect`, `TestA157OneInputThatTripsBothJudgements`, `TestA157OtherEntryRefusalsAndABindingDefect`, `TestA157EngineACapZeroNoLongerHidesABindingDefect`, `TestA157EngineADailyLossLimitStillEscalatesWithABindingDefect`, `TestA157EngineAFailedEscalationIsTheSameJoinedFaultWithABindingDefect`, `TestA157EngineAnUnmeasurableDayStillWinsOverABindingDefect` |
| M14a-kind-drop-unmeasurable | round1 과 같은 변이(판별 이름 변경 반영) | `a157_daily_loss_kind.go` | yes | **CAUGHT** | `TestA157DailyLossIsReportedBeforeABindingDefect`, `TestA157DailyLossRefusalKindOfRealRefusals`, `TestA157DailyLossRefusalKindOfTable`, `TestA137AnUnmeasurableDayIsAPolicyOutcomeNotALaneFault`, `TestA157EngineAnUnmeasurableDayStillWinsOverABindingDefect` |
| M14b-kind-drop-limit | round1 과 같은 변이(판별 이름 변경 반영) | `a157_daily_loss_kind.go` | yes | **CAUGHT** | `TestA157DailyLossIsReportedBeforeABindingDefect`, `TestA157DailyLossRefusalKindOfRealRefusals`, `TestA157DailyLossRefusalKindOfTable`, `TestA137ADailyLossRefusalEscalatesAndIsNotALaneFault`, `TestA137AFailedEscalationIsJoinedNotLost`, `TestA137ClassificationIsUnchangedForEveryPrecheckRefusal`, `TestA157EngineADailyLossLimitStillEscalatesWithABindingDefect`, `TestA157EngineAFailedEscalationIsTheSameJoinedFaultWithABindingDefect` |
| M14c-kind-unmeasurable-any-stage | round1 과 같은 변이(판별 이름 변경 반영) | `a157_daily_loss_kind.go` | yes | **CAUGHT** | `TestA157DailyLossRefusalKindOfTable` |
| M14d-kind-unmeasurable-without-marker | round1 과 같은 변이(판별 이름 변경 반영) | `a157_daily_loss_kind.go` | yes | **CAUGHT** | `TestA157DailyLossRefusalKindOfRealRefusals`, `TestA157DailyLossRefusalKindOfTable`, `TestA157OtherEntryRefusalsAndABindingDefect`, `TestA137AnUnmeasurableDayIsAPolicyOutcomeNotALaneFault`, `TestA137ClassificationIsUnchangedForEveryPrecheckRefusal`, `TestA146ACapZeroSkipHidesTheChecksAfterSizingUntilTheCapRises` |
| M15-duplicate-B3-after-sizing | round1 과 같은 변이(판별 이름 변경 반영) | `riskguardian_first_leg.go` | yes | **CAUGHT** | `TestA157ThePrecheckJudgesTheBindingOnceAndBeforeSizing` |
| M15b-named-closure-called-twice | (round2) 결속 닫힌 함수에 이름을 붙여 사이징 앞에서 부르고, issuedAt 앞에서 한 번 더 부름(리뷰어 실측 SURVIVED 모양) | `riskguardian_first_leg.go` | yes | **CAUGHT** | `TestA157ThePrecheckJudgesTheBindingOnceAndBeforeSizing` |
| M15c-sole-invocation-after-sizing | (round2) 이름 붙인 닫힌 함수의 유일한 호출을 진입 판정 뒤로 미룸(codex 지적 모양) | `riskguardian_first_leg.go` | yes | **CAUGHT** | `TestA157ThePrecheckJudgesTheBindingOnceAndBeforeSizing` |
| M15d-invocation-inside-a-loop | (round2) 즉시 호출을 `for range 1` 안으로(최상위 문장이 아님) | `riskguardian_first_leg.go` | yes | **CAUGHT** | `TestA157ThePrecheckJudgesTheBindingOnceAndBeforeSizing` |
| M18-drop-policy-version-assignment | (round2) 결속 판정이 봉인에 넘기는 policyVersion 대입 삭제 | `riskguardian_first_leg.go` | yes | **CAUGHT** | `TestA136QFinalCampaignFirstLegRefusesTheBuyThatReachesTheDailyCeiling`, `TestA138QFinalCampaignFirstLegRefusesTheEntryThatExceedsTheMaximum`, `TestA157AnAcceptedFirstLegSealsTheSameValues`, `TestQFinalCampaignFirstLegIssueRefusalsWriteZeroPairedKRUS`, `TestQFinalCampaignFirstLegProjectsGuardianOwnedLineageAllEightLanes`, `TestQFinalCampaignFirstLegWeeklyBindingPairedKRUS`, `TestA146ANotionalCapZeroIsTheSameSkipOnTheRealPath`, `TestA146CharacterizationABucketExhaustedScopeStillLatchesAnEffectiveMarket` |
| M19-drop-weekly-assignment | (round2) 결속 판정이 봉인에 넘기는 weekly 대입 삭제 | `riskguardian_first_leg.go` | yes | **CAUGHT** | `TestA136QFinalCampaignFirstLegRefusesTheBuyThatReachesTheDailyCeiling`, `TestA138QFinalCampaignFirstLegRefusesTheEntryThatExceedsTheMaximum`, `TestA157AnAcceptedFirstLegSealsTheSameValues`, `TestQFinalCampaignFirstLegProjectsGuardianOwnedLineageAllEightLanes`, `TestQFinalCampaignFirstLegWeeklyBindingPairedKRUS` |
| M16a-entrygate-no-nil-guard | round1 과 같은 변이(판별 이름 변경 반영) | `strategy_entry_supervisor.go` | yes | **CAUGHT** | `TestA157TheAssemblySupervisorEntryGate` |
| M16b-drop-daily-loss-notice | round1 과 같은 변이(판별 이름 변경 반영) | `strategy_entry_supervisor.go` | yes | **CAUGHT** | `TestA157TheAssemblyNoticesReachTheLog`, `TestA157TheAssemblySupervisorCarriesBothNotices` |
| M16c-drop-entry-skip-notice | round1 과 같은 변이(판별 이름 변경 반영) | `strategy_entry_supervisor.go` | yes | **CAUGHT** | `TestA157TheAssemblyNoticesReachTheLog`, `TestA157TheAssemblySupervisorCarriesBothNotices` |
| M16d-drop-entrygate | round1 과 같은 변이(판별 이름 변경 반영) | `strategy_entry_supervisor.go` | yes | **CAUGHT** | `TestA157TheAssemblySupervisorEntryGate` |
| M17-admit-ignores-classifier | round1 과 같은 변이(판별 이름 변경 반영) | `strategy_first_leg_admission.go` | yes | **CAUGHT** | `TestA137ADailyLossRefusalEscalatesAndIsNotALaneFault`, `TestA137AFailedEscalationIsJoinedNotLost`, `TestA137AnUnmeasurableDayIsAPolicyOutcomeNotALaneFault`, `TestA137ClassificationIsUnchangedForEveryPrecheckRefusal`, `TestA157EngineADailyLossLimitStillEscalatesWithABindingDefect`, `TestA157EngineAFailedEscalationIsTheSameJoinedFaultWithABindingDefect`, `TestA157EngineAnUnmeasurableDayStillWinsOverABindingDefect` |

## 메모

- **M15 · M15b · M15c · M15d 는 구조 시험 `TestA157ThePrecheckJudgesTheBindingOnceAndBeforeSizing` 만 잡음** — 네 모양 모두 행동이 같으므로 이것이 의도한 못임.
  round1 의 「유일한 못」 문장은 리터럴 복제(M15)에만 참이었고 round1.md 에 정정을 적었음.
- **M18 · M19 는 강화된 `TestA157AnAcceptedFirstLegSealsTheSameValues` 와 기존 첫 다리 시험들이 함께 잡음** — 봉인 시험이 이제 고정 시계에서
  전략 계획 · 캠페인 · 주간 사본 · issuedAt 을 정확한 값과 대조함.
- **M13 은 새 엔진 행 `TestA157EngineAFailedEscalationIsTheSameJoinedFaultWithABindingDefect` 로도 잡힘**(강화 실패 갈래).
- **M14c 는 생산 경로에서 동등 — 합성 표 행으로만 잡힘**(round1 정정과 같음). 그 밖의 35 개는 비동등.
