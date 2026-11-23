# Branch Test Map: `attestationWatch.eventFor`

- 편집 후(리뷰 r1 처분 뒤). 커버: `../../implementation/after-cover-attestation_watch.out`(`GOFLAGS=-trimpath go test -count=1 -run 'A196|A135' -covermode=set ./internal/app/engine`, rc 0). RED 원문 `../../tdd/red.txt`.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | critical 타입 / 강등 | `TestA196OverdueIsCriticalOncePerDay` · `TestA196NotificationsOffDowngradesOverdue` | yes — red.txt | yes — 366.32-368.3 count 1 |
| B2 | 만료 시각 | 날수·만료 시험 | 해당 없음(기존) | yes — 371.27-373.3 |
| B3 | 단계 switch | 아래 | 해당 없음 | yes — 374.2-384.17 |
| B4 | 날수 단계 | 아래 | 해당 없음 | yes — 385.34-386.38 |
| B5 | 감독 증거 기한 문구(notice·imminent) | `TestA196ImminentProofNoticeTellsTheProofDeadline` | yes — 변이 N1 로 실패 확인(mutation run-2) | yes — 386.38-390.4 |
| B6 | 만료 기한 문구 | `TestA196WithoutSupervisedProofCountsTheExpiry` | yes — red.txt | yes — 390.9-394.4 |
| B7 | proof_overdue 문구 | `TestA196OverdueIsCriticalOncePerDay` | yes — red.txt | yes — 395.25-398.130 |
| B8 | expired | `TestA196RefusalTellsTheConsequenceFirst` | yes — red.txt | yes — 399.20-401.123 |
| B9 | scope_lost | `TestA196ScopeLossWordingFollowsTheMissingKind` | yes — red.txt | yes — 402.22-406.31 |
| B10 | 누락 순회 | 같은 시험 | yes — 변이 N6 | yes — 406.31-407.27 |
| B11 | POST 셈 | 같은 시험 | yes — N6 | yes — 407.27-409.5 |
| B12 | 전부 POST | 같은 시험 | yes — N6 | yes — 411.34-415.4 |
| B13 | GET 섞임 | 같은 시험 | yes — N6 | yes — 415.9-418.21 |
| B14 | GET+POST → verify run 덧붙임 | 같은 시험「POST+GET」 | yes — N6 | yes — 418.21-420.5 |
| B15 | incomplete | `TestA196RefusalTellsTheConsequenceFirst` | yes — red.txt | yes — 423.23-425.157 |
| B16 | account_mismatch | 같은 시험 | yes — red.txt | yes — 426.28-429.86 |
| B17 | unreadable | 같은 시험 | yes — red.txt | yes — 430.10-432.120 |
| B18 | expires_at | 누출 시험 | 해당 없음 | yes — 438.27-440.3 |
| B19 | days_left | `TestA196NewSupervisedProofRestartsTheCount` | 해당 없음 | yes — 441.56-443.3 |
| B20 | missing_endpoints | 누출 시험 | 해당 없음 | yes — 444.24-446.3 |
| B21 | supervised_proof_at | 누출 시험 | 해당 없음 | yes — 447.25-449.3 |
| B22 | deadline_at·kind | `TestA196ProofDeadlineGovernsTheDayCount` | yes — red.txt | yes — 450.26-453.3 |
