# Function Logic Map: `Reconcile`

- Source: `internal/verifylive/reconcile.go` (184-371)
- Qualified function: `Reconcile`
- Revision: `current` (구현 base `aec9bde67`, `source_sha256` 8f7b7cf845cf3dc8…)
- AST evidence: `ast.json` — AST branches 46, 반환 35(줄 190, 193, 203, 210, 216, 223, 226, 229, 241, 244, 250, 259, 264, 268, 272, 278, 281, 287, 291, 298, 301, 306, 310, 313, 322, 324, 329, 332, 335, 338, 357, 360, 366, 368, 370), 호출 이름 37
- Risk scan: `risk-pattern-report.md`
- 이 change 에서의 역할: 판정 파이프라인 본체. a164 편집(acb6a6b36): `p.Expired` 분기 — 정적 거절 Q1↔만료 어휘(B8~B11), 승인 모드(B13), 창 시작 직후 by-id ①(B17·B18), 둘째 목록 뒤 by-id ②(B21·B22), A2 쌍 비교(B26·B27), 판정 선택(B30~B34), 만료 줄 직렬화(B41·B42). 플래그 없음 경로의 호출 순서·거절 코드 불변(1.2 시험 + a121 시험 전부 통과).

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
| B1 | if 줄 186: `if now == nil {` | `now = time.Now` | `now = time.Now` | 미덮임(시험별 coverage 측정 0) — 생산 호출자(cmd)는 항상 `time.Now` 를 줌 |
| B2 | if 줄 189: `if p.Reader == nil {` | 없음 | `return ReconcileResult{}, refuse(RefuseReadError, "reconcile: no read-only official reader was supplied")` | 미덮임(시험별 coverage 측정 0) — 생산 호출자는 항상 리더를 줌 |
| B3 | if 줄 192: `if p.Approve == nil {` | 없음 | `return ReconcileResult{}, refuse(RefuseNotApproved, "reconcile: no human approval step was supplied")` | 미덮임(시험별 coverage 측정 0) — 생산 호출자는 항상 승인 seam 을 줌 |
| B4 | if 줄 196: `if out == nil {` | `out = io.Discard` | `out = io.Discard` | 미덮임(시험별 coverage 측정 0) — 생산 호출자는 단말 출력을 줌 |
| B5 | if 줄 202: `if err != nil {` | 없음 | `return ReconcileResult{}, err` | `TestReconcileRefusesABrokenInnerLine` · `TestReconcileRefusesACompleteFinalLineWithoutNewline` · `TestReconcileRefusesANewerRecordFormat` · `TestReconcileRefusesATornFinalLineAtSelection` 외 1 |
| B6 | if 줄 209: `if err != nil {` | 없음 | `return ReconcileResult{}, err` | `TestExpiryReconcileIsIdempotent` · `TestReconcileIsIdempotent` · `TestReconcileRefusesAnUnresolvedM0Owner` · `TestReconcileRefusesSeveralCandidates` 외 1 |
| B7 | if 줄 215: `if err != nil {` | 없음 | `return ReconcileResult{}, err` | `TestExpiryReconcileStaticRefusalsComeBeforeApproval` · `TestReconcileRefusesACurrentAccountWhoseMaskDiffers` · `TestReconcileRefusesMixedAccountReferences` · `TestReconcileRefusesUnusableAccountReferences` |
| B8 | if 줄 221: `if p.Expired {` | 없음 | `if err := checkReconcileExpiryVocabulary(); err != nil {` | `TestAnOlderBinaryStillSeesTheExpiredArtifactAsOutstanding` · `TestExpiryReconcileAcceptsATerminalChainSibling` · `TestExpiryReconcileAcceptsAnotherIdentifiersExpiredClosedRow` · `TestExpiryReconcileAcceptsTwoStatementsAndAnExpiredClosedRow` 외 15 |
| B9 | else 줄 225: `} else if err := checkReconcileRetention(target, now()); err != nil {` | 없음 | `return ReconcileResult{}, err` | `TestExpiryReconcileDoesNotNeedTheQ1Measurement` · `TestReconcileRefusesAZeroCreatedAtOnTheOutstandingLine` · `TestReconcileRefusesAnArtifactOlderThanTheRetentionBound` · `TestReconcileRefusesAnyEvictionModelButDuration` 외 2 |
| B10 | if 줄 222: `if err := checkReconcileExpiryVocabulary(); err != nil {` | 없음 | `return ReconcileResult{}, err` | `TestExpiryReconcileRefusesWithoutAReceiptBackedVocabulary` |
| B11 | if 줄 225: `} else if err := checkReconcileRetention(target, now()); err != nil {` | 없음 | `return ReconcileResult{}, err` | `TestExpiryReconcileDoesNotNeedTheQ1Measurement` · `TestReconcileRefusesAZeroCreatedAtOnTheOutstandingLine` · `TestReconcileRefusesAnArtifactOlderThanTheRetentionBound` · `TestReconcileRefusesAnyEvictionModelButDuration` 외 2 |
| B12 | if 줄 228: `if reconcileFreshnessBound <= 0 {` | 없음 | `return ReconcileResult{}, refuse(RefuseFreshnessUnfixed, "the Q3 freshness bound is not fixed")` | `TestExpiryReconcileStaticRefusalsComeBeforeApproval` · `TestReconcileRefusesWhileTheFreshnessBoundIsUnfixed` · `TestReconcileStaticRefusalsComeBeforeTheHumanApproval` |
| B13 | if 줄 237: `if p.Expired {` | 없음 | `approval.Mode = ReconcileModeExpired` | `TestAnOlderBinaryStillSeesTheExpiredArtifactAsOutstanding` · `TestExpiryReconcileAcceptsATerminalChainSibling` · `TestExpiryReconcileAcceptsAnotherIdentifiersExpiredClosedRow` · `TestExpiryReconcileAcceptsTwoStatementsAndAnExpiredClosedRow` 외 13 |
| B14 | if 줄 240: `if err := approval.WriteText(out); err != nil {` | 승인 화면 쓰기 실패 | `return ReconcileResult{}, refuse(RefuseApprovalDisplay, "the approval could not be shown to the operator: "+err.Error())` | `TestReconcileRefusesWhenTheApprovalCannotBeShown` |
| B15 | if 줄 243: `if err := p.Approve(ctx, approval); err != nil {` | 없음 | `return ReconcileResult{}, refuse(RefuseNotApproved, "operator approval was not given: "+err.Error())` | `TestExpiryReconcileDeclinedApprovalReadsNothing` · `TestReconcileRefusesWithoutApproval` |
| B16 | if 줄 249: `if err := checkReconcileInstrument(ctx, p.Reader, symbol); err != nil {` | 없음 | `return ReconcileResult{}, err` | `TestReconcileRefusesWhenTheInstrumentControlFails` |
| B17 | if 줄 257: `if p.Expired {` | 없음 | `if byID[0], err = readReconcileByID(ctx, p.Reader, target.ID); err != nil {` | `TestAnOlderBinaryStillSeesTheExpiredArtifactAsOutstanding` · `TestExpiryReconcileAcceptsATerminalChainSibling` · `TestExpiryReconcileAcceptsAnotherIdentifiersExpiredClosedRow` · `TestExpiryReconcileAcceptsTwoStatementsAndAnExpiredClosedRow` 외 12 |
| B18 | if 줄 258: `if byID[0], err = readReconcileByID(ctx, p.Reader, target.ID); err != nil {` | by-id ① 계좌 범위 GET(a164) | `return ReconcileResult{}, err` | `TestExpiryReconcileRefusesEachGuardWithItsOwnCode` |
| B19 | if 줄 263: `if err != nil {` | 없음 | `return ReconcileResult{}, err` | `TestADeleteNotFoundFromEveryListReadDoesNotReconcile` · `TestReconcileRefusesADuplicateRowWithinOneRead` · `TestReconcileRefusesMalformedListResponsesEvenWhenBothReadsAgree` · `TestReconcileRefusesPaginationFaultsInEitherRead` |
| B20 | if 줄 267: `if err != nil {` | 없음 | `return ReconcileResult{}, err` | `TestReconcileRefusesADuplicateRowWithinOneRead` · `TestReconcileRefusesPaginationFaultsInEitherRead` |
| B21 | if 줄 270: `if p.Expired {` | 없음 | `if byID[1], err = readReconcileByID(ctx, p.Reader, target.ID); err != nil {` | `TestAnOlderBinaryStillSeesTheExpiredArtifactAsOutstanding` · `TestExpiryReconcileAcceptsATerminalChainSibling` · `TestExpiryReconcileAcceptsAnotherIdentifiersExpiredClosedRow` · `TestExpiryReconcileAcceptsTwoStatementsAndAnExpiredClosedRow` 외 12 |
| B22 | if 줄 271: `if byID[1], err = readReconcileByID(ctx, p.Reader, target.ID); err != nil {` | by-id ② 계좌 범위 GET(a164) | `return ReconcileResult{}, err` | `TestExpiryReconcileRefusesEachGuardWithItsOwnCode` |
| B23 | if 줄 277: `if err := checkReconcileInstrument(ctx, p.Reader, symbol); err != nil {` | 없음 | `return ReconcileResult{}, err` | `TestReconcileRefusesWhenTheInstrumentControlFails` |
| B24 | if 줄 280: `if err := checkReconcileWindow(target, start, now(), p.Expired); err != nil {` | 없음 | `return ReconcileResult{}, err` | `TestExpiryReconcileWindowStartsBeforeTheFirstByIDRead` · `TestReconcileFreshnessWindowCoversAppendAdmission` · `TestReconcileRechecksTheRetentionAgeImmediatelyBeforeAppend` · `TestReconcileRefusesWhenThePairExceedsTheFreshnessBound` |
| B25 | if 줄 286: `if !sameReconcileMultiset(basisRows, second.basis()) {` | 없음 | `return ReconcileResult{}, refuse(RefuseReadsDiffer, "the two reads disagree as a multiset of (group, id, status, trigger` | `TestReconcileRefusesARowAppearingInOnlyOneRead` · `TestReconcileRefusesAStatusTransitionBetweenTheTwoReads` |
| B26 | if 줄 289: `if p.Expired {` | 없음 | `if err := checkReconcileByIDPair(byID); err != nil {` | `TestAnOlderBinaryStillSeesTheExpiredArtifactAsOutstanding` · `TestExpiryReconcileAcceptsATerminalChainSibling` · `TestExpiryReconcileAcceptsAnotherIdentifiersExpiredClosedRow` · `TestExpiryReconcileAcceptsTwoStatementsAndAnExpiredClosedRow` 외 11 |
| B27 | if 줄 290: `if err := checkReconcileByIDPair(byID); err != nil {` | 없음 | `return ReconcileResult{}, err` | `TestExpiryReconcileRefusesEachGuardWithItsOwnCode` |
| B28 | if 줄 297: `if err := checkReconcileRows(first, symbol, p.Market); err != nil {` | 없음 | `return ReconcileResult{}, err` | `TestReconcileRefusesARowForAnotherMarket` · `TestReconcileRefusesARowForAnotherSymbol` · `TestReconcileRefusesARowMissingRequiredFields` · `TestReconcileRefusesOCORowsInEitherGroup` |
| B29 | if 줄 300: `if err := checkReconcileRows(second, symbol, p.Market); err != nil {` | 없음 | `return ReconcileResult{}, err` | `TestReconcileValidatesTheSecondSnapshotToo` |
| B30 | if 줄 303: `if p.Expired {` | 없음 | `// a164: 생존 판정(A5) → 진술(A3·A4) → 체인(A7) → CLOSED 어휘(A6).` | `TestAnOlderBinaryStillSeesTheExpiredArtifactAsOutstanding` · `TestExpiryReconcileAcceptsATerminalChainSibling` · `TestExpiryReconcileAcceptsAnotherIdentifiersExpiredClosedRow` · `TestExpiryReconcileAcceptsTwoStatementsAndAnExpiredClosedRow` 외 11 |
| B31 | else 줄 308: `} else {` | 없음 | `if err := checkReconcileAbsence(first, target); err != nil {` | `TestReconcileAbsenceGuardsRefuseInTheirContractOrder` · `TestReconcileAcceptsAnExpiredRowOfAnotherIdentifierOnceTheAllowlistHasIt` · `TestReconcileAcceptsExactlyThePageCap` · `TestReconcileAcceptsWellFormedEmptyListsOverHTTP` 외 28 |
| B32 | if 줄 305: `if err := checkReconcileExpiry(first, second, byID, target, entries, p.Market); err != nil {` | 없음 | `return ReconcileResult{}, err` | `TestExpiryReconcileRefusesAnOpenChainSibling` · `TestExpiryReconcileRefusesEachGuardWithItsOwnCode` |
| B33 | if 줄 309: `if err := checkReconcileAbsence(first, target); err != nil {` | 없음 | `return ReconcileResult{}, err` | `TestReconcileAbsenceGuardsRefuseInTheirContractOrder` · `TestReconcileRefusesALiveSuccessorInTheOpenConditionalGroup` · `TestReconcileRefusesAnotherIdentifierThatFiredInClosed` · `TestReconcileRefusesAnyOpenPlainOrderOnTheSymbol` 외 2 |
| B34 | if 줄 312: `if err := checkReconcileClosedStatus(first); err != nil {` | 없음 | `return ReconcileResult{}, err` | `TestReconcileRefusesAnOpenFamilyClosedStatusEvenIfTheAllowlistListsIt` · `TestReconcileRefusesClosedStatusOutsideTheAllowlist` |
| B35 | if 줄 320: `if err != nil {` | 잠금 실패 | `if errors.Is(err, errRecordLocked) {` | `TestReconcileRefusesWhileAnotherReconciliationHoldsTheRecordFile` |
| B36 | if 줄 321: `if errors.Is(err, errRecordLocked) {` | 없음 | `return ReconcileResult{}, refuse(RefuseRecordLocked, "another reconciliation holds this record file (the same file, poss` | `TestReconcileRefusesWhileAnotherReconciliationHoldsTheRecordFile` |
| B37 | if 줄 328: `if err != nil {` | 없음 | `return ReconcileResult{}, refuse(RefuseRecordUnreadable, err.Error())` | 미덮임(시험별 coverage 측정 0) — 잠근 fd 읽기 오류 — 주입 seam 없음 |
| B38 | if 줄 331: `if _, err := decodeRecordStrict(again); err != nil {` | 없음 | `return ReconcileResult{}, err` | `TestReconcileRefusesATornFinalLineBeforeAppend` |
| B39 | if 줄 334: `if sha256.Sum256(again) != fingerprint {` | 없음 | `return ReconcileResult{}, refuse(RefuseRecordChanged, "the record changed between candidate selection and the append")` | `TestReconcileFingerprintIsTheFileBytes` · `TestReconcileRefusesARecordThatChangedBeforeAppend` |
| B40 | if 줄 337: `if len(again) > 0 && again[len(again)-1] != '\n' {` | 없음 | `return ReconcileResult{}, refuse(RefuseRecordTailNoNewline, "the record does not end with a newline")` | 미덮임(시험별 coverage 측정 0) — 선택 때 개행 꼬리와 지문 일치가 이미 막아 사실상 도달 불가(방어) |
| B41 | if 줄 347: `if p.Expired {` | 만료 줄 직렬화(넷째 종결 표지) | `// a164 D5 — 넷째 종결과 만료 근거(목록 multiset + by-id statement 행, 도메인 태그 지문)·출처별 원 status 관측.` | `TestAnOlderBinaryStillSeesTheExpiredArtifactAsOutstanding` · `TestExpiryReconcileAcceptsATerminalChainSibling` · `TestExpiryReconcileAcceptsAnotherIdentifiersExpiredClosedRow` · `TestExpiryReconcileAcceptsTwoStatementsAndAnExpiredClosedRow` 외 9 |
| B42 | else 줄 352: `} else {` | 없음 | `reconciled.ReconciledAbsent = true` | `TestReconcileAcceptsAnExpiredRowOfAnotherIdentifierOnceTheAllowlistHasIt` · `TestReconcileAcceptsExactlyThePageCap` · `TestReconcileAcceptsWellFormedEmptyListsOverHTTP` · `TestReconcileAppendsInOneWriteEndingWithANewline` 외 16 |
| B43 | if 줄 356: `if err != nil {` | 없음 | `return ReconcileResult{}, err` | 미덮임(시험별 coverage 측정 0) — `crypto/rand` 실패만(두 직렬화 공통) |
| B44 | if 줄 359: `if err := checkReconcileWindow(target, start, at, p.Expired); err != nil {` | 없음 | `return ReconcileResult{}, err` | `TestReconcileRefusesAtTheWriteBoundaryGateItself` |
| B45 | if 줄 364: `if err := lock.appendLine(line, int64(len(again))); err != nil {` | 한 번의 쓰기 실패 — 찢긴 줄 가능(오류 문구가 알림) | `if errors.Is(err, errRecordSizeMoved) {` | `TestReconcileDetectsAWriterInterleavingAtTheAppend` · `TestReconcileReportsAFailedSyncLoudly` · `TestReconcileReportsAPartialAppendLoudly` · `TestReconcileVerifiesTheAppendByReadingItBack` |
| B46 | if 줄 365: `if errors.Is(err, errRecordSizeMoved) {` | 없음 | `return ReconcileResult{}, refuse(RefuseRecordChanged, "the record grew between the locked read and the append — another ` | 미덮임(시험별 coverage 측정 0) — 잠그지 않는 작성자 끼어들기 — 층을 내린 `TestLockedAppendRefusesWhenTheRecordGrewAfterTheLockedRead` 가 잠금 층에서 잼 |

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

