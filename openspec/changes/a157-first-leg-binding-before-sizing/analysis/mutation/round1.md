# a157 변이 원장 — round 1 (design D6 · tasks 2.2)

- 하네스: `analysis/harness/a157_mutate.py`(원장 원문 `round1.json` · 진행 `round1.log`).
- 대상: 커밋 `1e58d3b56fb67a2154ccb04284433b4ff0a260ed`(GREEN). 사본 `/tmp/claude-1000/-mnt-D-Axipient-workspace-TossOS/1711ae2d-4cc3-48b0-97b1-c0fa1bb7c4da/scratchpad/a157-mut-2093756`(저장소 밖, pid 접미). `GOFLAGS=-trimpath` · 사본 전용 `GOCACHE` · 순차 한 판.
- 시작 sha 단언(사본 = HEAD blob): `riskguardian_first_leg.go` 2edf83d6949f…, `a157_daily_loss_kind.go` cb7d5d6f3b72…, `strategy_first_leg_admission.go` 566cfe6f3d50…, `strategy_entry_supervisor.go` 32f3fe65c0b1…. 변이마다 원래 바이트로 되돌린 뒤 파일 sha 를 다시 단언, 끝에서 넷 모두 재단언.
- 무변이 대조군: 두 시험 명령 모두 GREEN(실패 0, 소요 51.4s · 25.8s) — 그 뒤에만 변이를 셈.
- 닿음(reached): 변이 전후 대상 함수의 본문 sha(`_funcsha`, gofmt 정규화 FuncDecl)가 달라졌는지 — 넷 모두 바뀌어야 「대상에 닿음」.
- 시험 명령(변이마다 둘 다):
  - `go test -count=1 -tags tossos_testseams ./internal/execgw/ -run 'TestA157|TestQFinalCampaignFirstLeg|TestA136QFinalCampaignFirstLeg|TestA138QFinalCampaignFirstLeg|TestA137'`
  - `go test -count=1 -tags tossos_testseams ./internal/app/engine/ -run 'TestA157|TestA137|TestA146|TestStrategyFirstLegAdmission'`
  (태그 빌드는 무태그 시험 파일 — 구조 시험 · 판별 표 · 조립 감독자 — 도 함께 돎.)

## 결과: 31/31 CAUGHT · SURVIVED 0 · BUILD-FAIL 0 · 닿지 않은 변이 0

