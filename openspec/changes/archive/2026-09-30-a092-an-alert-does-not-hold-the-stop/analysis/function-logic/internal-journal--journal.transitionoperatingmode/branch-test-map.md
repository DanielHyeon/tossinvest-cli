# Branch Test Map: `Journal.TransitionOperatingMode`

- Source: `internal/journal/operating_mode.go`; **편집 뒤** 측정 — `analysis/harness/``coverage-post-unit4-journal.json`(`./internal/journal` 시험 30개), 연결 워크트리 `2714e393`.
- 재번호: difflib 정렬(편집 전 `22db26e7` 소스 대비): B1~B21 → B1~B21(같은 분기, 줄 이동); 교체 B22 → B22, B23; B23~B28 → B24~B29(같은 분기, 줄 이동).

| Branch | AST anchor | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | switch at 360:2 | `switch` | `TestA092CurrentModeIsTheLatestCommitNotTheLatestClock`, `TestA092TransitionsCarryTheirCommitSequence` | 편집 전 RED: `analysis/mutation-unit4/red-*.log` · 변이 M03 · M01 | 블록 353.126-360.9을 시험 16개가 실행, PASS |
| B2 | case at 361:2 | `case account == "":` | `TestUnusableTransitionRequestsAreRefused` | 편집 전 RED: `analysis/mutation-unit4/red-*.log` · 변이 M03 · M01 | 블록 361.21-363.87을 시험 1개가 실행, PASS |
| B3 | case at 364:2 | `case !ValidOperatingMode(mode):` | `TestUnusableTransitionRequestsAreRefused` | 편집 전 RED: `analysis/mutation-unit4/red-*.log` · 변이 M03 · M01 | 블록 364.33-366.86을 시험 1개가 실행, PASS |
| B4 | case at 367:2 | `case !ValidModeActor(actor):` | `TestUnusableTransitionRequestsAreRefused` | 편집 전 RED: `analysis/mutation-unit4/red-*.log` · 변이 M03 · M01 | 블록 367.30-370.37을 시험 1개가 실행, PASS |
| B5 | case at 371:2 | `case cause == "":` | `TestUnusableTransitionRequestsAreRefused` | 편집 전 RED: `analysis/mutation-unit4/red-*.log` · 변이 M03 · M01 | 블록 371.19-374.22을 시험 1개가 실행, PASS |
| B6 | if at 378:2 | `if actor == ModeActorAuto` | `TestA092CurrentModeIsTheLatestCommitNotTheLatestClock`, `TestA092TransitionsCarryTheirCommitSequence` | 편집 전 RED: `analysis/mutation-unit4/red-*.log` · 변이 M03 · M01 | 블록 378.28-379.15을 시험 13개가 실행, PASS |
| B7 | switch at 379:3 | `switch mode` | `TestA092CurrentModeIsTheLatestCommitNotTheLatestClock`, `TestA092TransitionsCarryTheirCommitSequence` | 편집 전 RED: `analysis/mutation-unit4/red-*.log` · 변이 M03 · M01 | 블록 378.28-379.15을 시험 13개가 실행, PASS |
| B8 | case at 380:3 | `case ModeHaltAll:` | `TestARefusedTransitionIsNotAnnounced`, `TestTheFullTransitionMatrix` | 편집 전 RED: `analysis/mutation-unit4/red-*.log` · 변이 M03 · M01 | 블록 380.20-381.67을 시험 2개가 실행, PASS |
| B9 | case at 382:3 | `case ModeNormal:` | `TestTheFullTransitionMatrix` | 편집 전 RED: `analysis/mutation-unit4/red-*.log` · 변이 M03 · M01 | 블록 382.19-386.89을 시험 1개가 실행, PASS |
| B10 | if at 388:3 | `if !AutomaticTrigger(cause)` | (미실행) | 편집 전 RED: `analysis/mutation-unit4/red-*.log` · 변이 M03 · M01 | 측정 표본의 시험 0개 |
| B11 | if at 399:2 | `if err != nil` | (미실행) | 편집 전 RED: `analysis/mutation-unit4/red-*.log` · 변이 M03 · M01 | 측정 표본의 시험 0개 |
| B12 | if at 406:2 | `if err != nil` | (미실행) | 편집 전 RED: `analysis/mutation-unit4/red-*.log` · 변이 M03 · M01 | 측정 표본의 시험 0개 |
| B13 | if at 413:2 | `if err != nil` | (미실행) | 편집 전 RED: `analysis/mutation-unit4/red-*.log` · 변이 M03 · M01 | 측정 표본의 시험 0개 |
| B14 | switch at 416:2 | `switch` | `TestA092CurrentModeIsTheLatestCommitNotTheLatestClock`, `TestA092TransitionsCarryTheirCommitSequence` | 편집 전 RED: `analysis/mutation-unit4/red-*.log` · 변이 M03 · M01 | 블록 416.2-416.9을 시험 14개가 실행, PASS |
| B15 | case at 417:2 | `case direction == 0:` | `TestA092CurrentModeIsTheLatestCommitNotTheLatestClock`, `TestANoOpTransitionIsNotAnnounced` | 편집 전 RED: `analysis/mutation-unit4/red-*.log` · 변이 M03 · M01 | 블록 417.22-422.43을 시험 4개가 실행, PASS |
| B16 | case at 424:2 | `case direction < 0 && actor == ModeActorAuto:` | `TestConservativePrecedenceKeepsTheStricterMode`, `TestTheFullTransitionMatrix` | 편집 전 RED: `analysis/mutation-unit4/red-*.log` · 변이 M03 · M01 | 블록 424.47-428.43을 시험 2개가 실행, PASS |
| B17 | case at 430:2 | `case direction < 0:` | `TestA092CurrentModeIsTheLatestCommitNotTheLatestClock`, `TestA092TransitionsCarryTheirCommitSequence` | 편집 전 RED: `analysis/mutation-unit4/red-*.log` · 변이 M03 · M01 | 블록 430.21-431.21을 시험 4개가 실행, PASS |
| B18 | if at 431:3 | `if approval == ""` | (미실행) | 편집 전 RED: `analysis/mutation-unit4/red-*.log` · 변이 M03 · M01 | 측정 표본의 시험 0개 |
| B19 | if at 435:3 | `if req.Auditor == nil` | (미실행) | 편집 전 RED: `analysis/mutation-unit4/red-*.log` · 변이 M03 · M01 | 측정 표본의 시험 0개 |
| B20 | if at 444:2 | `if err != nil` | (미실행) | 편집 전 RED: `analysis/mutation-unit4/red-*.log` · 변이 M03 · M01 | 측정 표본의 시험 0개 |
| B21 | if at 448:2 | `if id == ""` | `TestA092CurrentModeIsTheLatestCommitNotTheLatestClock`, `TestA092TransitionsCarryTheirCommitSequence` | 편집 전 RED: `analysis/mutation-unit4/red-*.log` · 변이 M03 · M01 | 블록 448.14-450.3을 시험 14개가 실행, PASS |
| B22 | if at 455:2 | `if err != nil` | (미실행) | 편집 전 RED: `analysis/mutation-unit4/red-*.log` · 변이 M03 · M01 | 측정 표본의 시험 0개 |
| B23 | if at 462:2 | `if err != nil` | (미실행) | 편집 전 RED: `analysis/mutation-unit4/red-*.log` · 변이 M03 · M01 | 측정 표본의 시험 0개 |
| B24 | if at 475:2 | `if req.Auditor != nil` | `TestA092CurrentModeIsTheLatestCommitNotTheLatestClock`, `TestA092TransitionsCarryTheirCommitSequence` | 편집 전 RED: `analysis/mutation-unit4/red-*.log` · 변이 M03 · M01 | 블록 475.24-477.70을 시험 7개가 실행, PASS |
| B25 | if at 476:3 | `if err := req.Auditor.RecordAction(AuditActionOperatingMode,` | (미실행) | 편집 전 RED: `analysis/mutation-unit4/red-*.log` · 변이 M03 · M01 | 측정 표본의 시험 0개 |
| B26 | if at 483:2 | `if err := tx.Commit(); err != nil` | (미실행) | 편집 전 RED: `analysis/mutation-unit4/red-*.log` · 변이 M03 · M01 | 측정 표본의 시험 0개 |
| B27 | if at 490:2 | `if p := j.modeProjectorRef(); p != nil` | `TestA092CurrentModeIsTheLatestCommitNotTheLatestClock`, `TestA092TransitionsCarryTheirCommitSequence` | 편집 전 RED: `analysis/mutation-unit4/red-*.log` · 변이 M03 · M01 | 블록 490.41-492.3을 시험 14개가 실행, PASS |
| B28 | if at 493:2 | `if req.Announcer != nil` | `TestAFailedAnnouncementDoesNotUndoTheTransition`, `TestANoOpTransitionIsNotAnnounced` | 편집 전 RED: `analysis/mutation-unit4/red-*.log` · 변이 M03 · M01 | 블록 493.26-494.88을 시험 2개가 실행, PASS |
| B29 | if at 494:3 | `if err := req.Announcer.AnnounceOperatingMode(ctx, current.M` | `TestAFailedAnnouncementDoesNotUndoTheTransition` | 편집 전 RED: `analysis/mutation-unit4/red-*.log` · 변이 M03 · M01 | 블록 494.88-497.4을 시험 1개가 실행, PASS |
