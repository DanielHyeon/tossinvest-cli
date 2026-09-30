# Branch Test Map: `scanOperatingMode`

- Source: `internal/journal/operating_mode.go`; **편집 전** 측정 — `analysis/harness/coverage-pre-unit4-journal.json`(`./internal/journal` 시험 28개), 연결 워크트리 `22db26e7`.

| Branch | AST anchor | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 698:2 | `if err := row.Scan(&record.ID, &record.AccountRef, &record.M` | (미실행) | 편집 전(기준선) | 측정 표본의 시험 0개 |
| B2 | if at 700:3 | `if errors.Is(err, sql.ErrNoRows)` | `TestAFailedAnnouncementDoesNotUndoTheTransition`, `TestANoOpTransitionIsNotAnnounced` | 편집 전(기준선) | 블록 700.36-702.4을 시험 14개가 실행, PASS |
| B3 | if at 706:2 | `if err != nil` | (미실행) | 편집 전(기준선) | 측정 표본의 시험 0개 |