| 변이 | 내용 | 파일 | 닿음 | 판정 | 실패한 시험(하위 시험 제외) |
|---|---|---|---|---|---|
| M1-sizing-first | M1 결속 블록을 편집 전 순서로(진입 판정 먼저, 그 오류 즉시 반환) | `riskguardian_first_leg.go` | yes | **CAUGHT** | `TestA157CapZeroAndABindingDefect`, `TestA157OneInputThatTripsBothJudgements`, `TestA157OtherEntryRefusalsAndABindingDefect`, `TestA157ThePrecheckJudgesTheBindingOnceAndBeforeSizing`, `TestA157EngineACapZeroNoLongerHidesABindingDefect` |
| M2-move-B2 | M2 B2 만 사이징 뒤로 | `riskguardian_first_leg.go` | yes | **CAUGHT** | `TestA157CapZeroAndABindingDefect`, `TestA157OtherEntryRefusalsAndABindingDefect`, `TestA157ThePrecheckJudgesTheBindingOnceAndBeforeSizing`, `TestA157EngineACapZeroNoLongerHidesABindingDefect` |
| M3-move-B3 | M3 B3 만 사이징 뒤로 | `riskguardian_first_leg.go` | yes | **CAUGHT** | `TestA157CapZeroAndABindingDefect`, `TestA157ThePrecheckJudgesTheBindingOnceAndBeforeSizing` |
| M4-move-B4 | M4 B4 만 사이징 뒤로 | `riskguardian_first_leg.go` | yes | **CAUGHT** | `TestA157ABindingDefectAloneIsReportedAsItself`, `TestA157CapZeroAndABindingDefect`, `TestA157ThePrecheckJudgesTheBindingOnceAndBeforeSizing` |
| M5-move-B5 | M5 B5 만 사이징 뒤로 | `riskguardian_first_leg.go` | yes | **CAUGHT** | `TestA157CapZeroAndABindingDefect`, `TestA157OneInputThatTripsBothJudgements`, `TestA157ThePrecheckJudgesTheBindingOnceAndBeforeSizing` |
| M6-move-B6 | M6 B6 만 사이징 뒤로 | `riskguardian_first_leg.go` | yes | **CAUGHT** | `TestA157CapZeroAndABindingDefect`, `TestA157ThePrecheckJudgesTheBindingOnceAndBeforeSizing` |
| M7-move-B7 | M7 B7 만 사이징 뒤로 | `riskguardian_first_leg.go` | yes | **CAUGHT** | `TestA157CapZeroAndABindingDefect`, `TestA157ThePrecheckJudgesTheBindingOnceAndBeforeSizing` |
| M8-move-B8 | M8 B8 만 사이징 뒤로 | `riskguardian_first_leg.go` | yes | **CAUGHT** | `TestA157CapZeroAndABindingDefect`, `TestA157OtherEntryRefusalsAndABindingDefect`, `TestA157ThePrecheckJudgesTheBindingOnceAndBeforeSizing`, `TestA157EngineACapZeroNoLongerHidesABindingDefect` |
| M9-drop-B2 | M9 B2 거절 삭제 | `riskguardian_first_leg.go` | yes | **CAUGHT** | `TestA157ABindingDefectAloneIsReportedAsItself`, `TestA157CapZeroAndABindingDefect`, `TestA157OtherEntryRefusalsAndABindingDefect`, `TestA157EngineACapZeroNoLongerHidesABindingDefect` |
| M9-drop-B3 | M9 B3 거절 삭제 | `riskguardian_first_leg.go` | yes | **CAUGHT** | `TestA157ABindingDefectAloneIsReportedAsItself`, `TestA157CapZeroAndABindingDefect` |
| M9-drop-B4 | M9 B4 거절 삭제 | `riskguardian_first_leg.go` | yes | **CAUGHT** | `TestA157ABindingDefectAloneIsReportedAsItself`, `TestA157CapZeroAndABindingDefect` |
| M9-drop-B5 | M9 B5 거절 삭제 | `riskguardian_first_leg.go` | yes | **CAUGHT** | `TestA157ABindingDefectAloneIsReportedAsItself`, `TestA157CapZeroAndABindingDefect`, `TestA157OneInputThatTripsBothJudgements`, `TestQFinalCampaignFirstLegRejectsTamperWithoutWritesPairedKRUS` |
| M9-drop-B6 | M9 B6 거절 삭제 | `riskguardian_first_leg.go` | yes | **CAUGHT** | `TestA157ABindingDefectAloneIsReportedAsItself`, `TestA157CapZeroAndABindingDefect`, `TestQFinalCampaignFirstLegRejectsTamperWithoutWritesPairedKRUS` |
| M9-drop-B7 | M9 B7 거절 삭제 | `riskguardian_first_leg.go` | yes | **CAUGHT** | `TestA157ABindingDefectAloneIsReportedAsItself`, `TestA157CapZeroAndABindingDefect` |
| M9-drop-B8 | M9 B8 거절 삭제 | `riskguardian_first_leg.go` | yes | **CAUGHT** | `TestA157ABindingDefectAloneIsReportedAsItself`, `TestA157CapZeroAndABindingDefect`, `TestA157OtherEntryRefusalsAndABindingDefect`, `TestQFinalCampaignFirstLegWeeklyBindingPairedKRUS`, `TestA157EngineABindingDefectAloneIsAnAuthorityMismatch`, `TestA157EngineACapZeroNoLongerHidesABindingDefect` |
| M10-drop-nil-guard | M10 nil 가드 삭제 | `riskguardian_first_leg.go` | yes | **CAUGHT** | `TestA157ANilGuardianRefusesBeforeAnyBindingJudgement`, `panic` |
| M11-B5-qcandidate-self | M11 B5 q_candidate 를 Result.Quantity 자기 비교로 | `riskguardian_first_leg.go` | yes | **CAUGHT** | `TestA157ABindingDefectAloneIsReportedAsItself`, `TestA157CapZeroAndABindingDefect`, `TestA157OneInputThatTripsBothJudgements`, `TestQFinalCampaignFirstLegRejectsTamperWithoutWritesPairedKRUS` |
| M12a-B5-lane-self | M12 B5 소유자 LaneID 를 계보 자기 비교로 | `riskguardian_first_leg.go` | yes | **CAUGHT** | `TestA157ABindingDefectAloneIsReportedAsItself`, `TestA157CapZeroAndABindingDefect` |
| M12b-B5-campaign-self | M12 B5 소유자 CampaignID 를 계보 자기 비교로 | `riskguardian_first_leg.go` | yes | **CAUGHT** | `TestA157ABindingDefectAloneIsReportedAsItself`, `TestA157CapZeroAndABindingDefect` |
| M13-drop-daily-loss-precedence | M13 ④의 일일 손실 예외 삭제(결속 결함 무조건 우선) | `riskguardian_first_leg.go` | yes | **CAUGHT** | `TestA157DailyLossIsReportedBeforeABindingDefect`, `TestA157EngineADailyLossLimitStillEscalatesWithABindingDefect`, `TestA157EngineAnUnmeasurableDayStillWinsOverABindingDefect` |
| M13b-invert-daily-loss-test | (추가) ④의 판별 비교 반전(!= → ==) | `riskguardian_first_leg.go` | yes | **CAUGHT** | `TestA157CapZeroAndABindingDefect`, `TestA157DailyLossIsReportedBeforeABindingDefect`, `TestA157OneInputThatTripsBothJudgements`, `TestA157OtherEntryRefusalsAndABindingDefect`, `TestA157EngineACapZeroNoLongerHidesABindingDefect`, `TestA157EngineADailyLossLimitStillEscalatesWithABindingDefect`, `TestA157EngineAnUnmeasurableDayStillWinsOverABindingDefect` |
| M14a-kind-drop-unmeasurable | M14 판별 함수 측정 불가 갈래 삭제 | `a157_daily_loss_kind.go` | yes | **CAUGHT** | `TestA157DailyLossChainRefusalOfRealRefusals`, `TestA157DailyLossChainRefusalOfTable`, `TestA157DailyLossIsReportedBeforeABindingDefect`, `TestA137AnUnmeasurableDayIsAPolicyOutcomeNotALaneFault`, `TestA157EngineAnUnmeasurableDayStillWinsOverABindingDefect` |
| M14b-kind-drop-limit | M14 판별 함수 한도 갈래 삭제 | `a157_daily_loss_kind.go` | yes | **CAUGHT** | `TestA157DailyLossChainRefusalOfRealRefusals`, `TestA157DailyLossChainRefusalOfTable`, `TestA157DailyLossIsReportedBeforeABindingDefect`, `TestA137ADailyLossRefusalEscalatesAndIsNotALaneFault`, `TestA137AFailedEscalationIsJoinedNotLost`, `TestA137ClassificationIsUnchangedForEveryPrecheckRefusal`, `TestA157EngineADailyLossLimitStillEscalatesWithABindingDefect` |
| M14c-kind-unmeasurable-any-stage | (추가) 측정 불가 갈래의 체인 단계 조건 삭제 | `a157_daily_loss_kind.go` | yes | **CAUGHT** | `TestA157DailyLossChainRefusalOfTable` |
| M14d-kind-unmeasurable-without-marker | (추가) 측정 불가 갈래의 표지 조건 삭제 | `a157_daily_loss_kind.go` | yes | **CAUGHT** | `TestA157DailyLossChainRefusalOfRealRefusals`, `TestA157DailyLossChainRefusalOfTable`, `TestA157OtherEntryRefusalsAndABindingDefect`, `TestA137AnUnmeasurableDayIsAPolicyOutcomeNotALaneFault`, `TestA137ClassificationIsUnchangedForEveryPrecheckRefusal`, `TestA146ACapZeroSkipHidesTheChecksAfterSizingUntilTheCapRises` |
| M15-duplicate-B3-after-sizing | M15 결속 판정(B3)을 앞에 두고 사이징 뒤에도 복제 | `riskguardian_first_leg.go` | yes | **CAUGHT** | `TestA157ThePrecheckJudgesTheBindingOnceAndBeforeSizing` |
| M16a-entrygate-no-nil-guard | M16 EntryGate nil 가드 삭제(타입 있는 nil 대입) | `strategy_entry_supervisor.go` | yes | **CAUGHT** | `TestA157TheAssemblySupervisorEntryGate` |
| M16b-drop-daily-loss-notice | M16 DailyLossNotice 칸 삭제 | `strategy_entry_supervisor.go` | yes | **CAUGHT** | `TestA157TheAssemblyNoticesReachTheLog`, `TestA157TheAssemblySupervisorCarriesBothNotices` |
| M16c-drop-entry-skip-notice | M16 EntrySkipNotice 칸 삭제 | `strategy_entry_supervisor.go` | yes | **CAUGHT** | `TestA157TheAssemblyNoticesReachTheLog`, `TestA157TheAssemblySupervisorCarriesBothNotices` |
| M16d-drop-entrygate | (추가) EntryGate 배선 자체 삭제 | `strategy_entry_supervisor.go` | yes | **CAUGHT** | `TestA157TheAssemblySupervisorEntryGate` |
| M17-admit-ignores-classifier | (추가) admit 이 판별 함수를 무시(일일 손실 갈래 둘 다 꺼짐) | `strategy_first_leg_admission.go` | yes | **CAUGHT** | `TestA137ADailyLossRefusalEscalatesAndIsNotALaneFault`, `TestA137AFailedEscalationIsJoinedNotLost`, `TestA137AnUnmeasurableDayIsAPolicyOutcomeNotALaneFault`, `TestA137ClassificationIsUnchangedForEveryPrecheckRefusal`, `TestA157EngineADailyLossLimitStillEscalatesWithABindingDefect`, `TestA157EngineAnUnmeasurableDayStillWinsOverABindingDefect` |