AST 호출 이름 전체: `ReconcileBasisDigest`, `ReconcileExpiredBasisDigest`, `approval.WriteText`, `at.UTC`, `bindReconcileAccount`, `checkReconcileAbsence`, `checkReconcileByIDPair`, `checkReconcileClosedStatus`, `checkReconcileExpiry`, `checkReconcileExpiryVocabulary`, `checkReconcileInstrument`, `checkReconcileRetention`, `checkReconcileRows`, `checkReconcileWindow`, `decodeRecordStrict`, `encodeReconcileExpiredLine`, `encodeReconcileLine`, `err.Error`, `errors.Is`, `first.basis`, `fmt.Errorf`, `int64`, `len`, `lock.appendLine`, `lock.contents`, `lock.release`, `lockRecordForAppend`, `now`, `p.Approve`, `readReconcileByID`, `readReconcileSet`, `readRecordStrict`, `refuse`, `sameReconcileMultiset`, `second.basis`, `selectReconcileCandidate`, `sha256.Sum256`.

## State mutations and fallbacks

- 기록 파일 끝에 한 줄 추가가 유일한 부작용(B30 이후). 거절은 기록을 바꾸지 않음(`rcRefuse` 가 바이트 비교).
- a164 편집 뒤에도 쓰기 자리는 `appendLine` 하나 — 만료 모드도 같은 잠금·쓰기 경로를 탐(복사 금지).

