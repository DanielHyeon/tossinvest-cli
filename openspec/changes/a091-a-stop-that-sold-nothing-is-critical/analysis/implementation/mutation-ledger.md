# a091 변이 원장

하네스 `analysis/harness/mutate.py` · 분리 worktree(`/tmp/claude-1000/a091-mut`) · 무변이 대조군 GREEN 먼저 · 치환 원문 1회 단언 · 원복 바이트 대조 · `GOFLAGS=-trimpath`.

## 1판 — 커밋 `353623a6`(25 변이 — 2판까지의 기록이 「24」로 적은 것은 오기, 구현 리뷰 i1 보이스 B #8)

| id | 변이 | 패키지 · -run | 결과 |
|---|---|---|---|
| C | 무변이 대조군 5 조합 | — | GREEN |
| M1 | 새 종류 등록 제거 | `./internal/obs/` `TestA091` | CAUGHT — `TestA091TheStopSoldNothingKindIsCritical`, `TestA091NoOtherGradeMoved` |
| M1e | 새 종류 등록 제거(엔진 쪽 관측) | `./internal/app/engine/` `TestA091AProtectiveZero` | CAUGHT — `TestA091AProtectiveZeroIsACriticalRow` |
| M2 | submit 이 보호 여부를 거짓으로 | `./internal/app/engine/` `TestA091` | CAUGHT — `TestA091TheAugustSecondReplay`, `TestA091TheReminderWindowDecidesTheNextEpisode`, `TestA091TheReportFitsItsShare` |
| M3 | submit 이 보호 여부를 늘 참으로 | `./internal/app/engine/` `TestA091` | CAUGHT — `TestA091ATakeProfitThatSoldNothingKeepsItsGrade` |
| M4 | 알림 켜짐 게이트 제거 | `./internal/app/engine/` `TestA091` | CAUGHT — `TestA091TheAugustSecondReplay`, `TestA091WithAlertsOffNothingChangesButTheWording`, `TestA091TheLogAndTheAlertAreOneKind` |
| M5 | 생산 배선의 설정 덮기 제거 | `./internal/app/engine/` `TestA091TheProductionAssembly` | CAUGHT — `TestA091TheProductionAssemblyReadsTheLoadedSwitch` |
| M6 | 보유 0 배제 제거 | `./internal/app/engine/` `TestA091` | CAUGHT — `TestA091AZeroHoldingIsNotAFailedStop` |
| M7 | 합쳐진 오류에서 잎 하나만 취소여도 억제(errors.Is 의미) | `./internal/app/engine/` `TestA091` | CAUGHT — `TestA091OnlyACancellationOnlyFailureIsSuppressed`, `TestA091TheCancellationPredicate` |
| M7b | 취소 판정에서 ctx 확인 제거 | `./internal/app/engine/` `TestA091` | **SURVIVED** |
| M8 | 취소 판정 늘 거짓 | `./internal/app/engine/` `TestA091` | CAUGHT — `TestA091OnlyACancellationOnlyFailureIsSuppressed`, `TestA091TheCancellationPredicate` |
| M9 | 기록에서 WithoutCancel 제거 | `./internal/app/engine/` `TestA091` | CAUGHT — `TestA091OnlyACancellationOnlyFailureIsSuppressed`, `TestA091ShutdownWaitsForTheReportWithoutADeadline` |
| M10 | B2 critical 알림 빠뜨림 | `./internal/app/engine/` `TestA091` | CAUGHT — `TestA091AFloorThatCannotBeComputedIsTheSameReport`, `TestA091OnlyACancellationOnlyFailureIsSuppressed`, `TestA091ShutdownWaitsForTheReportWithoutADeadline` |
| M11 | B2 가 늘 알림(익절 · 꺼짐도) | `./internal/app/engine/` `TestA091` | CAUGHT — `TestA091WithAlertsOffNothingChangesButTheWording`, `TestA091ATakeProfitThatSoldNothingKeepsItsGrade` |
| M12 | 끝 경로 0주 분기 제거(옛 부분 캡 알림으로) | `./internal/app/engine/` `TestA091` | CAUGHT — `TestA091TheAugustSecondReplay`, `TestA091TheReminderWindowDecidesTheNextEpisode`, `TestA091TheReportFitsItsShare` |
| M13 | escalate 실패 줄에 계좌 복원 | `./internal/obs/` `TestA091` | CAUGHT — `TestA091TheEscalationLinesCarryNoAccount` |
| M13b | escalate 승격 줄에 계좌 복원 | `./internal/obs/` `TestA091` | CAUGHT — `TestA091TheEscalationLinesCarryNoAccount` |
| M13e | escalate 실패 줄에 계좌 복원(엔진 카나리) | `./internal/app/engine/` `TestA091NoLineOrRow` | CAUGHT — `TestA091NoLineOrRowCarriesTheAccount` |
| M14 | a091 로그 줄에 계좌 필드 | `./internal/app/engine/` `TestA091` | CAUGHT — `TestA091NoLineOrRowCarriesTheAccount` |
| M15 | 보고가 o.alert 경유(실패 로그 원문 계좌) | `./internal/app/engine/` `TestA091` | CAUGHT — `TestA091NoLineOrRowCarriesTheAccount` |
| M16 | 키에 원인 포함 | `./internal/app/engine/` `TestA091` | CAUGHT — `TestA091TheReminderWindowDecidesTheNextEpisode`, `TestA091AProtectiveZeroIsACriticalRow`, `TestA091AFloorThatCannotBeComputedIsTheSameReport` |
| M17 | 본문에서 시각 제거 | `./internal/app/engine/` `TestA091` | BUILD FAIL(무효 변이) |
| M18 | 본문에 원문 오류 | `./internal/app/engine/` `TestA091` | CAUGHT — `TestA091AFloorThatCannotBeComputedIsTheSameReport`, `TestA091NoLineOrRowCarriesTheAccount` |
| M19 | B2 반환값 변경(오류로) | `./internal/app/engine/` `TestA091` | CAUGHT — `TestA091TheOutcomeIsUnchanged`, `TestA091TheRowKeepsTheFirstCauseAndTheLogKeepsEach` |
| M20 | 끝 0주가 원안을 돌려줌(전량 제출) | `./internal/app/engine/` `TestA091` | CAUGHT — `TestA091TheAugustSecondReplay`, `TestA091TheReminderWindowDecidesTheNextEpisode`, `TestA091TheReportFitsItsShare` |
| M21 | 빈 Join 을 취소뿐으로 | `./internal/app/engine/` `TestA091` | BUILD FAIL(무효 변이) |

