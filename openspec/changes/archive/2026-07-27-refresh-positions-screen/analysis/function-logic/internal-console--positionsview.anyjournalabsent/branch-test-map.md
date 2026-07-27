# Branch Test Map: `positionsView.AnyJournalAbsent`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 행 순회 | `TestTheJournalAbsenceNoticeAppearsOncePerPage` | yes (구현 전 공지 0회·행 반복) | yes |
| B2 | 원장 부재 보유 2종목 → 공지 1회 | 같은 테스트 | yes | yes |
