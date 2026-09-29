# Branch Test Map: `Notifier.deliver`

- Source: `internal/obs/notifier.go` (:471-626); **편집 뒤** 측정 — `analysis/harness/coverage-post-unit3.json`(연결 워크트리 `fbc6df5f`, `./internal/obs` 시험 99개를 하나씩). 편집 전 번들은 `analysis/pre-edit/unit3/`에 보존.
- 재번호: 편집 전 B12(`:484` 래치) · B27(`:571` 래치) 삭제 → 편집 전 B13~B26 이 B12~B25 로 하나씩 당겨짐(편집 전 B18 `:520` → B17). B1~B11 불변.

| Branch | AST anchor | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 473:2 | 기본 시도 수 | (미실행) | 해당 없음(분기 불변) | 측정 표본의 시험 0개 |
| B2 | for at 479:2 | 시도 루프 | `TestA092AFailedEscalationLatchesUnconditionally`, `TestA092ARecordDoesNotWaitForAnotherSendersTransport` | 해당 없음(분기 불변) | 블록 479.51-480.25을 시험 49개가 실행, PASS |
| B3 | if at 480:3 | 발행기 없음 | (미실행) | 해당 없음(분기 불변) | 측정 표본의 시험 0개 |
| B4 | if at 485:3 | 발행 성공 | `TestA092AFailedEscalationLatchesUnconditionally`, `TestA092ARecordDoesNotWaitForAnotherSendersTransport` | 해당 없음(분기 불변) | 블록 485.17-487.22을 시험 33개가 실행, PASS |
| B5 | if at 487:4 | 정산 오류 없음 | `TestA092AFailedEscalationLatchesUnconditionally`, `TestA092ARecordDoesNotWaitForAnotherSendersTransport` | 해당 없음(분기 불변) | 블록 487.22-488.28을 시험 31개가 실행, PASS |
| B6 | switch at 488:5 | 정산 결과 분기 | `TestA092AFailedEscalationLatchesUnconditionally`, `TestA092ARecordDoesNotWaitForAnotherSendersTransport` | 해당 없음(분기 불변) | 블록 487.22-488.28을 시험 31개가 실행, PASS |
| B7 | case at 489:5 | 정산됨 | `TestA092ARecordDoesNotWaitForAnotherSendersTransport`, `TestA092BothAnnouncersBuildTheSameEvent` | 해당 없음(분기 불변) | 블록 489.32-490.40을 시험 25개가 실행, PASS |
| B8 | case at 491:5 | 선점(승인 · 남의 임차) — 래치 없음 | `TestA092AnAcknowledgementPreemptsASendInFlight`, `TestAcknowledgeCannotClearTheGateMidSend` | 편집 전 잠금이 승인을 막아 이 갈래가 전송 중 승인에서 불가 — `TestA092AnAcknowledgementPreemptsASendInFlight` 편집 전 3s 초과 FAIL | 블록 491.64-502.39을 시험 2개가 실행, PASS |
| B9 | case at 503:5 | 행 없음 → 미정산 | `TestA092AFailedEscalationLatchesUnconditionally`, `TestA092AReleaseAfterTheEpochReadIsHonoured` | 해당 없음(분기 불변) | 블록 503.33-504.85을 시험 4개가 실행, PASS |
| B10 | case at 505:5 | 모르는 결과 → 미정산 | (미실행) | 해당 없음(분기 불변) | 측정 표본의 시험 0개 |
| B11 | if at 531:4 | **unrecorded 판정**(편집 전 B11 로그 + B12 래치 → 로그만, 래치는 판정 반환) | `TestA092AFailedEscalationLatchesUnconditionally`, `TestA092AReleaseAfterTheEpochReadIsHonoured` | 변이 L02 · L16 · L04 · L05 | 블록 531.20-535.5을 시험 6개가 실행, PASS |
| B12 | if at 545:3 | 시도 기록 오류(편집 전 B13) | `TestACancelledSenderStillHandsTheLeaseBack` | 해당 없음(분기 불변) | 블록 545.21-546.20을 시험 1개가 실행, PASS |
| B13 | else at 549:10 | 오류 없음 분기(편집 전 B14) | `TestA092AReleaseAfterTheEpochReadIsHonoured`, `TestA092AReleaseBeforeTheEpochReadRelatches` | 해당 없음(분기 불변) | 블록 549.53-559.20을 시험 5개가 실행, PASS |
| B14 | if at 546:4 | 로그(편집 전 B15) | `TestACancelledSenderStillHandsTheLeaseBack` | 해당 없음(분기 불변) | 블록 546.20-548.5을 시험 1개가 실행, PASS |
| B15 | if at 549:10 | 임차 상실 · 정산됨 · 행 없음(편집 전 B16) | `TestA092AReleaseAfterTheEpochReadIsHonoured`, `TestA092AReleaseBeforeTheEpochReadRelatches` | 해당 없음(분기 불변) | 블록 549.53-559.20을 시험 5개가 실행, PASS |
| B16 | if at 559:4 | 전송 실패 로그(편집 전 B17) | `TestA092AReleaseAfterTheEpochReadIsHonoured`, `TestA092AReleaseBeforeTheEpochReadRelatches` | 해당 없음(분기 불변) | 블록 559.20-563.5을 시험 5개가 실행, PASS |
| B17 | if at 569:4 | **vanished — 자리에서 원칙 E 조건부 차단**(편집 전 B18) | `TestA092AReleaseAfterTheEpochReadIsHonoured`, `TestA092AReleaseBeforeTheEpochReadRelatches` | 변이 L03 · L13 · L04 · L05 | 블록 569.65-575.5을 시험 4개가 실행, PASS |
| B18 | if at 578:3 | 대기(편집 전 B19) | `TestACancelledSenderStillHandsTheLeaseBack`, `TestACriticalAlertStillEscalatesThroughTheSameNotifier` | 해당 없음(분기 불변) | 블록 578.25-579.20을 시험 14개가 실행, PASS |
| B19 | if at 579:4 | 문맥 종료(편집 전 B20) | `TestACancelledSenderStillHandsTheLeaseBack` | 해당 없음(분기 불변) | 블록 579.20-580.10을 시험 1개가 실행, PASS |
| B20 | switch at 596:2 | 반납 결과 분기(편집 전 B21) | `TestA092AFailedEscalationLatchesUnconditionally`, `TestA092AReleaseAfterTheEpochReadIsHonoured` | 해당 없음(분기 불변) | 블록 593.2-596.9을 시험 20개가 실행, PASS |
| B21 | case at 597:2 | 반납 오류(편집 전 B22) | (미실행) | 해당 없음(분기 불변) | 측정 표본의 시험 0개 |
| B22 | if at 598:3 | 로그(편집 전 B23) | (미실행) | 해당 없음(분기 불변) | 측정 표본의 시험 0개 |
| B23 | case at 601:2 | 반납 적용(편집 전 B24) | `TestA092AFailedEscalationLatchesUnconditionally`, `TestA092AReleaseAfterTheEpochReadIsHonoured` | 해당 없음(분기 불변) | 블록 601.49-601.49을 시험 20개가 실행, PASS |
| B24 | case at 604:2 | 반납 선점(편집 전 B25) | (미실행) | 해당 없음(분기 불변) | 측정 표본의 시험 0개 |
| B25 | if at 620:2 | **exhausted 판정** 로그(편집 전 B26; 편집 전 B27 래치는 판정 반환으로) | `TestA092AFailedEscalationLatchesUnconditionally`, `TestA092AReleaseAfterTheEpochReadIsHonoured` | 변이 L02 · L17 · L04 · L05 | 블록 620.18-624.3을 시험 19개가 실행, PASS |
