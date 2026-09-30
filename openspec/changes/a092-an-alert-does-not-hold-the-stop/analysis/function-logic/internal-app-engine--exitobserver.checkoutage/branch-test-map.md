# Branch Test Map: `ExitObserver.checkOutage`

- Source: `internal/app/engine/exitloop.go`; **편집 뒤** 측정 — `analysis/harness/``coverage-post-unit5-engine.json`(`./internal/app/engine` 시험 72개), 연결 워크트리 `e55102f0`.
- 재번호: difflib 정렬(편집 전 `b01e0cd0` 소스 대비): 새 분기 B6(:849).

| Branch | AST anchor | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 819:2 | `if since.IsZero()` | `TestA092CredentialTighteningRecordsWithoutSending`, `TestA111MissingManagedSymbolIsInvalidEvidenceAndDoesNotResetTheOutage` | RED: `analysis/mutation-unit5/red-b7-pre-edit.log`(B#7 · A#2) · 변이 N13 | 블록 819.20-821.3을 시험 2개가 실행, PASS |
| B2 | if at 822:2 | `if o.clk.Now().Sub(since) < o.outageAfter()` | `TestA092CredentialTighteningRecordsWithoutSending`, `TestA111MissingManagedSymbolIsInvalidEvidenceAndDoesNotResetTheOutage` | RED: `analysis/mutation-unit5/red-b7-pre-edit.log`(B#7 · A#2) · 변이 N13 | 블록 822.46-824.3을 시험 3개가 실행, PASS |
| B3 | if at 825:2 | `if o.outageRaised` | `TestASustainedOutageBlocksEntriesAndAlertsOnce` | RED: `analysis/mutation-unit5/red-b7-pre-edit.log`(B#7 · A#2) · 변이 N13 | 블록 825.20-827.3을 시험 1개가 실행, PASS |
| B4 | if at 843:2 | `if o.opts.Escalate == nil \|\| strings.TrimSpace(o.opts.Acco` | (미실행) | RED: `analysis/mutation-unit5/red-b7-pre-edit.log`(B#7 · A#2) · 변이 N13 | 측정 표본의 시험 0개 |
| B5 | if at 848:2 | `if err != nil` | `TestA092AnUnannouncedOutageTighteningStillCountsAsEscalated` | RED: `analysis/mutation-unit5/red-b7-pre-edit.log`(B#7 · A#2) · 변이 N13 | 블록 848.16-849.57을 시험 1개가 실행, PASS |
| B6 | if at 849:3 | `if !errors.Is(err, journal.ErrModeAnnouncementFailed)` | (미실행) | RED: `analysis/mutation-unit5/red-b7-pre-edit.log`(B#7 · A#2) · 변이 N13 | 측정 표본의 시험 0개 |
