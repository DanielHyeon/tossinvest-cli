# Branch Test Map: `Notifier.deliver`

- Source: `internal/obs/notifier.go` (:475-643); **편집 뒤** 측정 — `analysis/harness/coverage-post-r26b-obs.json`(연결 워크트리 `b910173a`, `internal/obs` 시험 116개를 하나씩, 실패 0). 편집 전 번들은 `analysis/pre-edit/r26b/internal-obs--notifier.deliver/`에 보존.
- 재번호: 없음.

| Branch | AST anchor | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 477:2 | 기본 시도 수 | (미실행) | 해당 없음(분기 불변) | 측정 표본의 시험 0개 |
| B2 | for at 483:2 | 시도 루프 | `TestA092AFailedEscalationLatchesUnconditionally`, `TestA092ARecordDoesNotWaitForAnotherSendersTransport` | 해당 없음(분기 불변) | 블록 483.51-484.25을 시험 50개가 실행, PASS |
| B3 | if at 484:3 | 발행기 없음 | (미실행) | 해당 없음(분기 불변) | 측정 표본의 시험 0개 |
| B4 | if at 489:3 | 발행 성공 | `TestA092AFailedEscalationLatchesUnconditionally`, `TestA092ARecordDoesNotWaitForAnotherSendersTransport` | 해당 없음(분기 불변) | 블록 489.17-491.22을 시험 34개가 실행, PASS |
| B5 | if at 491:4 | 정산 오류 없음 | `TestA092AFailedEscalationLatchesUnconditionally`, `TestA092ARecordDoesNotWaitForAnotherSendersTransport` | 해당 없음(분기 불변) | 블록 491.22-492.28을 시험 32개가 실행, PASS |
| B6 | switch at 492:5 | 정산 결과 분기 | `TestA092AFailedEscalationLatchesUnconditionally`, `TestA092ARecordDoesNotWaitForAnotherSendersTransport` | 해당 없음(분기 불변) | 블록 491.22-492.28을 시험 32개가 실행, PASS |
| B7 | case at 493:5 | 정산됨 | `TestA092ARecordDoesNotWaitForAnotherSendersTransport`, `TestA092ATakeoverIsAWarningThatNamesTheDeadSender` | 해당 없음(분기 불변) | 블록 493.32-494.40을 시험 26개가 실행, PASS |
| B8 | case at 495:5 | 선점(승인 · 남의 임차) — 래치 없음 | `TestA092AnAcknowledgementPreemptsASendInFlight`, `TestAcknowledgeCannotClearTheGateMidSend` | 해당 없음(분기 불변) | 블록 495.64-506.39을 시험 2개가 실행, PASS |
| B9 | case at 507:5 | 행 없음 → 미정산 | `TestA092AFailedEscalationLatchesUnconditionally`, `TestA092AReleaseAfterTheEpochReadIsHonoured` | 해당 없음(분기 불변) | 블록 507.33-508.85을 시험 4개가 실행, PASS |
| B10 | case at 509:5 | 모르는 결과 → 미정산 | (미실행) | 해당 없음(분기 불변) | 측정 표본의 시험 0개 |
| B11 | if at 535:4 | **unrecorded 판정** | `TestA092AFailedEscalationLatchesUnconditionally`, `TestA092AReleaseAfterTheEpochReadIsHonoured` | 해당 없음(분기 불변) | 블록 535.20-539.5을 시험 6개가 실행, PASS |
| B12 | if at 549:3 | 시도 기록 오류 | `TestACancelledSenderStillHandsTheLeaseBack` | 해당 없음(분기 불변) | 블록 549.21-550.20을 시험 1개가 실행, PASS |
| B13 | else at 553:10 | 오류 없음 분기 | `TestA092AReleaseAfterTheEpochReadIsHonoured`, `TestA092AReleaseBeforeTheEpochReadRelatches` | 해당 없음(분기 불변) | 블록 553.53-563.20을 시험 5개가 실행, PASS |
| B14 | if at 550:4 | 로그 | `TestACancelledSenderStillHandsTheLeaseBack` | 해당 없음(분기 불변) | 블록 550.20-552.5을 시험 1개가 실행, PASS |
| B15 | if at 553:10 | 임차 상실 · 정산됨 · 행 없음 | `TestA092AReleaseAfterTheEpochReadIsHonoured`, `TestA092AReleaseBeforeTheEpochReadRelatches` | 해당 없음(분기 불변) | 블록 553.53-563.20을 시험 5개가 실행, PASS |
| B16 | if at 563:4 | 전송 실패 로그 | `TestA092AReleaseAfterTheEpochReadIsHonoured`, `TestA092AReleaseBeforeTheEpochReadRelatches` | 해당 없음(분기 불변) | 블록 563.20-567.5을 시험 5개가 실행, PASS |
| B17 | if at 573:4 | **시도 기록이 선점 아님(행 없음 · 모르는 결과 — `isPreemption`) → 원칙 E 조건부 차단** | `TestA092AReleaseAfterTheEpochReadIsHonoured`, `TestA092AReleaseBeforeTheEpochReadRelatches` | X06 CAUGHT(`isPreemption` 확대) | 블록 573.54-579.5을 시험 4개가 실행, PASS |
| B18 | if at 582:3 | 대기 | `TestACancelledSenderStillHandsTheLeaseBack`, `TestACriticalAlertStillEscalatesThroughTheSameNotifier` | 해당 없음(분기 불변) | 블록 582.25-583.20을 시험 14개가 실행, PASS |
| B19 | if at 583:4 | 문맥 종료 | `TestACancelledSenderStillHandsTheLeaseBack` | 해당 없음(분기 불변) | 블록 583.20-584.10을 시험 1개가 실행, PASS |
| B20 | switch at 601:2 | 반납 결과 분기 | `TestA092AFailedEscalationLatchesUnconditionally`, `TestA092AReleaseAfterTheEpochReadIsHonoured` | 해당 없음(분기 불변) | 블록 597.2-601.9을 시험 20개가 실행, PASS |
| B21 | case at 602:2 | 반납 오류 | (미실행) | 해당 없음(분기 불변) | 측정 표본의 시험 0개 |
| B22 | if at 603:3 | 로그 | (미실행) | 해당 없음(분기 불변) | 측정 표본의 시험 0개 |
| B23 | case at 606:2 | 반납 적용 → 소진 판정 | `TestA092AFailedEscalationLatchesUnconditionally`, `TestA092AReleaseAfterTheEpochReadIsHonoured` | 해당 없음(분기 불변) | 블록 606.49-606.49을 시험 20개가 실행, PASS |
| B24 | case at 609:2 | **반납이 선점 아님(행 없음 · 모르는 결과 — `isPreemption`) → 로그** | `TestA092AReleaseAfterTheEpochReadIsHonoured`, `TestA092AReleaseBeforeTheEpochReadRelatches` | X06 CAUGHT | 블록 609.39-614.20을 시험 3개가 실행, PASS |
| B25 | if at 614:3 | **그 자리의 원칙 E 조건부 차단(새 분기)** | `TestA092AReleaseAfterTheEpochReadIsHonoured`, `TestA092AReleaseBeforeTheEpochReadRelatches` | 해당 없음(분기 불변) | 블록 614.20-619.4을 시험 3개가 실행, PASS |
| B26 | case at 621:2 | 반납 선점(AlreadySettled · LeaseLost) → lost | (미실행) | 해당 없음(분기 불변) | 측정 표본의 시험 0개 |
| B27 | if at 637:2 | **exhausted 판정** 로그 | `TestA092AFailedEscalationLatchesUnconditionally`, `TestA092AReleaseAfterTheEpochReadIsHonoured` | 해당 없음(분기 불변) | 블록 637.18-641.3을 시험 19개가 실행, PASS |
