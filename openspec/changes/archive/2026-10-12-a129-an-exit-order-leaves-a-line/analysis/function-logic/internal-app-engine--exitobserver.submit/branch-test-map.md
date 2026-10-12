# Branch Test Map: `ExitObserver.submit`

**편집 뒤 좌표**(AST 15 분기). 편집 전 표(13 분기, base `381e73da`)는 git 이력(이 change 의 계약 커밋)에 있다.
진입 실측: 편집 뒤 엔진 패키지 전체 `-coverprofile`(set 모드, 2026-10-12, 244s, rc 0) 블록 count.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:1420` 하한 산술 실패 | `TestA129DEveryOtherOutcomeLeavesOneLineAfterItsRelease/B1_floor_arithmetic_failed` | yes (red-1: 컴파일 실패) | yes |
| B2 | `:1427` 하한 0 → 해제 | `…/B2_zero_floor` · `TestA129OAFailedReleaseStillSurfaces/B2_zero_floor`(해제 실패 전파) · `TestAZeroFloorSubmitsNothingAndLeavesTheLevelProposable` | yes | yes |
| B3 | `:1449` Guardian 거절 | `…/B3_guardian_refused` · `TestA129OAFailedReleaseStillSurfaces/B3_guardian_refused` | yes | yes |
| B4 | `:1460` intent 부착 실패 | `…/B4_attach_failed` | yes | yes |
| B5 | `:1466` 주문 모양 실패 | 없음 — 블록 `1466-1471` count 0(편집 전과 같음: 관측가가 늘 수라 도달 경로 없음) | no | no |
| B6 | `:1480` 결과 switch | B7~B15 의 시험 | yes | yes |
| B7 | `:1481` 확정 | `TestA129AConfirmedStopAndItsFillsEachLeaveALine` · `TestABaselineBreachProposesTheWholePosition` | yes | yes |
| B8 | `:1490` 결과 미상 | `…/B8_in_doubt` · `…/B8_unresolved_in_doubt` · `TestAnInDoubtSubmissionKeepsTheProposalArmed` | yes | yes |
| B9 | `:1495` 미해소 결과 미상 → `order.unresolved_in_doubt` | `…/B8_unresolved_in_doubt` | yes | yes |
| B10 | `:1500` 종목 진행 중 | `…/B9_symbol_in_flight`(편집 전 번호로 이름 붙임) · `TestA129OAFailedReleaseStillSurfaces/B10_symbol_in_flight` | yes | yes |
| B11 | `:1505` attempt 있음 · 비수용 미입증 | `…/B10_unsettled_attempt` · `TestA094AnUnrecordedOutcomeKeepsTheProposalArmed` | yes | yes |
| B12 | `:1509` B11 안 `err == nil` | `…/B10_unsettled_attempt`(가짜 Place 가 nil 오류) | no | yes |
| B13 | `:1515` 기본(거부·미발송) | `…/B12_broker_rejected` · `…/B12_not_dispatched` · `TestARefusedProposalReleasesTheLevelAndAlerts` · `TestA129JARepeatedRefusalIsWrittenOnceUntilTheOutcomeChanges` · `TestA129MTheRefusalLatchDoesNotOutliveItsEpisode` · `TestA129OAFailedReleaseStillSurfaces/B13_broker_rejected` | yes | yes |
| B14 | `:1517` 상세 비고 오류 있음 | 없음 — count 0(무편집, 줄에 무관) | no | no |
| B15 | `:1522` 거부 vs 미발송 이벤트 선택 | `…/B12_broker_rejected`(참) · `…/B12_not_dispatched`(거짓) | yes | yes |

변이 원장 최종판: `analysis/mutation/round4.log` — CONTROL GREEN, **19/19 CAUGHT(전부 시험 실패, 컴파일 실패 0)**. round1(12/12) · round2 는 이력이며 round2 의 M15~M18 「CAUGHT」 는 미사용 변수 컴파일 실패였다(하네스가 BUILD-FAILED 를 따로 세게 고친 뒤 round4 에서 실제 시험으로 잡힘 확인).
