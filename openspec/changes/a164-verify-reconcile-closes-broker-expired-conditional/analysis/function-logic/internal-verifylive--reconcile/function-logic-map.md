# Function Logic Map: `Reconcile`

- Source: `internal/verifylive/reconcile.go` (168-316)
- Qualified function: `Reconcile`
- Revision: `current` (구현 base `aec9bde67`, `source_sha256` d62248e0355fb169…)
- AST evidence: `ast.json` — AST branches 31, 반환 30(줄 174, 177, 187, 194, 200, 206, 209, 218, 221, 227, 234, 238, 243, 246, 252, 258, 261, 264, 267, 275, 277, 282, 285, 288, 291, 302, 305, 311, 313, 315), 호출 이름 31
- Risk scan: `risk-pattern-report.md`
- 이 change 에서의 역할: 판정 파이프라인 본체. a164 는 `p.Expired` 일 때만 갈리는 자리 다섯을 더한다(정적 거절의 Q1↔만료 어휘, 승인 모드, 창 시작 직후·둘째 목록 읽기 직후의 by-id 두 번, 판정 보조 선택, 줄 직렬화). 플래그 없음 경로의 호출 순서·거절 코드는 불변이어야 함(design D1).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `p.RecordPath` | cmd 가 `--config-dir` 에서 유도한 경로 | `reconcilePreflight` | 읽기 실패 `record-unreadable`(B5) |
| `p.Reader` | 공식 GET 셋(→ a164 넷)만 노출하는 인터페이스 | `ReconcileReader` 봉인(`reconcile_seal_test.go`) | nil 이면 `read-error`(B2) |
| `p.Approve` | 사람 y/N seam | cmd `verifyReconcileApprove` | nil 이면 `not-approved`(B3) |
| `p.Now` | 시계 | 생산 `time.Now` | nil 이면 `time.Now`(B1) |
| 패키지 상태 `reconcileRetention`·`reconcileFreshnessBound` | 생산 nil · 15 s | `reconcile.go:346`·`:356` | Q1 없음 → `retention-unmeasured`(B8), Q3 ≤0 → `freshness-unfixed`(B9) |

## Branches and early returns

