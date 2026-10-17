# Branch Test Map: `Notifier.deliver`

- Source: `internal/obs/notifier.go` (:477-645); **편집 뒤** 측정 — `analysis/harness/coverage-post-r26d-obs.json`(연결 워크트리 `15b64676`, `internal/obs` 시험 118개를 하나씩, 실패 0). 편집 전 번들은 `analysis/pre-edit/r26b/internal-obs--notifier.deliver/`에 보존.
- 재번호: 없음.

| Branch | AST anchor | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 479:2 | 기본 시도 수 | (미실행) | 해당 없음(분기 불변) | 측정 표본의 시험 0개 |
| B2 | for at 485:2 | 시도 루프 | `TestA092AFailedEscalationLatchKeepsTheAccountOut`, `TestA092AFailedEscalationLatchesUnconditionally` | 해당 없음(분기 불변) | 블록 485.51-486.25을 시험 51개가 실행, PASS |
| B3 | if at 486:3 | 발행기 없음 | (미실행) | 해당 없음(분기 불변) | 측정 표본의 시험 0개 |
| B4 | if at 491:3 | 발행 성공 | `TestA092AFailedEscalationLatchKeepsTheAccountOut`, `TestA092AFailedEscalationLatchesUnconditionally` | 해당 없음(분기 불변) | 블록 491.17-493.22을 시험 35개가 실행, PASS |
| B5 | if at 493:4 | 정산 오류 없음 | `TestA092AFailedEscalationLatchKeepsTheAccountOut`, `TestA092AFailedEscalationLatchesUnconditionally` | 해당 없음(분기 불변) | 블록 493.22-494.28을 시험 33개가 실행, PASS |
| B6 | switch at 494:5 | 정산 결과 분기 | `TestA092AFailedEscalationLatchKeepsTheAccountOut`, `TestA092AFailedEscalationLatchesUnconditionally` | 해당 없음(분기 불변) | 블록 493.22-494.28을 시험 33개가 실행, PASS |
| B7 | case at 495:5 | 정산됨 | `TestA092ARecordDoesNotWaitForAnotherSendersTransport`, `TestA092ATakeoverIsAWarningThatNamesTheDeadSender` | 해당 없음(분기 불변) | 블록 495.32-496.40을 시험 26개가 실행, PASS |
| B8 | case at 497:5 | 선점(승인 · 남의 임차) — 래치 없음 | `TestA092AnAcknowledgementPreemptsASendInFlight`, `TestAcknowledgeCannotClearTheGateMidSend` | 해당 없음(분기 불변) | 블록 497.64-508.39을 시험 2개가 실행, PASS |
| B9 | case at 509:5 | 행 없음 → 미정산 | `TestA092AFailedEscalationLatchKeepsTheAccountOut`, `TestA092AFailedEscalationLatchesUnconditionally` | 해당 없음(분기 불변) | 블록 509.33-510.85을 시험 5개가 실행, PASS |
| B10 | case at 511:5 | 모르는 결과 → 미정산 | (미실행) | 해당 없음(분기 불변) | 측정 표본의 시험 0개 |
| B11 | if at 537:4 | **unrecorded 판정** | `TestA092AFailedEscalationLatchKeepsTheAccountOut`, `TestA092AFailedEscalationLatchesUnconditionally` | 해당 없음(분기 불변) | 블록 537.20-541.5을 시험 7개가 실행, PASS |
| B12 | if at 551:3 | 시도 기록 오류 | `TestACancelledSenderStillHandsTheLeaseBack` | 해당 없음(분기 불변) | 블록 551.21-552.20을 시험 1개가 실행, PASS |
| B13 | else at 555:10 | 오류 없음 분기 | `TestA092AReleaseAfterTheEpochReadIsHonoured`, `TestA092AReleaseBeforeTheEpochReadRelatches` | 해당 없음(분기 불변) | 블록 555.53-565.20을 시험 5개가 실행, PASS |
| B14 | if at 552:4 | 로그 | `TestACancelledSenderStillHandsTheLeaseBack` | 해당 없음(분기 불변) | 블록 552.20-554.5을 시험 1개가 실행, PASS |
| B15 | if at 555:10 | 임차 상실 · 정산됨 · 행 없음 | `TestA092AReleaseAfterTheEpochReadIsHonoured`, `TestA092AReleaseBeforeTheEpochReadRelatches` | 해당 없음(분기 불변) | 블록 555.53-565.20을 시험 5개가 실행, PASS |
| B16 | if at 565:4 | 전송 실패 로그 | `TestA092AReleaseAfterTheEpochReadIsHonoured`, `TestA092AReleaseBeforeTheEpochReadRelatches` | 해당 없음(분기 불변) | 블록 565.20-569.5을 시험 5개가 실행, PASS |
| B17 | if at 575:4 | **시도 기록이 선점 아님(행 없음 · 모르는 결과 — `isPreemption`) → 원칙 E 조건부 차단** | `TestA092AReleaseAfterTheEpochReadIsHonoured`, `TestA092AReleaseBeforeTheEpochReadRelatches` | X06 CAUGHT(`isPreemption` 확대) | 블록 575.54-581.5을 시험 4개가 실행, PASS |
| B18 | if at 584:3 | 대기 | `TestACancelledSenderStillHandsTheLeaseBack`, `TestACriticalAlertStillEscalatesThroughTheSameNotifier` | 해당 없음(분기 불변) | 블록 584.25-585.20을 시험 14개가 실행, PASS |
| B19 | if at 585:4 | 문맥 종료 | `TestACancelledSenderStillHandsTheLeaseBack` | 해당 없음(분기 불변) | 블록 585.20-586.10을 시험 1개가 실행, PASS |
| B20 | switch at 603:2 | 반납 결과 분기 | `TestA092AFailedEscalationLatchKeepsTheAccountOut`, `TestA092AFailedEscalationLatchesUnconditionally` | 해당 없음(분기 불변) | 블록 599.2-603.9을 시험 21개가 실행, PASS |
| B21 | case at 604:2 | 반납 오류 | (미실행) | 해당 없음(분기 불변) | 측정 표본의 시험 0개 |
| B22 | if at 605:3 | 로그 | (미실행) | 해당 없음(분기 불변) | 측정 표본의 시험 0개 |
| B23 | case at 608:2 | 반납 적용 → 소진 판정 | `TestA092AFailedEscalationLatchKeepsTheAccountOut`, `TestA092AFailedEscalationLatchesUnconditionally` | 해당 없음(분기 불변) | 블록 608.49-608.49을 시험 21개가 실행, PASS |
| B24 | case at 611:2 | **반납이 선점 아님(행 없음 · 모르는 결과 — `isPreemption`) → 로그** | `TestA092AReleaseAfterTheEpochReadIsHonoured`, `TestA092AReleaseBeforeTheEpochReadRelatches` | X06 CAUGHT | 블록 611.39-616.20을 시험 3개가 실행, PASS |
| B25 | if at 616:3 | **그 자리의 원칙 E 조건부 차단(새 분기)** | `TestA092AReleaseAfterTheEpochReadIsHonoured`, `TestA092AReleaseBeforeTheEpochReadRelatches` | 해당 없음(분기 불변) | 블록 616.20-621.4을 시험 3개가 실행, PASS |
| B26 | case at 623:2 | 반납 선점(AlreadySettled · LeaseLost) → lost | (미실행) | 해당 없음(분기 불변) | 측정 표본의 시험 0개 |
| B27 | if at 639:2 | **exhausted 판정** 로그 | `TestA092AFailedEscalationLatchKeepsTheAccountOut`, `TestA092AFailedEscalationLatchesUnconditionally` | 해당 없음(분기 불변) | 블록 639.18-643.3을 시험 20개가 실행, PASS |
