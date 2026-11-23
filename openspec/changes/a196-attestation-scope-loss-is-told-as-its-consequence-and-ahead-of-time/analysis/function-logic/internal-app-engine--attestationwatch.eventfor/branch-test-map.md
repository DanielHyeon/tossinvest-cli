# Branch Test Map: `attestationwatch.eventfor`

- 편집 후. 커버: `../../implementation/after-cover-attestation_watch.out`(`GOFLAGS=-trimpath go test -count=1 -run 'A196|A135' -covermode=set ./internal/app/engine`, rc 0). RED 원문 `../../tdd/red.txt`.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | critical 타입 / 알림 꺼짐 강등 | `TestA196OverdueIsCriticalOncePerDay` · `TestA196NotificationsOffDowngradesOverdue` | yes — red.txt | yes — 361.32-363.3 count 1 |
| B2 | 만료 시각 | 날수·만료 시험 | 해당 없음(기존) | yes — 366.27-368.3 |
| B3 | 단계 switch | 아래 | 해당 없음 | yes — 369.2-377.17 |
| B4 | 날수 단계 | 아래 | 해당 없음 | yes — 378.34-379.38 |
| B5 | 감독 증거 기한 문구 | `TestA196ProofDeadlineGovernsTheDayCount` | yes — red.txt | yes — 379.38-383.4 |
| B6 | 만료 기한 문구 | `TestA196WithoutSupervisedProofCountsTheExpiry` | yes — red.txt(deadline_kind 없음 FAIL) | yes — 383.9-387.4 |
| B7 | proof_overdue 문구 | `TestA196OverdueIsCriticalOncePerDay` | yes — red.txt | yes — 388.25-391.152 |
| B8 | expired 거부 단정 | `TestA196RefusalTellsTheConsequenceFirst` | yes — red.txt | yes — 392.20-394.123 |
| B9 | scope_lost 거부 단정 | 같은 시험 | yes — red.txt | yes — 395.22-398.66 |
| B10 | incomplete | 같은 시험 | yes — red.txt | yes — 399.23-401.157 |
| B11 | account_mismatch | 같은 시험 | yes — red.txt | yes — 402.28-405.86 |
| B12 | unreadable | 같은 시험 | yes — red.txt | yes — 406.10-408.120 |
| B13 | expires_at | 누출 시험 | 해당 없음 | yes — 414.27-416.3 |
| B14 | days_left | `TestA196NewSupervisedProofRestartsTheCount` | 해당 없음 | yes — 417.56-419.3 |
| B15 | missing_endpoints | 누출 시험 | 해당 없음 | yes — 420.24-422.3 |
| B16 | supervised_proof_at | 누출 시험 | 해당 없음 | yes — 423.25-425.3 |
| B17 | deadline_at·kind | `TestA196ProofDeadlineGovernsTheDayCount` | yes — red.txt | yes — 426.26-429.3 |
