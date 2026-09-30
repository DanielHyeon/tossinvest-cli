# a091 변이 원장

하네스 `analysis/harness/mutate.py` · 분리 worktree(`/tmp/claude-1000/a091-mut`) · 무변이 대조군 GREEN 먼저 · 치환 원문 1회 단언 · 원복 바이트 대조 · `GOFLAGS=-trimpath`.

## 1판 — 커밋 `353623a6`(24 변이)

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

**결과: 24 변이 전부 CAUGHT(생존 0 · 무효 0).**