생존 1(M7b — 취소 판정의 ctx 확인 제거), 무효 2(M17 · M21 — 변이가 import/변수를 고아로 만들어 빌드 실패). 대조군 첫 판은 RED 였다(`-trimpath` 에서 액션 열거 시험이 `runtime.Caller` 경로를 못 엶) — 상대 경로로 수리(`353623a6`) 뒤 재실행.

## 2판 — 커밋 `3ec1efd2`(생존 · 무효 셋만)

수리: M7b 를 죽이는 시험 팔 (viii)(루프 살아 있는 취소뿐 오류 → ① 보고), M17 을 유효 변이로(`time.Time{}.Format("")`), M21 을 유효 변이로 + 잎 없는 다중 오류 시험(`a091Multi{nil, nil}`).

| id | 변이 | 패키지 · -run | 결과 |
|---|---|---|---|
| C | 무변이 대조군 1 조합 | — | GREEN |
| M7b | 취소 판정에서 ctx 확인 제거 | `./internal/app/engine/` `TestA091` | CAUGHT — `TestA091OnlyACancellationOnlyFailureIsSuppressed` |
| M17 | 본문에서 시각 제거 | `./internal/app/engine/` `TestA091` | CAUGHT — `TestA091TheReminderWindowDecidesTheNextEpisode`, `TestA091AProtectiveZeroIsACriticalRow` |
| M21 | 잎 없는 다중 오류를 취소뿐으로 | `./internal/app/engine/` `TestA091` | CAUGHT — `TestA091TheCancellationPredicate` |

