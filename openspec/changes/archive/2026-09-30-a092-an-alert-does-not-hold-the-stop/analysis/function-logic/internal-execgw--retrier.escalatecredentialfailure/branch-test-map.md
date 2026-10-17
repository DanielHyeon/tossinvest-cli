# Branch Test Map: `Retrier.escalateCredentialFailure`

- Source: `internal/execgw/retry.go`; **편집 뒤** 측정 — `analysis/harness/``coverage-post-unit5-execgw.json`(`./internal/execgw` 시험 21개), 연결 워크트리 `e55102f0`.
- 재번호: difflib 정렬(편집 전 `b01e0cd0` 소스 대비): 새 분기 B3(:415).

| Branch | AST anchor | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 410:2 | `if r.Escalate == nil \|\| strings.TrimSpace(r.AccountRef) ==` | `TestAuthClassificationStillLatchesThroughTheSentinel`, `TestAuthFailureLatchesEntryImmediately` | RED: `analysis/mutation-unit5/red-b7-pre-edit.log`(B#7 · A#2) · 변이 N12 | 블록 410.64-412.3을 시험 4개가 실행, PASS |
| B2 | if at 413:2 | `if _, _, err := r.Escalate.EscalateOperatingMode(ctx, r.Acco` | (미실행) | RED: `analysis/mutation-unit5/red-b7-pre-edit.log`(B#7 · A#2) · 변이 N12 | 측정 표본의 시험 0개 |
| B3 | if at 415:3 | `if errors.Is(err, journal.ErrModeAnnouncementFailed)` | `TestA092AnUnannouncedCredentialTighteningIsReportedAsTightened` | RED: `analysis/mutation-unit5/red-b7-pre-edit.log`(B#7 · A#2) · 변이 N12 | 블록 415.56-420.4을 시험 1개가 실행, PASS |
