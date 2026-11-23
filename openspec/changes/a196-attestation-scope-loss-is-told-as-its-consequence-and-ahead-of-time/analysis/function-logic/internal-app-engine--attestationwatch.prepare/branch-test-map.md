# Branch Test Map: `attestationwatch.prepare`

- 편집 후. 커버: `../../implementation/after-cover-attestation_watch.out`(`GOFLAGS=-trimpath go test -count=1 -run 'A196|A135' -covermode=set ./internal/app/engine`, rc 0). RED 원문 `../../tdd/red.txt`.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 판정 패닉 회수 | `TestA135CheckPanicDoesNotStopTheWatch` | 해당 없음(기존) | yes — 306.35-308.20 count 1 |
| B2 | 회수 + 로거 | 없음 | 해당 없음 | no — 308.20-312.5 count 0(편집 전과 같음) |
| B3 | 기한 변경 → 재설정 / 같은 증거 재발급 → 재설정 없음 | `TestA196SameProofReissueIsNotAnnouncedAgain` · `TestA196NewSupervisedProofRestartsTheCount` | yes — red.txt(SameProofReissue FAIL: 재발급이 재알림) | yes — 323.45-326.3 count 1 |
| B4 | healthy | `TestA135WatchAnnouncesEachStageOnce` | 해당 없음 | yes — 327.29-329.3 |
| B5 | 이미 보냄 | `TestA196ScopeLossOncePerDayAcrossReissues` | yes — red.txt(재발급마다 재알림 FAIL) | yes — 331.17-333.3 |