AST 열거 순서 그대로다. 「Required test」 칸은 손으로 고르지 않았다 — 시험 함수 하나씩 `-coverprofile` 을 떠서 분기 본문
블록이 실행된 시험만 적었다(측정 도구 `analysis/measure/` — `pertest.sh` · `btm_cov.py` · `gen_flm.py`).

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | if 줄 170: `if now == nil {` | `now = time.Now` | `now = time.Now` | 미덮임(시험별 coverage 측정 0) — 생산 호출자(cmd)는 항상 `time.Now` 를 줌 |
| B2 | if 줄 173: `if p.Reader == nil {` | 없음 | `return ReconcileResult{}, refuse(RefuseReadError, "reconcile: no read-only official reader was supplied")` | 미덮임(시험별 coverage 측정 0) — 생산 호출자는 항상 리더를 줌 |
| B3 | if 줄 176: `if p.Approve == nil {` | 없음 | `return ReconcileResult{}, refuse(RefuseNotApproved, "reconcile: no human approval step was supplied")` | 미덮임(시험별 coverage 측정 0) — 생산 호출자는 항상 승인 seam 을 줌 |
| B4 | if 줄 180: `if out == nil {` | `out = io.Discard` | `out = io.Discard` | 미덮임(시험별 coverage 측정 0) — 생산 호출자는 단말 출력을 줌 |
| B5 | if 줄 186: `if err != nil {` | 없음 | `return ReconcileResult{}, err` | `TestReconcileRefusesABrokenInnerLine` · `TestReconcileRefusesACompleteFinalLineWithoutNewline` · `TestReconcileRefusesANewerRecordFormat` · `TestReconcileRefusesATornFinalLineAtSelection` 외 1 |
| B6 | if 줄 193: `if err != nil {` | 없음 | `return ReconcileResult{}, err` | `TestReconcileIsIdempotent` · `TestReconcileRefusesAnUnresolvedM0Owner` · `TestReconcileRefusesSeveralCandidates` · `TestReconcileRefusesZeroCandidates` |
| B7 | if 줄 199: `if err != nil {` | 없음 | `return ReconcileResult{}, err` | `TestReconcileRefusesACurrentAccountWhoseMaskDiffers` · `TestReconcileRefusesMixedAccountReferences` · `TestReconcileRefusesUnusableAccountReferences` |
| B8 | if 줄 205: `if err := checkReconcileRetention(target, now()); err != nil {` | 없음 | `return ReconcileResult{}, err` | `TestReconcileRefusesAZeroCreatedAtOnTheOutstandingLine` · `TestReconcileRefusesAnArtifactOlderThanTheRetentionBound` · `TestReconcileRefusesAnyEvictionModelButDuration` · `TestReconcileRefusesWhileTheRetentionMeasurementIsAbsent` 외 1 |
| B9 | if 줄 208: `if reconcileFreshnessBound <= 0 {` | 없음 | `return ReconcileResult{}, refuse(RefuseFreshnessUnfixed, "the Q3 freshness bound is not fixed")` | `TestReconcileRefusesWhileTheFreshnessBoundIsUnfixed` · `TestReconcileStaticRefusalsComeBeforeTheHumanApproval` |
| B10 | if 줄 217: `if err := approval.WriteText(out); err != nil {` | 승인 화면 쓰기 실패 | `return ReconcileResult{}, refuse(RefuseApprovalDisplay, "the approval could not be shown to the operator: "+err.Error())` | `TestReconcileRefusesWhenTheApprovalCannotBeShown` |
| B11 | if 줄 220: `if err := p.Approve(ctx, approval); err != nil {` | 없음 | `return ReconcileResult{}, refuse(RefuseNotApproved, "operator approval was not given: "+err.Error())` | `TestReconcileRefusesWithoutApproval` |
| B12 | if 줄 226: `if err := checkReconcileInstrument(ctx, p.Reader, symbol); err != nil {` | 없음 | `return ReconcileResult{}, err` | `TestReconcileRefusesWhenTheInstrumentControlFails` |
| B13 | if 줄 233: `if err != nil {` | 없음 | `return ReconcileResult{}, err` | `TestADeleteNotFoundFromEveryListReadDoesNotReconcile` · `TestReconcileRefusesADuplicateRowWithinOneRead` · `TestReconcileRefusesMalformedListResponsesEvenWhenBothReadsAgree` · `TestReconcileRefusesPaginationFaultsInEitherRead` |
| B14 | if 줄 237: `if err != nil {` | 없음 | `return ReconcileResult{}, err` | `TestReconcileRefusesADuplicateRowWithinOneRead` · `TestReconcileRefusesPaginationFaultsInEitherRead` |
| B15 | if 줄 242: `if err := checkReconcileInstrument(ctx, p.Reader, symbol); err != nil {` | 없음 | `return ReconcileResult{}, err` | `TestReconcileRefusesWhenTheInstrumentControlFails` |
| B16 | if 줄 245: `if err := checkReconcileWindow(target, start, now()); err != nil {` | 없음 | `return ReconcileResult{}, err` | `TestReconcileFreshnessWindowCoversAppendAdmission` · `TestReconcileRechecksTheRetentionAgeImmediatelyBeforeAppend` · `TestReconcileRefusesWhenThePairExceedsTheFreshnessBound` |
| B17 | if 줄 251: `if !sameReconcileMultiset(basisRows, second.basis()) {` | 없음 | `return ReconcileResult{}, refuse(RefuseReadsDiffer, "the two reads disagree as a multiset of (group, id, status, trigger` | `TestReconcileRefusesARowAppearingInOnlyOneRead` · `TestReconcileRefusesAStatusTransitionBetweenTheTwoReads` |
| B18 | if 줄 257: `if err := checkReconcileRows(first, symbol, p.Market); err != nil {` | 없음 | `return ReconcileResult{}, err` | `TestReconcileRefusesARowForAnotherMarket` · `TestReconcileRefusesARowForAnotherSymbol` · `TestReconcileRefusesARowMissingRequiredFields` · `TestReconcileRefusesOCORowsInEitherGroup` |
| B19 | if 줄 260: `if err := checkReconcileRows(second, symbol, p.Market); err != nil {` | 없음 | `return ReconcileResult{}, err` | `TestReconcileValidatesTheSecondSnapshotToo` |
| B20 | if 줄 263: `if err := checkReconcileAbsence(first, target); err != nil {` | 없음 | `return ReconcileResult{}, err` | `TestReconcileRefusesALiveSuccessorInTheOpenConditionalGroup` · `TestReconcileRefusesAnotherIdentifierThatFiredInClosed` · `TestReconcileRefusesAnyOpenPlainOrderOnTheSymbol` · `TestReconcileRefusesTheTargetExpiredInClosedWithThePermanentMessage` 외 1 |
| B21 | if 줄 266: `if err := checkReconcileClosedStatus(first); err != nil {` | 없음 | `return ReconcileResult{}, err` | `TestReconcileRefusesAnOpenFamilyClosedStatusEvenIfTheAllowlistListsIt` · `TestReconcileRefusesClosedStatusOutsideTheAllowlist` |
| B22 | if 줄 273: `if err != nil {` | 잠금 실패 | `if errors.Is(err, errRecordLocked) {` | `TestReconcileRefusesWhileAnotherReconciliationHoldsTheRecordFile` |
| B23 | if 줄 274: `if errors.Is(err, errRecordLocked) {` | 없음 | `return ReconcileResult{}, refuse(RefuseRecordLocked, "another reconciliation holds this record file (the same file, poss` | `TestReconcileRefusesWhileAnotherReconciliationHoldsTheRecordFile` |
| B24 | if 줄 281: `if err != nil {` | 없음 | `return ReconcileResult{}, refuse(RefuseRecordUnreadable, err.Error())` | 미덮임(시험별 coverage 측정 0) — 잠근 fd 읽기 오류 — 주입 seam 없음 |
| B25 | if 줄 284: `if _, err := decodeRecordStrict(again); err != nil {` | 없음 | `return ReconcileResult{}, err` | `TestReconcileRefusesATornFinalLineBeforeAppend` |
| B26 | if 줄 287: `if sha256.Sum256(again) != fingerprint {` | 없음 | `return ReconcileResult{}, refuse(RefuseRecordChanged, "the record changed between candidate selection and the append")` | `TestReconcileFingerprintIsTheFileBytes` · `TestReconcileRefusesARecordThatChangedBeforeAppend` |
| B27 | if 줄 290: `if len(again) > 0 && again[len(again)-1] != '\n' {` | 없음 | `return ReconcileResult{}, refuse(RefuseRecordTailNoNewline, "the record does not end with a newline")` | 미덮임(시험별 coverage 측정 0) — 선택 때 개행 꼬리(B5)와 지문 일치(B26)가 이미 막아 사실상 도달 불가(방어) |
| B28 | if 줄 301: `if err != nil {` | 없음 | `return ReconcileResult{}, err` | 미덮임(시험별 coverage 측정 0) — `crypto/rand` 실패만 |
| B29 | if 줄 304: `if err := checkReconcileWindow(target, start, at); err != nil {` | 없음 | `return ReconcileResult{}, err` | 미덮임(시험별 coverage 측정 0) — 쓰기 직전 재검 거절 — 행동 시험은 B16 에서 먼저 거절됨. 위치는 구조 핀 `TestReconcileFinalGateSitsAtTheWriteBoundary`·`TestReconcileRechecksTheWindowAtBothContractPoints` 가 고정 |
| B30 | if 줄 309: `if err := lock.appendLine(line, int64(len(again))); err != nil {` | 한 번의 쓰기 실패 — 찢긴 줄 가능(오류 문구가 알림) | `if errors.Is(err, errRecordSizeMoved) {` | `TestReconcileDetectsAWriterInterleavingAtTheAppend` · `TestReconcileReportsAFailedSyncLoudly` · `TestReconcileReportsAPartialAppendLoudly` · `TestReconcileVerifiesTheAppendByReadingItBack` |
| B31 | if 줄 310: `if errors.Is(err, errRecordSizeMoved) {` | 없음 | `return ReconcileResult{}, refuse(RefuseRecordChanged, "the record grew between the locked read and the append — another ` | 미덮임(시험별 coverage 측정 0) — 잠그지 않는 작성자 끼어들기 — 층을 내린 `TestLockedAppendRefusesWhenTheRecordGrewAfterTheLockedRead` 가 잠금 층에서 잼 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `readRecordStrict`·`decodeRecordStrict` | 엄격 해독·지문 | 실패 = 거절, 재시도 없음 | AST |
| `selectReconcileCandidate` | 후보 정확히 하나 | 0·복수·M0 = 거절 | AST |
| `bindReconcileAccount` | 계좌 결속 | mixed → mismatch 순 | AST |
| `checkReconcileRetention` | Q1 정적 거절(승인 앞) | 생산 nil = 거절 | AST — a164: 만료 모드는 부르지 않음(D7) |
| `checkReconcileInstrument` | 양성 대조 두 번 | 오류·다른 심볼 = 거절 | AST |
| `readReconcileSet` | 세 그룹 두 번 읽기 | 오류 = `read-error` 등, 자동 재시도 없음 | AST — a164 는 그대로 부름(P2-5 ⑤) |
| `checkReconcileWindow` | Q3 창 + Q1 나이 재검(두 번) | 초과 = 거절 | AST — a164: 모드 인자로 Q1 재검 생략 |
| `checkReconcileAbsence`·`checkReconcileClosedStatus` | 부재 판정 | 거절 | AST — a164: 만료 모드는 `checkReconcileExpiry` |
| `lockRecordForAppend`·`contents`·`appendLine` | 잠금·재읽기·한 번의 쓰기 | 잠김·크기 이동 = 거절, 그 밖 오류 = 큰 소리 | AST |

