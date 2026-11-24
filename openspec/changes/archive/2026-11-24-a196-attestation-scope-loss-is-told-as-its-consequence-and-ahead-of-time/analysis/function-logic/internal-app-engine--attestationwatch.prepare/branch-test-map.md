# Branch Test Map: `attestationwatch.prepare`

- 편집 후. 커버: `../../implementation/after-cover-attestation_watch.out`(`GOFLAGS=-trimpath go test -count=1 -run 'A196|A135' -covermode=set ./internal/app/engine`, rc 0). RED 원문 `../../tdd/red.txt`.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 판정 패닉 회수 | `TestA135CheckPanicDoesNotStopTheWatch` | 해당 없음(기존) | yes — 311.35-313.20 count 1 |
| B2 | 회수 + 로거 | 없음 | 해당 없음 | no — 313.20-317.5 count 0(편집 전과 같음) |
| B3 | 기한 변경 → 재설정 / 같은 증거 재발급 → 재설정 없음 | `TestA196SameProofReissueIsNotAnnouncedAgain` · `TestA196NewSupervisedProofRestartsTheCount` | yes — red.txt(SameProofReissue FAIL: 재발급이 재알림) | yes — 328.45-331.3 count 1 |
| B4 | 거절 에피소드 판정 switch(G12) | 아래 | 해당 없음(새 갈래) | yes — 335.2-336.17 count 1 |
| B5 | 거절 아님(healthy·notice·imminent) → 에피소드 없음 | `TestA196RecurrenceAfterRecoveryReachesTheRealOutbox`(회복 판정) | yes — 변이 E1 이 이 시험으로 죽음(mutation run-4) | yes — 337.48-337.48 count 1(빈 갈래) |
| B6 | 거절 단계 → 거절 신원 | `TestA196ScopeLossOncePerDayAcrossReissues` | 해당 없음 | yes — 338.10-339.115 count 1 |
| B7 | 거절 신원이 직전과 다름 → 새 에피소드 시작 | `TestA196RecurrenceAfterRecoveryReachesTheRealOutbox` · `TestA196RestartedWatchStartsANewRejectionEpisode` | yes — 변이 E2(매 판정 새 에피소드)가 `TestA196ScopeLossOncePerDayAcrossReissues` 로 죽음 | yes — 341.26-343.3 count 1 |
| B8 | healthy | `TestA135WatchAnnouncesEachStageOnce` | 해당 없음 | yes — 345.29-347.3 |
| B9 | 이미 보냄 | `TestA196ScopeLossOncePerDayAcrossReissues` | yes — red.txt(재발급마다 재알림 FAIL) | yes — 349.17-351.3 |