## 읽는 법 · 메모

- **부분 이동(M2~M8)은 행동 시험만으로도 잡힘** — 각 변이에서 `TestA157CapZeroAndABindingDefect`(상한 0 + 그 Bn 행)가 실패함. 구조 시험
  `TestA157ThePrecheckJudgesTheBindingOnceAndBeforeSizing` 도 함께 실패하지만, 구조 시험 없이도 CAUGHT(freeze 리뷰 P1-3 의 「부분 이동」 구멍이 닫힘).
- **M15(복제)는 구조 시험만 잡음** — 설계대로(D6): 앞뒤 두 자리의 같은 판정은 행동으로 구별되지 않으므로 AST 셈이 못임.
  **(round2 정정, 구현 리뷰 P2)** round1 의 구조 시험은 **리터럴을 옮겨 적는** 복제만 잡았음 — 이름 붙인 닫힌 함수를 앞뒤로 두 번 부르는 모양(M15b,
  리뷰어 실측 SURVIVED)과 유일한 호출을 사이징 뒤로 미루는 모양(M15c)은 리터럴 개수 · 위치가 그대로라 통과했음. 구조 시험을 「FuncLit 하나 ·
  즉시 호출 하나 · 최상위 대입 · 진입 판정 문장보다 앞」으로 강화한 뒤 `round2.md` 에서 둘 다(+M15d 반복문 안 호출) CAUGHT.
