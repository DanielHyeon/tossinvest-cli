# Branch Test Map: `ExitObserver.submit`

**편집 전 좌표**(AST 18 분기, base `064f59fd9` → 재고정 `b93ed133a`, `exitloop.go` `1509`–`1634`). 덮는 시험은 이름 grep 으로 존재만 확인했고 커버리지 count 는 **미측정**(tasks 1.6 — 실행 0). 편집 뒤 표는 tasks 3.9.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:1514` 하한 산술 실패 | `TestA129DEveryOtherOutcomeLeavesOneLineAfterItsRelease`(부분시험 B1) | 미측정 | 미측정 |
| B2 | `:1521` 하한 0 → 해제 | `TestAZeroFloorSubmitsNothingAndLeavesTheLevelProposable` · `TestA129OAFailedReleaseStillSurfaces` | 미측정 | 미측정 |
| B3 | `:1543` Guardian 거절 → 알림(자리 1) · 해제 | `TestA129DEveryOtherOutcomeLeavesOneLineAfterItsRelease`(B3) · `TestA129OAFailedReleaseStillSurfaces` | 미측정 | 미측정 |
| B4 | `:1554` intent 부착 실패 | `TestA129DEveryOtherOutcomeLeavesOneLineAfterItsRelease`(B4) | 미측정 | 미측정 |
| B5 | `:1560` 주문 모양 실패 | 없음 — 도달 경로 없음(a129 BTM 기록: 관측가가 늘 수) | no | no |
| B6 | `:1567` 보호면 제출 우선순위 Critical | 미측정(이름 grep 으로 특정 못 함) | no | no |
| B7 | `:1576` 보호면 Place 뒤 하한가 시도 표시 | `TestA087SubmitUsesOneChoiceAndMarksOnlyAfterPlace` | 미측정 | 미측정 |
| B8 | `:1581` 결과 switch | B9~B18 의 시험 | 미측정 | 미측정 |
| B9 | `:1582` 확정 | `TestA129AConfirmedStopAndItsFillsEachLeaveALine` · `TestABaselineBreachProposesTheWholePosition` | 미측정 | 미측정 |
| B10 | `:1591` 결과 미상 | `TestAnInDoubtSubmissionKeepsTheProposalArmed` | 미측정 | 미측정 |
| B11 | `:1596` 미해소 결과 미상 이벤트 | `TestA129DEveryOtherOutcomeLeavesOneLineAfterItsRelease`(unresolved) | 미측정 | 미측정 |
| B12 | `:1601` 종목 진행 중 → 지연 시계 · 해제(Cancelled) | `TestA129OAFailedReleaseStillSurfaces` | 미측정 | 미측정 |
| B13 | `:1606` attempt 있음 · 비수용 미입증 → 무장 유지 | `TestA094AnUnrecordedOutcomeKeepsTheProposalArmed` | 미측정 | 미측정 |
| B14 | `:1610` B13 안 `err == nil` | `TestA094AnUnrecordedOutcomeKeepsTheProposalArmed` | 미측정 | 미측정 |
| B15 | `:1616` 기본(브로커 확정 거절 · 미발송) → 알림(자리 2) · 해제(Refused) — **a197 편집 자리** | `TestARefusedProposalReleasesTheLevelAndAlerts` · `TestA094AnUnacceptedStopIsReleasedAndProposedAgain` · `TestA129JARepeatedRefusalIsWrittenOnceUntilTheOutcomeChanges` | 미측정 | 미측정 |
| B16 | `:1618` 상세 비고가 비고 오류 있음 | 미측정 | no | no |
| B17 | 보호 손절 ∧ 확정 거절 ∧ 반대 미체결 → 대기열(a197) | `TestA197AnExternalBuyIsCancelledAtTheCycleEndAndTheStopFollows` · `TestA197ATakeProfitRefusalOpensNothing` | yes | yes |
| B18 | `:1627` 거부 vs 미발송 이벤트 선택 | `TestA129DEveryOtherOutcomeLeavesOneLineAfterItsRelease` | 미측정 | 미측정 |

a197 이 B15 에 더할 갈래(토글 · 보호 · 상태 · 사유)의 RED 시험은 design §4 1~11 이다(tasks 3.1, 아직 없음).