AST 호출 이름 전체: `ReconcileBasisDigest`, `approval.WriteText`, `at.UTC`, `bindReconcileAccount`, `checkReconcileAbsence`, `checkReconcileClosedStatus`, `checkReconcileInstrument`, `checkReconcileRetention`, `checkReconcileRows`, `checkReconcileWindow`, `decodeRecordStrict`, `encodeReconcileLine`, `err.Error`, `errors.Is`, `first.basis`, `fmt.Errorf`, `int64`, `len`, `lock.appendLine`, `lock.contents`, `lock.release`, `lockRecordForAppend`, `now`, `p.Approve`, `readReconcileSet`, `readRecordStrict`, `refuse`, `sameReconcileMultiset`, `second.basis`, `selectReconcileCandidate`, `sha256.Sum256`.

## State mutations and fallbacks

- 기록 파일 끝에 한 줄 추가가 유일한 부작용(B30 이후). 거절은 기록을 바꾸지 않음(`rcRefuse` 가 바이트 비교).
- a164 편집 뒤에도 쓰기 자리는 `appendLine` 하나 — 만료 모드도 같은 잠금·쓰기 경로를 탐(복사 금지).

## Safety conclusion

- Safe edit boundary: `p.Expired` 가 거짓이면 호출 순서·인자·거절 코드가 base 와 같아야 함 — 구조 핀(`TestReconcileRechecksTheWindowAtBothContractPoints`·`TestReconcileFinalGateSitsAtTheWriteBoundary`·`TestReconcileStaticRefusalsComeBeforeTheHumanApproval`)과 a121 행동 시험 전부가 그대로 통과해야 함.
- High-risk impact: yes — 잘못 닫으면 살아 있는 조건주문이 정리·HOLD·노출 상한에서 빠짐. 만료 모드의 생존 판정(A5)은 부재 모드와 같은 보조 함수로만 판정.