- **M14c(측정 불가 갈래의 단계 조건 삭제)는 판별 표 시험만 잡음.** 생산의 측정 불가 표지는 `dailyLossChainRefusal` 이 체인 단계 거절에만 붙이므로
  (a137_daily_loss.go:203-209) 실제 Guardian 경로에서는 행동이 같음 — 표 시험의 합성 행(「다른 단계의 입력 불가 + 표지」)이 조건을 못 박음. 비동등으로 셈.
- **M16a(타입 있는 nil)는 `TestA157TheAssemblySupervisorEntryGate` 의 「게이트 없는 엔진 → nil 인터페이스」 단언이 잡음**(Block 의 panic 까지 가지 않고 칸에서 멈춤).
- **M10(nil 가드 삭제)은 panic 으로 잡힘** — 결속 판정이 `g.policyVersion` 을 읽는 자리에서 nil 역참조. 가드가 있으면 거절(`RISK_CALCULATION_INVALID/guardian`).
- 동등 변이: **M14c 는 생산 경로에서 동등 — 합성 표 행으로만 잡힘**(round2 정정, 구현 리뷰 Opus P3-5). 표지 생성 자리는
  `dailyLossChainRefusal` 하나이고 그 거절은 늘 체인 단계라, 생산이 만들 수 있는 오류에서는 행동이 같음. 표 행은 조건을 못 박으려 유지함.
  그 밖의 30 개는 비동등.
- 빌드 메모(변이 아님, freeze 리뷰 P3-6): 「B9 계보 투영을 사이징 앞으로」는 `entry.decision.QFinal` · `entry.intent` 가 없어 컴파일되지 않음 — 시험의 증거가 아니라 구조의 증거.
