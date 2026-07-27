# Branch Test Map: `positionsView.AnyUnknown`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 행 순회 | `TestTheUnknownStateNoticeAppearsOncePerPage` | yes (구현 전 공지 3회) | yes |
| B2 | 미판독 행 존재 → 공지 1회 / 부재 → 공지 없음 | 같은 테스트 + `TestTheJournalAbsenceNoticeAppearsOncePerPage`(판독 케이스에서 미판독 공지 부재) | yes | yes |