## Safety conclusion

- Safe edit boundary: `p.Expired` 가 거짓이면 호출 순서·인자·거절 코드가 base 와 같아야 함 — 구조 핀(`TestReconcileRechecksTheWindowAtBothContractPoints`·`TestReconcileFinalGateSitsAtTheWriteBoundary`·`TestReconcileStaticRefusalsComeBeforeTheHumanApproval`)과 a121 행동 시험 전부가 그대로 통과해야 함.
- High-risk impact: yes — 잘못 닫으면 살아 있는 조건주문이 정리·HOLD·노출 상한에서 빠짐. 만료 모드의 생존 판정(A5)은 부재 모드와 같은 보조 함수로만 판정.

## a164 GREEN 재추출

- 편집 전(구현 base `aec9bde67`) 분기 열 길이 31 에서 편집 뒤 46 로. `source_sha256` d62248e0355f… → 8f7b7cf845cf….
- 옛/새 분기 열(kind + 원문 줄)을 difflib 으로 정렬한 결과: equal 옛 B1..B7 = 새 B1..B7 · replace 옛 B8..B8(1) → 새 B8..B11(4) · equal 옛 B9..B9 = 새 B12..B12 · insert 옛 B10..B9(0) → 새 B13..B13(1) · equal 옛 B10..B12 = 새 B14..B16 · insert 옛 B13..B12(0) → 새 B17..B18(2) · equal 옛 B13..B14 = 새 B19..B20 · insert 옛 B15..B14(0) → 새 B21..B22(2) · equal 옛 B15..B15 = 새 B23..B23 · replace 옛 B16..B16(1) → 새 B24..B24(1) · equal 옛 B17..B17 = 새 B25..B25 · insert 옛 B18..B17(0) → 새 B26..B27(2) · equal 옛 B18..B19 = 새 B28..B29 · insert 옛 B20..B19(0) → 새 B30..B32(3) · equal 옛 B20..B27 = 새 B33..B40 · insert 옛 B28..B27(0) → 새 B41..B42(2) · equal 옛 B28..B28 = 새 B43..B43 · replace 옛 B29..B29(1) → 새 B44..B44(1) · equal 옛 B30..B31 = 새 B45..B46.
- 위 표의 B 번호는 **새** AST 기준이고, 시험 칸은 GREEN 코드에서 시험별 프로필을 다시 떠 측정한 값이다(손 재번호 없음).
