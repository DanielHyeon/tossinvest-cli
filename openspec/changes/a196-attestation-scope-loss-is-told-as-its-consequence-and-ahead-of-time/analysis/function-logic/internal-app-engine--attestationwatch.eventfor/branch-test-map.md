# Branch Test Map: `attestationWatch.eventFor`

- 편집 후(리뷰 r1 처분 뒤). 커버: `../../implementation/after-cover-attestation_watch.out`(`GOFLAGS=-trimpath go test -count=1 -run 'A196|A135' -covermode=set ./internal/app/engine`, rc 0). RED 원문 `../../tdd/red.txt`.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | critical 타입 / 강등 | `TestA196OverdueIsCriticalOncePerDay` · `TestA196NotificationsOffDowngradesOverdue` | yes — red.txt | yes — 385.32-387.3 count 1 |
| B2 | 만료 시각 | 날수·만료 시험 | 해당 없음(기존) | yes — 390.27-392.3 |
| B3 | 단계 switch | 아래 | 해당 없음 | yes — 393.2-407.17 |
| B4 | 날수 단계 | 아래 | 해당 없음 | yes — 408.34-409.38 |
| B5 | 감독 증거 기한 문구(notice·imminent) | `TestA196ImminentProofNoticeTellsTheProofDeadline` | yes — 변이 N1 로 실패 확인(mutation run-2) | yes — 409.38-413.4 |
| B6 | 만료 기한 문구 | `TestA196WithoutSupervisedProofCountsTheExpiry` | yes — red.txt | yes — 413.9-417.4 |
| B7 | proof_overdue 문구 | `TestA196OverdueIsCriticalOncePerDay` | yes — red.txt | yes — 418.25-421.130 |
| B8 | expired | `TestA196RefusalTellsTheConsequenceFirst` | yes — red.txt | yes — 422.20-424.123 |
| B9 | scope_lost | `TestA196ScopeLossWordingFollowsTheMissingKind` | yes — red.txt | yes — 425.22-429.31 |
| B10 | 누락 순회 | 같은 시험 | yes — 변이 N6 | yes — 429.31-430.27 |
| B11 | POST 셈 | 같은 시험 | yes — N6 | yes — 430.27-432.5 |
| B12 | 전부 POST | 같은 시험 | yes — N6 | yes — 434.34-438.4 |
| B13 | GET 섞임 | 같은 시험 | yes — N6 | yes — 438.9-441.21 |
| B14 | GET+POST → verify run 덧붙임 | 같은 시험「POST+GET」 | yes — N6 | yes — 441.21-443.5 |
| B15 | incomplete | `TestA196RefusalTellsTheConsequenceFirst` | yes — red.txt | yes — 446.23-448.157 |
| B16 | account_mismatch | 같은 시험 | yes — red.txt | yes — 449.28-452.86 |
| B17 | unreadable | 같은 시험 | yes — red.txt | yes — 453.10-455.120 |
| B18 | expires_at | 누출 시험 | 해당 없음 | yes — 461.27-463.3 |
| B19 | days_left | `TestA196NewSupervisedProofRestartsTheCount` | 해당 없음 | yes — 464.56-466.3 |
| B20 | missing_endpoints | 누출 시험 | 해당 없음 | yes — 467.24-469.3 |
| B21 | supervised_proof_at | 누출 시험 | 해당 없음 | yes — 470.25-472.3 |
| B22 | deadline_at·kind | `TestA196ProofDeadlineGovernsTheDayCount` | yes — red.txt | yes — 473.26-476.3 |
