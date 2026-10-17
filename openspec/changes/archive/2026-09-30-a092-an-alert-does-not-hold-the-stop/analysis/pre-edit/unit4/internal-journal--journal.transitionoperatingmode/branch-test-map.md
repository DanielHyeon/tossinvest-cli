# Branch Test Map: `Journal.TransitionOperatingMode`

- Source: `internal/journal/operating_mode.go`; **편집 전** 측정 — `analysis/harness/coverage-pre-unit4-journal.json`(`./internal/journal` 시험 28개), 연결 워크트리 `22db26e7`.

| Branch | AST anchor | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | switch at 353:2 | `switch` | `TestAFailedAnnouncementDoesNotUndoTheTransition`, `TestANoOpTransitionIsNotAnnounced` | 편집 전(기준선) | 블록 346.126-353.9을 시험 14개가 실행, PASS |
| B2 | case at 354:2 | `case account == "":` | `TestUnusableTransitionRequestsAreRefused` | 편집 전(기준선) | 블록 354.21-356.87을 시험 1개가 실행, PASS |
| B3 | case at 357:2 | `case !ValidOperatingMode(mode):` | `TestUnusableTransitionRequestsAreRefused` | 편집 전(기준선) | 블록 357.33-359.86을 시험 1개가 실행, PASS |
| B4 | case at 360:2 | `case !ValidModeActor(actor):` | `TestUnusableTransitionRequestsAreRefused` | 편집 전(기준선) | 블록 360.30-363.37을 시험 1개가 실행, PASS |
| B5 | case at 364:2 | `case cause == "":` | `TestUnusableTransitionRequestsAreRefused` | 편집 전(기준선) | 블록 364.19-367.22을 시험 1개가 실행, PASS |
| B6 | if at 371:2 | `if actor == ModeActorAuto` | `TestAFailedAnnouncementDoesNotUndoTheTransition`, `TestANoOpTransitionIsNotAnnounced` | 편집 전(기준선) | 블록 371.28-372.15을 시험 11개가 실행, PASS |
| B7 | switch at 372:3 | `switch mode` | `TestAFailedAnnouncementDoesNotUndoTheTransition`, `TestANoOpTransitionIsNotAnnounced` | 편집 전(기준선) | 블록 371.28-372.15을 시험 11개가 실행, PASS |
| B8 | case at 373:3 | `case ModeHaltAll:` | `TestARefusedTransitionIsNotAnnounced`, `TestTheFullTransitionMatrix` | 편집 전(기준선) | 블록 373.20-374.67을 시험 2개가 실행, PASS |
| B9 | case at 375:3 | `case ModeNormal:` | `TestTheFullTransitionMatrix` | 편집 전(기준선) | 블록 375.19-379.89을 시험 1개가 실행, PASS |
| B10 | if at 381:3 | `if !AutomaticTrigger(cause)` | (미실행) | 편집 전(기준선) | 측정 표본의 시험 0개 |
| B11 | if at 392:2 | `if err != nil` | (미실행) | 편집 전(기준선) | 측정 표본의 시험 0개 |
| B12 | if at 399:2 | `if err != nil` | (미실행) | 편집 전(기준선) | 측정 표본의 시험 0개 |
| B13 | if at 406:2 | `if err != nil` | (미실행) | 편집 전(기준선) | 측정 표본의 시험 0개 |
| B14 | switch at 409:2 | `switch` | `TestAFailedAnnouncementDoesNotUndoTheTransition`, `TestANoOpTransitionIsNotAnnounced` | 편집 전(기준선) | 블록 409.2-409.9을 시험 12개가 실행, PASS |
| B15 | case at 410:2 | `case direction == 0:` | `TestANoOpTransitionIsNotAnnounced`, `TestConcurrentEscalationsConvergeOnTheStrictestMode` | 편집 전(기준선) | 블록 410.22-415.43을 시험 3개가 실행, PASS |
| B16 | case at 417:2 | `case direction < 0 && actor == ModeActorAuto:` | `TestConservativePrecedenceKeepsTheStricterMode`, `TestTheFullTransitionMatrix` | 편집 전(기준선) | 블록 417.47-421.43을 시험 2개가 실행, PASS |
| B17 | case at 423:2 | `case direction < 0:` | `TestTheFullTransitionMatrix`, `TestTransitionIDsDoNotCollideInsideOneSecond` | 편집 전(기준선) | 블록 423.21-424.21을 시험 2개가 실행, PASS |
| B18 | if at 424:3 | `if approval == ""` | (미실행) | 편집 전(기준선) | 측정 표본의 시험 0개 |
| B19 | if at 428:3 | `if req.Auditor == nil` | (미실행) | 편집 전(기준선) | 측정 표본의 시험 0개 |
| B20 | if at 437:2 | `if err != nil` | (미실행) | 편집 전(기준선) | 측정 표본의 시험 0개 |
| B21 | if at 441:2 | `if id == ""` | `TestAFailedAnnouncementDoesNotUndoTheTransition`, `TestANoOpTransitionIsNotAnnounced` | 편집 전(기준선) | 블록 441.14-443.3을 시험 12개가 실행, PASS |
| B22 | if at 444:2 | `if _, err := tx.ExecContext(ctx,` | (미실행) | 편집 전(기준선) | 측정 표본의 시험 0개 |
| B23 | if at 460:2 | `if req.Auditor != nil` | `TestConservativePrecedenceKeepsTheStricterMode`, `TestFlattenIsNeverGatedByTheMode` | 편집 전(기준선) | 블록 460.24-462.70을 시험 5개가 실행, PASS |
| B24 | if at 461:3 | `if err := req.Auditor.RecordAction(AuditActionOperatingMode,` | (미실행) | 편집 전(기준선) | 측정 표본의 시험 0개 |
| B25 | if at 468:2 | `if err := tx.Commit(); err != nil` | (미실행) | 편집 전(기준선) | 측정 표본의 시험 0개 |
| B26 | if at 475:2 | `if p := j.modeProjectorRef(); p != nil` | `TestAFailedAnnouncementDoesNotUndoTheTransition`, `TestANoOpTransitionIsNotAnnounced` | 편집 전(기준선) | 블록 475.41-477.3을 시험 12개가 실행, PASS |
| B27 | if at 478:2 | `if req.Announcer != nil` | `TestAFailedAnnouncementDoesNotUndoTheTransition`, `TestANoOpTransitionIsNotAnnounced` | 편집 전(기준선) | 블록 478.26-479.88을 시험 2개가 실행, PASS |
| B28 | if at 479:3 | `if err := req.Announcer.AnnounceOperatingMode(ctx, current.M` | `TestAFailedAnnouncementDoesNotUndoTheTransition` | 편집 전(기준선) | 블록 479.88-482.4을 시험 1개가 실행, PASS |
