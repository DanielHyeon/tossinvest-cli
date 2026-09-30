# Branch Test Map: `currentModeFromRow`

- Source: `internal/journal/operating_mode.go`; **편집 전** 측정 — `analysis/harness/coverage-pre-unit4-journal.json`(`./internal/journal` 시험 28개), 연결 워크트리 `22db26e7`.

| Branch | AST anchor | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | switch at 657:2 | `switch` | `TestAFailedAnnouncementDoesNotUndoTheTransition`, `TestANoOpTransitionIsNotAnnounced` | 편집 전(기준선) | 블록 655.79-657.9을 시험 14개가 실행, PASS |
| B2 | case at 658:2 | `case errors.Is(err, errNoOperatingMode):` | `TestAFailedAnnouncementDoesNotUndoTheTransition`, `TestANoOpTransitionIsNotAnnounced` | 편집 전(기준선) | 블록 658.42-660.66을 시험 14개가 실행, PASS |
| B3 | case at 661:2 | `case err != nil:` | (미실행) | 편집 전(기준선) | 측정 표본의 시험 0개 |
