# Branch Test Map: `stagekey`

- 편집 후. 커버: `../../implementation/after-cover-attestation_watch.out`(`GOFLAGS=-trimpath go test -count=1 -run 'A196|A135' -covermode=set ./internal/app/engine`, rc 0). RED 원문 `../../tdd/red.txt`.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 단계 switch | 아래 | 해당 없음(새 구조) | yes — 340.59-343.17 count 1 |
| B2 | 날수 키에 기한 | `TestA196SameProofReissueIsNotAnnouncedAgain` | yes — red.txt | yes — 344.34-345.64 count 1 |
| B3 | 지남 키 = 기한 + KST 날짜 | `TestA196OverdueIsCriticalOncePerDay` | yes — red.txt | yes — 346.25-347.54 count 1 |
| B4 | 범위 상실 키 = 누락 집합 + KST 날짜 | `TestA196RestartedWatchReusesTheRejectionKey` | yes — red.txt(재발급 Key 다름 FAIL) | yes — 348.22-350.74 count 1 |
| (끝) | 거절 단계 = KST 날짜 | `TestA196RejectionDayIsTheKSTDate` · `TestA135InvalidStatesAreRemindedDaily` | base 에서도 PASS(회귀 핀) | yes — 352.2-352.36 count 1 |
