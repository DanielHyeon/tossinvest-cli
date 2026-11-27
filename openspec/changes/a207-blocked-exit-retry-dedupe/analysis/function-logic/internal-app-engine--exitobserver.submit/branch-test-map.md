# Branch Test Map: `ExitObserver.submit`

Source: `internal/app/engine/exitloop.go` (1518-1644) — AST branches 18. RED/GREEN 은 a207 시험의 실제 실행(`analysis/tdd/`)으로만 yes 로 적고, 기존 시험은 이 change 에서
선택 실행으로 통과만 확인함(커버리지 count 미측정).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:1523` 하한 산술 실패 | `TestA129DEveryOtherOutcomeLeavesOneLineAfterItsRelease` | no | no |
| B2 | `:1530` 하한 0 → 해제(REFUSED) | `TestAZeroFloorSubmitsNothingAndLeavesTheLevelProposable` | no | no |
| B3 | `:1552` Guardian 거절 → 알림 · 해제 | `TestA129DEveryOtherOutcomeLeavesOneLineAfterItsRelease` | no | no |
| B4 | `:1563` intent 부착 실패 | `TestA129DEveryOtherOutcomeLeavesOneLineAfterItsRelease` | no | no |
| B5 | `:1569` 주문 모양 실패 → 해제 | 도달 시험 없음(a197 BTM 기록) | no | no |
| B6 | `:1576` 보호면 Critical 우선순위 | 미측정 | no | no |
| B7 | `:1585` 보호면 Place 뒤 하한가 시도 표시 | `TestA087SubmitUsesOneChoiceAndMarksOnlyAfterPlace` | no | no |
| B8 | `:1590` 결과 switch | B9~B18 | no | no |
| B9 | `:1591` 확정 — **a207 호출 하나** `noteUnblocked`(`:1599`, 대기열에 넣기만) | `TestA207AStopRefusedForNCyclesIsSubmittedOnTheFirstClearCycle` · `TestA207AnAcceptedStopWithoutRefusalsIsNotAnUnblock` | yes | yes |
| B10 | `:1601` 결과 미상 → 무장 유지 | `TestAnInDoubtSubmissionKeepsTheProposalArmed` | no | no |
| B11 | `:1606` 미해소 결과 미상 이벤트 | `TestA129DEveryOtherOutcomeLeavesOneLineAfterItsRelease` | no | no |
| B12 | `:1611` 종목 진행 중 → 지연 시계 · 해제(CANCELLED) | `TestA129OAFailedReleaseStillSurfaces` | no | no |
| B13 | `:1616` attempt 있음 · 비수용 미입증 → 무장 유지 | `TestA094AnUnrecordedOutcomeKeepsTheProposalArmed` | no | no |
| B14 | `:1620` B13 안 err 없음 | `TestA094AnUnrecordedOutcomeKeepsTheProposalArmed` | no | no |
| B15 | `:1626` 기본(확정 거절 · 미발송) → 알림 · 해제(REFUSED) | `TestARefusedProposalReleasesTheLevelAndAlerts` · `TestA207AStopRefusedForNCyclesIsSubmittedOnTheFirstClearCycle` | yes | yes |
| B16 | `:1628` 상세 비고 · 오류 있음 | 미측정 | no | no |
| B17 | `:1633` 보호 손절 ∧ 확정 거절 ∧ 반대 미체결 → a197 대기열 | `TestA197AnExternalBuyIsCancelledAtTheCycleEndAndTheStopFollows` | no | no |
| B18 | `:1637` 거부 vs 미발송 이벤트 | `TestA129DEveryOtherOutcomeLeavesOneLineAfterItsRelease` | no | no |