**2판 결과: 25 변이 전부 CAUGHT(생존 0 · 무효 0).** — 「24」는 오기(M1e · M7b · M13b · M13e 로 늘어난 행을 세지 않음).

## 3판 — 커밋 `2567df05`(구현 리뷰 i1 뒤, 40 변이)

i1 보이스 A · B 가 찾은 생존 자리를 변이로 더했다(M22~M36): 보유 0 배제 폭(M22 · M23) · 원인 문구 · 표지(M24 · M25) · payload 필드(M26 · M27) ·
B2 가린 줄 · 기록 실패 줄(M28 · M29) · 익절 문구(M30) · 본문 · 로그의 포지션 표지(M31 · M32) · Unwrap nil(M33) · 재시도 대기 중 취소(M34, execgw) ·
옛 종류 B2 줄의 계좌(M35) · 생산 로그 싱크(M36). 이 판의 수리(ZeroFloorLog · 종목 표지 · 익절 문구)로 M17 · M18 원문을 새 본문 모양에 맞췄다.

| id | 변이 | 패키지 · -run | 결과 |
|---|---|---|---|
| C | 무변이 대조군 5 조합 | — | GREEN |
| M1 | 새 종류 등록 제거 | `./internal/obs/` `TestA091` | CAUGHT — `TestA091TheStopSoldNothingKindIsCritical`, `TestA091NoOtherGradeMoved` |
| M1e | 새 종류 등록 제거(엔진 쪽 관측) | `./internal/app/engine/` `TestA091AProtectiveZero` | CAUGHT — `TestA091AProtectiveZeroIsACriticalRow` |
| M2 | submit 이 보호 여부를 거짓으로 | `./internal/app/engine/` `TestA091` | CAUGHT — `TestA091TheAugustSecondReplay`, `TestA091TheReminderWindowDecidesTheNextEpisode`, `TestA091TheReportFitsItsShare` |
| M3 | submit 이 보호 여부를 늘 참으로 | `./internal/app/engine/` `TestA091` | CAUGHT — `TestA091ATakeProfitThatSoldNothingKeepsItsGrade`, `TestA091NoLineOrRowCarriesTheAccount` |
| M4 | 알림 켜짐 게이트 제거 | `./internal/app/engine/` `TestA091` | CAUGHT — `TestA091TheAugustSecondReplay`, `TestA091WithAlertsOffNothingChangesButTheWording`, `TestA091TheLogAndTheAlertAreOneKind` |
| M5 | 생산 배선의 설정 덮기 제거 | `./internal/app/engine/` `TestA091TheProductionAssembly` | CAUGHT — `TestA091TheProductionAssemblyReadsTheLoadedSwitch` |
| M6 | 보유 0 배제 제거 | `./internal/app/engine/` `TestA091` | CAUGHT — `TestA091AZeroHoldingIsNotAFailedStop`, `TestA091EveryFloorBoundHasItsGradeAndItsWords` |
| M7 | 합쳐진 오류에서 잎 하나만 취소여도 억제(errors.Is 의미) | `./internal/app/engine/` `TestA091` | CAUGHT — `TestA091OnlyACancellationOnlyFailureIsSuppressed`, `TestA091TheCancellationPredicate` |
| M7b | 취소 판정에서 ctx 확인 제거 | `./internal/app/engine/` `TestA091` | CAUGHT — `TestA091OnlyACancellationOnlyFailureIsSuppressed` |
| M8 | 취소 판정 늘 거짓 | `./internal/app/engine/` `TestA091` | CAUGHT — `TestA091OnlyACancellationOnlyFailureIsSuppressed`, `TestA091TheCancellationPredicate`, `TestA091AShutdownLeavesOnlyALine` |
| M9 | 기록에서 WithoutCancel 제거 | `./internal/app/engine/` `TestA091` | CAUGHT — `TestA091OnlyACancellationOnlyFailureIsSuppressed`, `TestA091AShutdownBetweenJudgementAndRecordStillRecords`, `TestA091ShutdownWaitsForTheReportWithoutADeadline` |
| M10 | B2 critical 알림 빠뜨림 | `./internal/app/engine/` `TestA091` | CAUGHT — `TestA091TheReportFitsItsShare`, `TestA091AFloorThatCannotBeComputedIsTheSameReport`, `TestA091OnlyACancellationOnlyFailureIsSuppressed` |
| M11 | B2 가 늘 알림(익절 · 꺼짐도) | `./internal/app/engine/` `TestA091` | CAUGHT — `TestA091WithAlertsOffNothingChangesButTheWording`, `TestA091ATakeProfitThatSoldNothingKeepsItsGrade`, `TestA091AShutdownLeavesOnlyALine` |
| M12 | 끝 경로 0주 분기 제거(옛 부분 캡 알림으로) | `./internal/app/engine/` `TestA091` | CAUGHT — `TestA091TheAugustSecondReplay`, `TestA091TheReminderWindowDecidesTheNextEpisode`, `TestA091TheReportFitsItsShare` |
| M13 | escalate 실패 줄에 계좌 복원 | `./internal/obs/` `TestA091` | CAUGHT — `TestA091TheEscalationLinesCarryNoAccount` |
| M13b | escalate 승격 줄에 계좌 복원 | `./internal/obs/` `TestA091` | CAUGHT — `TestA091TheEscalationLinesCarryNoAccount` |
| M13e | escalate 실패 줄에 계좌 복원(엔진 카나리) | `./internal/app/engine/` `TestA091NoLineOrRow` | CAUGHT — `TestA091NoLineOrRowCarriesTheAccount` |
| M14 | a091 로그 줄에 계좌 필드 · 원문 오류 | `./internal/app/engine/` `TestA091` | CAUGHT — `TestA091WithAlertsOffNothingChangesButTheWording`, `TestA091NoLineOrRowCarriesTheAccount` |
| M15 | 보고가 o.alert 경유(실패 로그 원문 계좌) | `./internal/app/engine/` `TestA091` | CAUGHT — `TestA091NoLineOrRowCarriesTheAccount` |
| M16 | 키에 원인 포함 | `./internal/app/engine/` `TestA091` | CAUGHT — `TestA091TheReminderWindowDecidesTheNextEpisode`, `TestA091AProtectiveZeroIsACriticalRow`, `TestA091AFloorThatCannotBeComputedIsTheSameReport` |
| M17 | 본문에서 시각 제거 | `./internal/app/engine/` `TestA091` | CAUGHT — `TestA091TheReminderWindowDecidesTheNextEpisode`, `TestA091AProtectiveZeroIsACriticalRow`, `TestA091WithAlertsOffNothingChangesButTheWording` |
| M18 | 본문에 원문 오류 | `./internal/app/engine/` `TestA091` | CAUGHT — `TestA091AProtectiveZeroIsACriticalRow`, `TestA091AFloorThatCannotBeComputedIsTheSameReport`, `TestA091WithAlertsOffNothingChangesButTheWording` |
| M19 | B2 반환값 변경(오류로) | `./internal/app/engine/` `TestA091` | CAUGHT — `TestA091TheOutcomeIsUnchanged`, `TestA091TheRowKeepsTheFirstCauseAndTheLogKeepsEach` |
| M20 | 끝 0주가 원안을 돌려줌(전량 제출) | `./internal/app/engine/` `TestA091` | CAUGHT — `TestA091TheAugustSecondReplay`, `TestA091TheReminderWindowDecidesTheNextEpisode`, `TestA091ALaterStopStillGoesOutInTheSameCycle` |
| M21 | 잎 없는 다중 오류를 취소뿐으로 | `./internal/app/engine/` `TestA091` | CAUGHT — `TestA091TheCancellationPredicate` |
| M22 | 보유 0 배제를 로컬 매도까지 넓힘 | `./internal/app/engine/` `TestA091` | CAUGHT — `TestA091EveryFloorBoundHasItsGradeAndItsWords` |
| M23 | 보유 0 배제를 스냅숏 부재 · 낡음까지 넓힘 | `./internal/app/engine/` `TestA091` | CAUGHT — `TestA091EveryFloorBoundHasItsGradeAndItsWords` |
| M24 | 부재 · 낡음 문구 맞바꿈 | `./internal/app/engine/` `TestA091` | CAUGHT — `TestA091EveryFloorBoundHasItsGradeAndItsWords` |
| M25 | 계산 실패 원인 표지 바꿈 | `./internal/app/engine/` `TestA091` | CAUGHT — `TestA091WithAlertsOffNothingChangesButTheWording`, `TestA091EveryFloorBoundHasItsGradeAndItsWords` |
| M26 | payload 에서 position_id 제거 | `./internal/app/engine/` `TestA091` | CAUGHT — `TestA091EveryFloorBoundHasItsGradeAndItsWords` |
| M27 | payload 에서 floor_bound 제거 | `./internal/app/engine/` `TestA091` | CAUGHT — `TestA091EveryFloorBoundHasItsGradeAndItsWords` |
| M28 | critical B2 의 가린 오류 줄 생략 | `./internal/app/engine/` `TestA091` | BUILD FAIL(무효 변이) |
| M29 | 기록 실패 줄 생략 | `./internal/app/engine/` `TestA091` | CAUGHT — `TestA091NoLineOrRowCarriesTheAccount` |
| M30 | 익절 본문도 「손절」 | `./internal/app/engine/` `TestA091` | CAUGHT — `TestA091ATakeProfitThatSoldNothingKeepsItsGrade` |
| M31 | 본문에서 종목 표지 제거 | `./internal/app/engine/` `TestA091` | CAUGHT — `TestA091AProtectiveZeroIsACriticalRow`, `TestA091WithAlertsOffNothingChangesButTheWording`, `TestA091ATakeProfitThatSoldNothingKeepsItsGrade` |
| M32 | 로그 줄에서 포지션 표지 제거 | `./internal/app/engine/` `TestA091` | CAUGHT — `TestA091WithAlertsOffNothingChangesButTheWording` |
| M33 | Unwrap nil 을 취소뿐으로 | `./internal/app/engine/` `TestA091` | CAUGHT — `TestA091TheCancellationPredicate` |
| M34 | 재시도 대기 중 취소를 취소로 반환(execgw) | `./internal/app/engine/` `TestA091` | CAUGHT — `TestA091OnlyACancellationOnlyFailureIsSuppressed` |
| M35 | 옛 종류 B2 줄에 계좌(카나리 범위) | `./internal/app/engine/` `TestA091NoLineOrRow` | CAUGHT — `TestA091NoLineOrRowCarriesTheAccount` |
| M36 | 생산 배선의 0주 로그 싱크 제거 | `./internal/app/engine/` `TestA091TheProductionAssembly` | CAUGHT — `TestA091TheProductionAssemblyReadsTheLoadedSwitch` |

M28 은 이 판에서 무효(닫는 중괄호 없는 치환)였다 — 한 줄 조건문으로 고쳐 커밋 `540aebe6` 에서 재실행: **M28 CAUGHT — `TestA091NoLineOrRowCarriesTheAccount`**.

**3판 결과: 40 변이 전부 CAUGHT(생존 0 · 무효 0).**
