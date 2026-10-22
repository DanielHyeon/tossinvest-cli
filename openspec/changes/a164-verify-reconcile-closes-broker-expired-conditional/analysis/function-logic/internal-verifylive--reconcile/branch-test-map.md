# Branch Test Map: `Reconcile`

- Source: `internal/verifylive/reconcile.go` (168-316)
- AST branches 31 — 시험 칸은 시험별 coverage 측정(본문 블록 실행)으로 채움

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if 줄 170: `if now == nil {` | 미덮임(시험별 coverage 측정 0) — 생산 호출자(cmd)는 항상 `time.Now` 를 줌 | 해당 없음 — 편집 전 기준선 | 기준선 통과(측정, 아래 「측정」) |
| B2 | if 줄 173: `if p.Reader == nil {` | 미덮임(시험별 coverage 측정 0) — 생산 호출자는 항상 리더를 줌 | 해당 없음 — 편집 전 기준선 | 기준선 통과(측정, 아래 「측정」) |
| B3 | if 줄 176: `if p.Approve == nil {` | 미덮임(시험별 coverage 측정 0) — 생산 호출자는 항상 승인 seam 을 줌 | 해당 없음 — 편집 전 기준선 | 기준선 통과(측정, 아래 「측정」) |
| B4 | if 줄 180: `if out == nil {` | 미덮임(시험별 coverage 측정 0) — 생산 호출자는 단말 출력을 줌 | 해당 없음 — 편집 전 기준선 | 기준선 통과(측정, 아래 「측정」) |
| B5 | if 줄 186: `if err != nil {` | `TestReconcileRefusesABrokenInnerLine` · `TestReconcileRefusesACompleteFinalLineWithoutNewline` · `TestReconcileRefusesANewerRecordFormat` · `TestReconcileRefusesATornFinalLineAtSelection` 외 1 | 해당 없음 — 편집 전 기준선 | 기준선 통과(측정, 아래 「측정」) |
| B6 | if 줄 193: `if err != nil {` | `TestReconcileIsIdempotent` · `TestReconcileRefusesAnUnresolvedM0Owner` · `TestReconcileRefusesSeveralCandidates` · `TestReconcileRefusesZeroCandidates` | 해당 없음 — 편집 전 기준선 | 기준선 통과(측정, 아래 「측정」) |
| B7 | if 줄 199: `if err != nil {` | `TestReconcileRefusesACurrentAccountWhoseMaskDiffers` · `TestReconcileRefusesMixedAccountReferences` · `TestReconcileRefusesUnusableAccountReferences` | 해당 없음 — 편집 전 기준선 | 기준선 통과(측정, 아래 「측정」) |
| B8 | if 줄 205: `if err := checkReconcileRetention(target, now()); err != nil {` | `TestReconcileRefusesAZeroCreatedAtOnTheOutstandingLine` · `TestReconcileRefusesAnArtifactOlderThanTheRetentionBound` · `TestReconcileRefusesAnyEvictionModelButDuration` · `TestReconcileRefusesWhileTheRetentionMeasurementIsAbsent` 외 1 | 해당 없음 — 편집 전 기준선 | 기준선 통과(측정, 아래 「측정」) |
| B9 | if 줄 208: `if reconcileFreshnessBound <= 0 {` | `TestReconcileRefusesWhileTheFreshnessBoundIsUnfixed` · `TestReconcileStaticRefusalsComeBeforeTheHumanApproval` | 해당 없음 — 편집 전 기준선 | 기준선 통과(측정, 아래 「측정」) |
| B10 | if 줄 217: `if err := approval.WriteText(out); err != nil {` | `TestReconcileRefusesWhenTheApprovalCannotBeShown` | 해당 없음 — 편집 전 기준선 | 기준선 통과(측정, 아래 「측정」) |
| B11 | if 줄 220: `if err := p.Approve(ctx, approval); err != nil {` | `TestReconcileRefusesWithoutApproval` | 해당 없음 — 편집 전 기준선 | 기준선 통과(측정, 아래 「측정」) |
| B12 | if 줄 226: `if err := checkReconcileInstrument(ctx, p.Reader, symbol); err != nil {` | `TestReconcileRefusesWhenTheInstrumentControlFails` | 해당 없음 — 편집 전 기준선 | 기준선 통과(측정, 아래 「측정」) |
| B13 | if 줄 233: `if err != nil {` | `TestADeleteNotFoundFromEveryListReadDoesNotReconcile` · `TestReconcileRefusesADuplicateRowWithinOneRead` · `TestReconcileRefusesMalformedListResponsesEvenWhenBothReadsAgree` · `TestReconcileRefusesPaginationFaultsInEitherRead` | 해당 없음 — 편집 전 기준선 | 기준선 통과(측정, 아래 「측정」) |
| B14 | if 줄 237: `if err != nil {` | `TestReconcileRefusesADuplicateRowWithinOneRead` · `TestReconcileRefusesPaginationFaultsInEitherRead` | 해당 없음 — 편집 전 기준선 | 기준선 통과(측정, 아래 「측정」) |
| B15 | if 줄 242: `if err := checkReconcileInstrument(ctx, p.Reader, symbol); err != nil {` | `TestReconcileRefusesWhenTheInstrumentControlFails` | 해당 없음 — 편집 전 기준선 | 기준선 통과(측정, 아래 「측정」) |
| B16 | if 줄 245: `if err := checkReconcileWindow(target, start, now()); err != nil {` | `TestReconcileFreshnessWindowCoversAppendAdmission` · `TestReconcileRechecksTheRetentionAgeImmediatelyBeforeAppend` · `TestReconcileRefusesWhenThePairExceedsTheFreshnessBound` | 해당 없음 — 편집 전 기준선 | 기준선 통과(측정, 아래 「측정」) |
| B17 | if 줄 251: `if !sameReconcileMultiset(basisRows, second.basis()) {` | `TestReconcileRefusesARowAppearingInOnlyOneRead` · `TestReconcileRefusesAStatusTransitionBetweenTheTwoReads` | 해당 없음 — 편집 전 기준선 | 기준선 통과(측정, 아래 「측정」) |
| B18 | if 줄 257: `if err := checkReconcileRows(first, symbol, p.Market); err != nil {` | `TestReconcileRefusesARowForAnotherMarket` · `TestReconcileRefusesARowForAnotherSymbol` · `TestReconcileRefusesARowMissingRequiredFields` · `TestReconcileRefusesOCORowsInEitherGroup` | 해당 없음 — 편집 전 기준선 | 기준선 통과(측정, 아래 「측정」) |
| B19 | if 줄 260: `if err := checkReconcileRows(second, symbol, p.Market); err != nil {` | `TestReconcileValidatesTheSecondSnapshotToo` | 해당 없음 — 편집 전 기준선 | 기준선 통과(측정, 아래 「측정」) |
| B20 | if 줄 263: `if err := checkReconcileAbsence(first, target); err != nil {` | `TestReconcileRefusesALiveSuccessorInTheOpenConditionalGroup` · `TestReconcileRefusesAnotherIdentifierThatFiredInClosed` · `TestReconcileRefusesAnyOpenPlainOrderOnTheSymbol` · `TestReconcileRefusesTheTargetExpiredInClosedWithThePermanentMessage` 외 1 | 해당 없음 — 편집 전 기준선 | 기준선 통과(측정, 아래 「측정」) |
| B21 | if 줄 266: `if err := checkReconcileClosedStatus(first); err != nil {` | `TestReconcileRefusesAnOpenFamilyClosedStatusEvenIfTheAllowlistListsIt` · `TestReconcileRefusesClosedStatusOutsideTheAllowlist` | 해당 없음 — 편집 전 기준선 | 기준선 통과(측정, 아래 「측정」) |
| B22 | if 줄 273: `if err != nil {` | `TestReconcileRefusesWhileAnotherReconciliationHoldsTheRecordFile` | 해당 없음 — 편집 전 기준선 | 기준선 통과(측정, 아래 「측정」) |
| B23 | if 줄 274: `if errors.Is(err, errRecordLocked) {` | `TestReconcileRefusesWhileAnotherReconciliationHoldsTheRecordFile` | 해당 없음 — 편집 전 기준선 | 기준선 통과(측정, 아래 「측정」) |
| B24 | if 줄 281: `if err != nil {` | 미덮임(시험별 coverage 측정 0) — 잠근 fd 읽기 오류 — 주입 seam 없음 | 해당 없음 — 편집 전 기준선 | 기준선 통과(측정, 아래 「측정」) |
| B25 | if 줄 284: `if _, err := decodeRecordStrict(again); err != nil {` | `TestReconcileRefusesATornFinalLineBeforeAppend` | 해당 없음 — 편집 전 기준선 | 기준선 통과(측정, 아래 「측정」) |
| B26 | if 줄 287: `if sha256.Sum256(again) != fingerprint {` | `TestReconcileFingerprintIsTheFileBytes` · `TestReconcileRefusesARecordThatChangedBeforeAppend` | 해당 없음 — 편집 전 기준선 | 기준선 통과(측정, 아래 「측정」) |
| B27 | if 줄 290: `if len(again) > 0 && again[len(again)-1] != '\n' {` | 미덮임(시험별 coverage 측정 0) — 선택 때 개행 꼬리(B5)와 지문 일치(B26)가 이미 막아 사실상 도달 불가(방어) | 해당 없음 — 편집 전 기준선 | 기준선 통과(측정, 아래 「측정」) |
| B28 | if 줄 301: `if err != nil {` | 미덮임(시험별 coverage 측정 0) — `crypto/rand` 실패만 | 해당 없음 — 편집 전 기준선 | 기준선 통과(측정, 아래 「측정」) |
| B29 | if 줄 304: `if err := checkReconcileWindow(target, start, at); err != nil {` | 미덮임(시험별 coverage 측정 0) — 쓰기 직전 재검 거절 — 행동 시험은 B16 에서 먼저 거절됨. 위치는 구조 핀 `TestReconcileFinalGateSitsAtTheWriteBoundary`·`TestReconcileRechecksTheWindowAtBothContractPoints` 가 고정 | 해당 없음 — 편집 전 기준선 | 기준선 통과(측정, 아래 「측정」) |
| B30 | if 줄 309: `if err := lock.appendLine(line, int64(len(again))); err != nil {` | `TestReconcileDetectsAWriterInterleavingAtTheAppend` · `TestReconcileReportsAFailedSyncLoudly` · `TestReconcileReportsAPartialAppendLoudly` · `TestReconcileVerifiesTheAppendByReadingItBack` | 해당 없음 — 편집 전 기준선 | 기준선 통과(측정, 아래 「측정」) |
| B31 | if 줄 310: `if errors.Is(err, errRecordSizeMoved) {` | 미덮임(시험별 coverage 측정 0) — 잠그지 않는 작성자 끼어들기 — 층을 내린 `TestLockedAppendRefusesWhenTheRecordGrewAfterTheLockedRead` 가 잠금 층에서 잼 | 해당 없음 — 편집 전 기준선 | 기준선 통과(측정, 아래 「측정」) |

## 측정

- 기준선: 구현 base `aec9bde67` worktree, `go test -run '^<Test>$' -coverprofile` 을 시험 하나씩(verifylive 의
  `reconcile_*_test.go`·`record_filled_test.go`·`record_replay_test.go`·`report_test.go`, official 의 `reconcile_reads*_test.go`,
  cmd 의 `verify_reconcile*_test.go` 를 `-tags tossos_testseams` 로) 돌려 279 시험 프로필을 얻고, 분기 줄의 본문 `{` 블록 실행 수가
  0 보다 큰 시험만 적었다. 기준선 스위트는 verifylive 634 · official 455 · cmd(Reconcile 계열) 62 전부 통과.
