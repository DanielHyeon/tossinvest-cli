# Branch Test Map: `Notifier.deliver`

- Source: `internal/obs/notifier.go` (:471-626); **편집 뒤** 측정 — `analysis/harness/coverage-post-r25fix.json`(연결 워크트리 `55963f29`, `./internal/obs` 시험 99개를 하나씩). 편집 전 번들은 `analysis/pre-edit/unit3/`에 보존.
- 재번호: 편집 전 B12(`:484` 래치) · B27(`:571` 래치) 삭제 → 편집 전 B13~B26 이 B12~B25 로 하나씩 당겨짐(편집 전 B18 `:520` → B17). B1~B11 불변.
- 25라운드 수리(`55963f29`) 재번호: 새 B24(반납 행 없음 갈래) · B25(그 조건부 차단), 옛 B24(반납 선점) → B26, 옛 B25(소진 로그) → B27.

| Branch | AST anchor | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 476:2 | 기본 시도 수 | (미실행) | 해당 없음(분기 불변) | 측정 표본의 시험 0개 |
| B2 | for at 482:2 | 시도 루프 | `TestA092AFailedEscalationLatchesUnconditionally`, `TestA092ARecordDoesNotWaitForAnotherSendersTransport` | 해당 없음(분기 불변) | 블록 482.51-483.25을 시험 49개가 실행, PASS |
| B3 | if at 483:3 | 발행기 없음 | (미실행) | 해당 없음(분기 불변) | 측정 표본의 시험 0개 |
| B4 | if at 488:3 | 발행 성공 | `TestA092AFailedEscalationLatchesUnconditionally`, `TestA092ARecordDoesNotWaitForAnotherSendersTransport` | 해당 없음(분기 불변) | 블록 488.17-490.22을 시험 33개가 실행, PASS |
| B5 | if at 490:4 | 정산 오류 없음 | `TestA092AFailedEscalationLatchesUnconditionally`, `TestA092ARecordDoesNotWaitForAnotherSendersTransport` | 해당 없음(분기 불변) | 블록 490.22-491.28을 시험 31개가 실행, PASS |
| B6 | switch at 491:5 | 정산 결과 분기 | `TestA092AFailedEscalationLatchesUnconditionally`, `TestA092ARecordDoesNotWaitForAnotherSendersTransport` | 해당 없음(분기 불변) | 블록 490.22-491.28을 시험 31개가 실행, PASS |
| B7 | case at 492:5 | 정산됨 | `TestA092ARecordDoesNotWaitForAnotherSendersTransport`, `TestA092BothAnnouncersBuildTheSameEvent` | 해당 없음(분기 불변) | 블록 492.32-493.40을 시험 25개가 실행, PASS |
| B8 | case at 494:5 | 선점(승인 · 남의 임차) — 래치 없음 | `TestA092AnAcknowledgementPreemptsASendInFlight`, `TestAcknowledgeCannotClearTheGateMidSend` | 단위 ③ 편집 전 3s 초과 FAIL | 블록 494.64-505.39을 시험 2개가 실행, PASS |
| B9 | case at 506:5 | 행 없음 → 미정산 | `TestA092AFailedEscalationLatchesUnconditionally`, `TestA092AReleaseAfterTheEpochReadIsHonoured` | 해당 없음(분기 불변) | 블록 506.33-507.85을 시험 4개가 실행, PASS |
| B10 | case at 508:5 | 모르는 결과 → 미정산 | (미실행) | 해당 없음(분기 불변) | 측정 표본의 시험 0개 |
| B11 | if at 534:4 | **unrecorded 판정** | `TestA092AFailedEscalationLatchesUnconditionally`, `TestA092AReleaseAfterTheEpochReadIsHonoured` | 변이 L02 · L16 · L04 · L05 | 블록 534.20-538.5을 시험 6개가 실행, PASS |
| B12 | if at 548:3 | 시도 기록 오류 | `TestACancelledSenderStillHandsTheLeaseBack` | 해당 없음(분기 불변) | 블록 548.21-549.20을 시험 1개가 실행, PASS |
| B13 | else at 552:10 | 오류 없음 분기 | `TestA092AReleaseAfterTheEpochReadIsHonoured`, `TestA092AReleaseBeforeTheEpochReadRelatches` | 해당 없음(분기 불변) | 블록 552.53-562.20을 시험 5개가 실행, PASS |
| B14 | if at 549:4 | 로그 | `TestACancelledSenderStillHandsTheLeaseBack` | 해당 없음(분기 불변) | 블록 549.20-551.5을 시험 1개가 실행, PASS |
| B15 | if at 552:10 | 임차 상실 · 정산됨 · 행 없음 | `TestA092AReleaseAfterTheEpochReadIsHonoured`, `TestA092AReleaseBeforeTheEpochReadRelatches` | 해당 없음(분기 불변) | 블록 552.53-562.20을 시험 5개가 실행, PASS |
| B16 | if at 562:4 | 전송 실패 로그 | `TestA092AReleaseAfterTheEpochReadIsHonoured`, `TestA092AReleaseBeforeTheEpochReadRelatches` | 해당 없음(분기 불변) | 블록 562.20-566.5을 시험 5개가 실행, PASS |
| B17 | if at 572:4 | **vanished — 자리에서 원칙 E 조건부 차단** | `TestA092AReleaseAfterTheEpochReadIsHonoured`, `TestA092AReleaseBeforeTheEpochReadRelatches` | 변이 L03 · L13 | 블록 572.65-578.5을 시험 4개가 실행, PASS |
| B18 | if at 581:3 | 대기 | `TestACancelledSenderStillHandsTheLeaseBack`, `TestACriticalAlertStillEscalatesThroughTheSameNotifier` | 해당 없음(분기 불변) | 블록 581.25-582.20을 시험 14개가 실행, PASS |
| B19 | if at 582:4 | 문맥 종료 | `TestACancelledSenderStillHandsTheLeaseBack` | 해당 없음(분기 불변) | 블록 582.20-583.10을 시험 1개가 실행, PASS |
| B20 | switch at 600:2 | 반납 결과 분기 | `TestA092AFailedEscalationLatchesUnconditionally`, `TestA092AReleaseAfterTheEpochReadIsHonoured` | 해당 없음(분기 불변) | 블록 596.2-600.9을 시험 20개가 실행, PASS |
| B21 | case at 601:2 | 반납 오류 | (미실행) | 해당 없음(분기 불변) | 측정 표본의 시험 0개 |
| B22 | if at 602:3 | 로그 | (미실행) | 해당 없음(분기 불변) | 측정 표본의 시험 0개 |
| B23 | case at 605:2 | 반납 적용 → 소진 판정 | `TestA092AFailedEscalationLatchesUnconditionally`, `TestA092AReleaseAfterTheEpochReadIsHonoured` | 해당 없음(분기 불변) | 블록 605.49-605.49을 시험 20개가 실행, PASS |
| B24 | case at 608:2 | **반납 행 없음 · 모르는 결과 — 선점 아님(25라운드 codex P0 수리, 새 분기)** | `TestA092AReleaseAfterTheEpochReadIsHonoured`, `TestA092AReleaseBeforeTheEpochReadRelatches` | 수리 전(`22db26e7`): 이 결과가 선점 갈래로 빠져 래치 없음 — 변이 L20(수리 되돌림) CAUGHT | 블록 608.103-613.20을 시험 3개가 실행, PASS |
| B25 | if at 613:3 | **그 자리의 원칙 E 조건부 차단(새 분기)** | `TestA092AReleaseAfterTheEpochReadIsHonoured`, `TestA092AReleaseBeforeTheEpochReadRelatches` | 변이 L20 | 블록 613.20-618.4을 시험 3개가 실행, PASS |
| B26 | case at 620:2 | 반납 선점(AlreadySettled · LeaseLost) → lost | (미실행) | 해당 없음(분기 불변) | 측정 표본의 시험 0개 |
| B27 | if at 636:2 | **exhausted 판정** 로그 | `TestA092AFailedEscalationLatchesUnconditionally`, `TestA092AReleaseAfterTheEpochReadIsHonoured` | 해당 없음(분기 불변) | 블록 636.18-640.3을 시험 19개가 실행, PASS |
