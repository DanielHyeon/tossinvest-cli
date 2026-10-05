# a121 GREEN 변이 원장

- TREE 스탬프(사본 = 작업 트리, .go·go.mod·go.sum sha256): `4f240dec0054684197bace2f43aa43ca3f1092986db5eae87a81a182cc082784` — 작업 트리 재계산과 일치.
- 하네스: `mutation/mutate.py`(저장소 뿌리를 파일 위치에서 유도; 사본 트리에서만 변이 → 패키지 시험 → 원복, 끝에 스탬프 재대조). 변이 목록 `mutation/mutants.json`, 원장 원문 `mutation/ledger.json`.
- 대조군(무변이): verifylive(census 포함 전 시험)·official·cmd(tagged) 전부 PASS.
- 결과: **변이 130 — CAUGHT 127 · SURVIVED 3(전부 동등 처분, 결속 가드 명기) · BUILD-FAIL 0.** (4.2 codex 수리 뒤: CG-1~CG-6 가드 변이 20 추가 — 전부 CAUGHT)
- 경위: ledger 1·2 의 BUILD-FAIL 은 변이 정의 결함(미사용 변수 등)이라 고쳐 다시 돌렸다. 생존 V-cap 은 생산의 중복 분기를 지워 없앴고, 생존 O-*·V-status-openfamily·V-projection-monotone 은 새 GREEN 시험으로 잡았다. V-final-window·V-window-before-instr2 는 Manager 판정(동등 수용 + 두 재검 호출 자리 AST 핀 `TestReconcileRechecksTheWindowAtBothContractPoints`)으로 이 핀이 잡는다 — 결과값만 무력화하는(`&& false`) 형은 두 재검이 서로를 가리는 같은 동등 부류다. CG2-size-check 는 층을 내린 단위 시험 `TestLockedAppendRefusesWhenTheRecordGrewAfterTheLockedRead` 로 잡았다. A-GREEN P2 의 GO2·MJ10 포함.

## 생존 3 처분

| 변이 | 무엇 | 처분 |
|---|---|---|
| V-tail-append | 추가 직전 개행 검사 제거 | **동등** — 선택 전 개행 꼬리 강제(V-tail-pre CAUGHT) + 잠근 fd 로 읽은 바이트의 지문이 선택 때와 같아야 함(V-fingerprint CAUGHT) ⇒ 판정 바이트가 같으므로 개행 꼬리는 지문이 함의(sha256 충돌 제외). 결속: `TestReconcileRefusesACompleteFinalLineWithoutNewline` + `TestReconcileFingerprintIsTheFileBytes`. |
| C-rebind | 명시 재결속 제거(새 클라이언트 지연 해석) | **동등(CG-6 수리 뒤)** — 명시 재결속(`WithAccountSeq`)을 빼면 읽기용 클라이언트가 암묵 해석하는데, 첫 계좌 범위 읽기 전 신원 재확인(`ensureIdentity`)이 그 클라이언트의 `Accounts()` 를 먼저 부르고(첫 행 seq 를 캐시) 계좌가 하나·계좌번호·seq 가 검증값과 같음을 요구하므로 헤더 seq 는 검증한 seq 와 같다. 결속: `TestReconcileReaderRechecksIdentityBeforeItsFirstScopedRead`(CG6-seq-ignored·CG6-count-ignored CAUGHT) + `TestReconcileNarrowReaderBindsTheValidatedSequence`(헤더 5·/accounts 2). 측정 한계 ⑦(암묵 캐시 vs 명시 재결속 동일 헤더)의 연장. |
| C-env-creds | 자격을 환경 변수에서도(preflight 가 덮음) | **동등(명령 경로)** — Manager 판정 2026-10-05. 생산 팩토리는 preflight 뒤에만 불리고 preflight 는 환경 변수 자격을 하나라도 거절. 결속: `TestReconcilePreflightRefusals`(env 3사례) + 순서 핀 `TestReconcileRefusesPreflightBeforeAnyLockOrRead`. 팩토리를 env 와 함께 단독 호출하는 시험은 변이 시 실 API 로 나갈 수 있어 두지 않음(안전 불변식 1). |

## 전 변이

| 변이 | 파일 | 무엇 | 결과 | 잡은 시험(앞부분) |
|---|---|---|---|---|
| V-terminal-third | `internal/verifylive/record.go` | 셋째 종결 제거 | CAUGHT | TestReconcileAppendsOneReconciledAbsentLineWhenEveryConditionHolds,TestReconciledProjectionIsMonotoneLikeOutst |
| V-buildreport-reconciled | `internal/verifylive/report.go` | BuildReport 칸 제거 | CAUGHT | TestReconcileTextLabelsReconciledAbsentInReportAndStatus |
| V-buildprogress-reconciled | `internal/verifylive/report.go` | BuildProgress 칸 제거 | CAUGHT | TestReconcileTextLabelsReconciledAbsentInReportAndStatus |
| V-report-writetext | `internal/verifylive/report.go` | Report.WriteText 출력 제거 | CAUGHT | TestReconcileTextLabelsReconciledAbsentInReportAndStatus |
| V-progress-writetext | `internal/verifylive/report.go` | Progress.WriteText 출력 제거 | CAUGHT | TestReconcileTextLabelsReconciledAbsentInReportAndStatus |
| V-label | `internal/verifylive/report.go` | 라벨을 취소로 | CAUGHT | TestReconcileTextLabelsReconciledAbsentInReportAndStatus |
| V-projection-monotone | `internal/verifylive/projection_reconciled.go` | 투영 단조성 제거 | CAUGHT | TestReconciledProjectionIsMonotoneLikeOutstanding |
| V-marshal-plain | `internal/verifylive/reconcile_record.go` | 대사 줄을 일반 모양으로(영 시각 cancelled_at) | CAUGHT | TestReconcileLineMatchesTheDesignShape,TestReconcileLineCarriesNoHoldEvenWhenTheOutstandingLineHadOne |
| V-marshal-all | `internal/verifylive/reconcile_record.go` | 모든 artifact 를 축약형으로 | CAUGHT | TestAbortEndsAHeldChain,TestAbortSendsNothingWhenItIsNotApproved,TestALeftoverOrderIsCancelledOnTheNextRun,Tes |
| V-tail-pre | `internal/verifylive/reconcile_record.go` | 선택 전 개행 검사 제거 | CAUGHT | TestReconcileRefusesACompleteFinalLineWithoutNewline |
| V-undecodable | `internal/verifylive/reconcile_record.go` | 엄격 해독 제거 | CAUGHT | TestReconcileRefusesATornFinalLineAtSelection,TestReconcileRefusesABrokenInnerLine,TestReconcileRefusesATornFi |
| V-format | `internal/verifylive/reconcile_record.go` | 형식 판정 제거 | CAUGHT | TestReconcileRefusesANewerRecordFormat |
| V-unreadable | `internal/verifylive/reconcile_record.go` | 읽기 오류 무시 | CAUGHT | TestReconcileRefusesAnUnreadableRecord |
| V-fingerprint | `internal/verifylive/reconcile.go` | 지문 비교 제거 | CAUGHT | TestReconcileRefusesARecordThatChangedBeforeAppend,TestReconcileFingerprintIsTheFileBytes |
| V-tail-append | `internal/verifylive/reconcile.go` | 추가 직전 개행 검사 제거 | SURVIVED | — |
| V-final-window | `internal/verifylive/reconcile.go` | 최종(쓰기 경계) 재검 호출 제거 | CAUGHT | TestReconcileFinalGateSitsAtTheWriteBoundary,TestReconcileRechecksTheWindowAtBothContractPoints |
| V-append-decode | `internal/verifylive/reconcile.go` | 추가 직전 엄격 해독 생략 | CAUGHT | TestReconcileRefusesATornFinalLineBeforeAppend |
| V-select-kind | `internal/verifylive/reconcile_check.go` | 조건주문 종류 거름 제거 | CAUGHT | TestReconcileRefusesZeroCandidates |
| V-select-many | `internal/verifylive/reconcile_check.go` | 복수 후보 허용 | CAUGHT | TestReconcileRefusesSeveralCandidates |
| V-m0-filter | `internal/verifylive/reconcile_check.go` | M0 제외 제거 | CAUGHT | TestWithoutM0UnsettledDropsTheNamedParentAtItsOwnLayer |
| V-m0-error | `internal/verifylive/reconcile_check.go` | M0 오류 무시 | CAUGHT | TestReconcileRefusesAnUnresolvedM0Owner,TestWithoutM0UnsettledRefusesAnAmbiguousOwner |
| V-mixed | `internal/verifylive/reconcile_check.go` | mixed 거절 제거 | CAUGHT | TestReconcileRefusesMixedAccountReferences |
| V-unusable | `internal/verifylive/reconcile_check.go` | 마스크 경계 제거 | CAUGHT | TestReconcileRefusesUnusableAccountReferences |
| V-mismatch | `internal/verifylive/reconcile_check.go` | mismatch 제거 | CAUGHT | TestReconcileRefusesACurrentAccountWhoseMaskDiffers |
| V-approve-ignore | `internal/verifylive/reconcile.go` | 승인 거절 무시 | CAUGHT | TestReconcileRefusesWithoutApproval |
| V-approve-text | `internal/verifylive/reconcile.go` | 승인 출력 생략 | CAUGHT | TestReconcileRefusesWhenTheApprovalCannotBeShown,TestReconcileApprovalShowsBothMasksAndTheAccountCountBeforeAn |
| V-approval-market | `internal/verifylive/reconcile.go` | 승인 시장 생략 | CAUGHT | TestReconcileApprovalShowsBothMasksAndTheAccountCountBeforeAnyListRead |
| V-q1-nil | `internal/verifylive/reconcile_check.go` | 측정 부재를 무한 한도로 | CAUGHT | TestReconcileRefusesWhileTheRetentionMeasurementIsAbsent,TestReconcileStaticRefusalsComeBeforeTheHumanApproval |
| V-q1-eviction | `internal/verifylive/reconcile_check.go` | 축출 모형 판정 약화 | CAUGHT | TestReconcileRefusesAnyEvictionModelButDuration |
| V-q1-zero | `internal/verifylive/reconcile_check.go` | 영 CreatedAt 거절 제거 | CAUGHT | TestReconcileRefusesAZeroCreatedAtOnTheOutstandingLine,TestReconcileStaticRefusalsComeBeforeTheHumanApproval |
| V-q1-age | `internal/verifylive/reconcile_check.go` | 나이 한도 제거 | CAUGHT | TestReconcileRefusesAnArtifactOlderThanTheRetentionBound,TestReconcileRechecksTheRetentionAgeImmediatelyBefore |
| V-q3-unfixed | `internal/verifylive/reconcile.go` | Q3 미확정 거절 제거 | CAUGHT | TestReconcileRefusesWhileTheFreshnessBoundIsUnfixed,TestReconcileStaticRefusalsComeBeforeTheHumanApproval |
| V-q3-exceed | `internal/verifylive/reconcile_check.go` | 신선도 한도 제거 | CAUGHT | TestReconcileRefusesWhenThePairExceedsTheFreshnessBound,TestReconcileFreshnessWindowCoversAppendAdmission |
| V-q3-start-early | `internal/verifylive/reconcile.go` | 순서 교란: 창 시작을 승인 전으로 | CAUGHT | TestReconcileApprovalWaitDoesNotConsumeTheFreshnessWindow |
| V-instr-echo | `internal/verifylive/reconcile_check.go` | 에코 대조 제거 | CAUGHT | TestReconcileRefusesWhenTheInstrumentControlFails |
| V-instr-second | `internal/verifylive/reconcile.go` | 둘째 종목 조회 제거 | CAUGHT | TestReconcileReadsTheGroupsInTheFixedOrder,TestReconcileRechecksTheRetentionAgeImmediatelyBeforeAppend,TestRec |
| V-window-before-instr2 | `internal/verifylive/reconcile.go` | 순서 교란: 재검을 둘째 조회 앞으로(최종 재검은 유지) | CAUGHT | TestReconcileRechecksTheWindowAtBothContractPoints |
| V-second-read | `internal/verifylive/reconcile.go` | 둘째 읽기 생략 | CAUGHT | TestReconcileValidatesTheSecondSnapshotToo,TestReconcileReadsTheGroupsInTheFixedOrder,TestReconcileRefusesWhen |
| V-read-order | `internal/verifylive/reconcile_check.go` | 읽기 순서 교란 | CAUGHT | TestReconcileValidatesTheSecondSnapshotToo,TestReconcileRefusesWhileAnotherReconciliationHoldsTheRecordFile,Te |
| V-limit | `internal/verifylive/reconcile_check.go` | limit 변경 | CAUGHT | TestReconcileReadsTheGroupsInTheFixedOrder |
| V-repeated | `internal/verifylive/reconcile_check.go` | 반복 커서 제거 | CAUGHT | TestReconcileRefusesPaginationFaultsInEitherRead |
| V-dup | `internal/verifylive/reconcile_check.go` | 읽기 안 중복 제거 | CAUGHT | TestReconcileRefusesADuplicateRowWithinOneRead |
| V-empty-cursor | `internal/verifylive/reconcile_check.go` | 빈 커서 제거 | CAUGHT | TestReconcileRefusesPaginationFaultsInEitherRead |
| V-readerr | `internal/verifylive/reconcile_check.go` | 조건주문 읽기 오류 무시 | CAUGHT | TestReconcileRefusesPaginationFaultsInEitherRead,TestReconcileRefusesMalformedListResponsesEvenWhenBothReadsAg |
| V-readerr-plain | `internal/verifylive/reconcile_check.go` | 일반 주문 읽기 오류 무시 | CAUGHT | TestReconcileRefusesPaginationFaultsInEitherRead,TestReconcileRefusesMalformedListResponsesEvenWhenBothReadsAg |
| V-multiset | `internal/verifylive/reconcile.go` | multiset 비교 제거 | CAUGHT | TestReconcileRefusesAStatusTransitionBetweenTheTwoReads,TestReconcileRefusesARowAppearingInOnlyOneRead |
| V-multiset-set | `internal/verifylive/reconcile_check.go` | 길이 비교 제거 | CAUGHT | TestReconcileRefusesARowAppearingInOnlyOneRead |
| V-oco | `internal/verifylive/reconcile_check.go` | OCO 거절 제거 | CAUGHT | TestReconcileValidatesTheSecondSnapshotToo,TestReconcileRefusesOCORowsInEitherGroup |
| V-incomplete | `internal/verifylive/reconcile_check.go` | 필드 결측 제거 | CAUGHT | TestReconcileValidatesTheSecondSnapshotToo,TestReconcileRefusesARowMissingRequiredFields |
| V-symbol | `internal/verifylive/reconcile_check.go` | 심볼 대조 제거 | CAUGHT | TestReconcileValidatesTheSecondSnapshotToo,TestReconcileRefusesARowForAnotherSymbol |
| V-market | `internal/verifylive/reconcile_check.go` | 시장 대조 제거 | CAUGHT | TestReconcileValidatesTheSecondSnapshotToo,TestReconcileRefusesARowForAnotherMarket |
| V-rows-after-absence | `internal/verifylive/reconcile.go` | 순서 교란: 행 검사를 부재 뒤로 | CAUGHT | TestReconcileRefusesOCORowsInEitherGroup |
| V-open-cond | `internal/verifylive/reconcile_check.go` | OPEN 조건주문 거절 제거 | CAUGHT | TestReconcileRefusesALiveSuccessorInTheOpenConditionalGroup |
| V-open-plain | `internal/verifylive/reconcile_check.go` | OPEN 일반 주문 거절 제거 | CAUGHT | TestReconcileRefusesAnyOpenPlainOrderOnTheSymbol |
| V-q6 | `internal/verifylive/reconcile_check.go` | Q6 분기 제거 | CAUGHT | TestReconcileRefusesTheTargetExpiredInClosedWithThePermanentMessage |
| V-target | `internal/verifylive/reconcile_check.go` | CLOSED target 검사 제거 | CAUGHT | TestReconcileRefusesTheTargetFoundInClosed,TestReconcileRefusesTheTargetExpiredInClosedWithThePermanentMessage |
| V-fired | `internal/verifylive/reconcile_check.go` | COMPLETED 흔적 제거 | CAUGHT | TestReconcileRefusesAnotherIdentifierThatFiredInClosed |
| V-fired-trigger | `internal/verifylive/reconcile_check.go` | triggeredOrderId 흔적 제거 | CAUGHT | TestReconcileRefusesAnotherIdentifierThatFiredInClosed |
| V-status-allowlist | `internal/verifylive/reconcile_check.go` | 허용 목록 무시 | CAUGHT | TestReconcileRefusesClosedStatusOutsideTheAllowlist |
| V-status-openfamily | `internal/verifylive/reconcile_check.go` | OPEN 계열 명시 거절 제거 | CAUGHT | TestReconcileRefusesAnOpenFamilyClosedStatusEvenIfTheAllowlistListsIt |
| V-status-before-absence | `internal/verifylive/reconcile.go` | 순서 교란: 허용 목록을 부재 앞으로(Q6 가림) | CAUGHT | TestReconcileRefusesTheTargetFoundInClosed,TestReconcileRefusesTheTargetExpiredInClosedWithThePermanentMessage |
| V-basis-empty | `internal/verifylive/reconcile.go` | 근거를 빈 집합으로 | CAUGHT | TestReconcileAcceptsAnExpiredRowOfAnotherIdentifierOnceTheAllowlistHasIt |
| V-basis-sort | `internal/verifylive/reconcile_record.go` | 근거 정렬 제거 | CAUGHT | TestReconcileBasisDigestIsVersionedDomainTaggedAndOrderIndependent |
| V-basis-domain | `internal/verifylive/reconcile_record.go` | 도메인 분리 제거 | CAUGHT | TestReconcileBasisDigestIsVersionedDomainTaggedAndOrderIndependent |
| V-line-createdat | `internal/verifylive/reconcile.go` | CreatedAt 를 추가 시각으로 | CAUGHT | TestReconcileLineMatchesTheDesignShape |
| V-line-step | `internal/verifylive/reconcile_record.go` | StepID 를 카탈로그 값으로 | CAUGHT | TestReconcileLineMatchesTheDesignShape |
| V-line-mutating | `internal/verifylive/reconcile_record.go` | mutating true | CAUGHT | TestReconcileLineMatchesTheDesignShape |
| V-line-calls | `internal/verifylive/reconcile_record.go` | Calls 탑재 | CAUGHT | TestReconcileLineMatchesTheDesignShape,TestReconcileLineNeverBecomesEndpointEvidence |
| O-not-object | `internal/official/reconcile_reads.go` | result 객체 검사 제거 | CAUGHT | TestDecodeReconcilePageNamesTheGuardThatRefused |
| O-collection-missing | `internal/official/reconcile_reads.go` | 컬렉션 결측 검사 제거 | CAUGHT | TestDecodeReconcilePageNamesTheGuardThatRefused |
| O-collection-array | `internal/official/reconcile_reads.go` | 컬렉션 배열 검사 제거 | CAUGHT | TestDecodeReconcilePageNamesTheGuardThatRefused,TestDecodeReconcileConditionalPageRefusesSchemaHoles,TestDecod |
| O-row-object | `internal/official/reconcile_reads.go` | 행 객체 검사 제거 | CAUGHT | TestDecodeReconcilePageNamesTheGuardThatRefused,TestDecodeReconcileConditionalPageRefusesSchemaHoles,TestDecod |
| O-hasnext-missing | `internal/official/reconcile_reads.go` | hasNext 결측 검사 제거 | CAUGHT | TestDecodeReconcilePageNamesTheGuardThatRefused |
| O-hasnext-type | `internal/official/reconcile_reads.go` | hasNext 형 검사 제거 | CAUGHT | TestDecodeReconcilePageNamesTheGuardThatRefused,TestDecodeReconcilePagesAcceptWellFormedPages,TestDecodeReconc |
| O-cursor-missing | `internal/official/reconcile_reads.go` | nextCursor 결측 검사 제거 | CAUGHT | TestDecodeReconcilePageNamesTheGuardThatRefused |
| O-cursor-type | `internal/official/reconcile_reads.go` | nextCursor 형 검사 제거 | CAUGHT | TestDecodeReconcilePageNamesTheGuardThatRefused |
| O-cursor-consistency | `internal/official/reconcile_reads.go` | 커서 일관성 제거 | CAUGHT | TestDecodeReconcilePageNamesTheGuardThatRefused,TestDecodeReconcileConditionalPageRefusesSchemaHoles,TestDecod |
| O-second | `internal/official/reconcile_reads.go` | second 노출 제거 | CAUGHT | TestDecodeReconcileConditionalPageKeepsTheSecondLegAndTheTrigger |
| O-trigger | `internal/official/reconcile_reads.go` | triggeredOrderId 누락 | CAUGHT | TestDecodeReconcileConditionalPageKeepsTheSecondLegAndTheTrigger |
| O-plain-status | `internal/official/reconcile_reads.go` | 일반 주문 해독 우회(빈 페이지) | CAUGHT | TestReconcileClientReadsRefuseANullResult |
| O-cond-bypass | `internal/official/reconcile_reads.go` | 조건주문 해독 우회 | CAUGHT | TestReconcileClientReadsRefuseANullResult |
| O-status-query | `internal/official/reconcile_reads.go` | 일반 주문 OPEN 필터 누락 | CAUGHT | TestReconcileClientReadsSendTheDocumentedGetQueries |
| O-instrument-echo | `internal/official/reconcile_reads.go` | 에코 대신 입력 반환 | CAUGHT | TestReconcileInstrumentReturnsTheBrokersEchoNotItsInput |
| C-noconfig | `cmd/tossctl/verify_reconcile.go` | config-dir 필수 제거 | CAUGHT | TestReconcilePreflightRefusals |
| C-env-either | `cmd/tossctl/verify_reconcile.go` | 환경 변수 반쪽 허용 | CAUGHT | TestReconcilePreflightRefusals,TestReconcileRefusesPreflightBeforeAnyLockOrRead |
| C-record | `cmd/tossctl/verify_reconcile.go` | --record 거절 제거 | CAUGHT | TestReconcilePreflightRefusals |
| C-market | `cmd/tossctl/verify_reconcile.go` | --market 필수 제거 | CAUGHT | TestReconcilePreflightRefusals |
| C-lock | `cmd/tossctl/verify_reconcile.go` | flock 생략 | CAUGHT | TestReconcileShowsTheAccountsOnTheSameChannelAsTheQuestion,TestReconcileHoldsTheExecutionLockAndRateBudgetThro |
| C-lease | `cmd/tossctl/verify_reconcile.go` | lease 생략 | CAUGHT | TestReconcileShowsTheAccountsOnTheSameChannelAsTheQuestion,TestReconcileHoldsTheExecutionLockAndRateBudgetThro |
| C-lock-after-factory | `cmd/tossctl/verify_reconcile.go` | 순서 교란: 잠금 전 계좌 읽기 | CAUGHT | TestReconcileRefusesWhenTheExecutionLockIsHeld,TestReconcileRefusesWhenTheRateBudgetLeaseIsHeld |
| C-malformed | `cmd/tossctl/verify_reconcile.go` | 기형 행 거절 제거 | CAUGHT | TestReconcileNarrowReaderRefusesAccountShapes |
| C-count | `cmd/tossctl/verify_reconcile.go` | 복수 계좌 허용 | CAUGHT | TestReconcileNarrowReaderRefusesAccountShapes |
| C-seq | `cmd/tossctl/verify_reconcile.go` | seq 0 허용 | CAUGHT | TestReconcileNarrowReaderRefusesAccountShapes |
| C-rebind | `cmd/tossctl/verify_reconcile.go` | 명시 재결속 제거(새 클라이언트 지연 해석) | SURVIVED | — |
| C-rebind-reuse | `cmd/tossctl/verify_reconcile.go` | 명시 재결속 대신 암묵 캐시(측정 한계) | CAUGHT | TestReconcileNarrowReaderBindsTheValidatedSequence |
| C-terminal | `cmd/tossctl/verify_reconcile.go` | 기본 단말을 대화형으로 | CAUGHT | TestReconcileDefaultTerminalIsNotInteractiveOnAPipeOrDevNull |
| C-terminal-chardev | `cmd/tossctl/verify_reconcile.go` | ModeCharDevice 판정 | CAUGHT | TestReconcileDefaultTerminalIsNotInteractiveOnAPipeOrDevNull |
| C-approve-noninteractive | `cmd/tossctl/verify_reconcile.go` | 비대화형 승인 허용 | CAUGHT | TestReconcileDefaultApprovalStopsTheCommandBeforeAnyListRead,TestReconcileDefaultApprovalRefusesANonInteractiv |
| C-approve-any | `cmd/tossctl/verify_reconcile.go` | 아무 입력이나 승인 | CAUGHT | TestReconcileDefaultApprovalAcceptsOnlyAnExplicitInteractiveYes |
| C-approve-default-yes | `cmd/tossctl/verify_reconcile.go` | 빈 입력을 승인(기본 Y) | CAUGHT | TestReconcileDefaultApprovalAcceptsOnlyAnExplicitInteractiveYes |
| C-approve-wired | `cmd/tossctl/verify_reconcile.go` | 명령이 기본 승인을 우회 | CAUGHT | TestReconcileShowsTheAccountsOnTheSameChannelAsTheQuestion,TestReconcileDefaultApprovalStopsTheCommandBeforeAn |
| C-register | `cmd/tossctl/verify.go` | 명령 미등록 | CAUGHT | TestVerifyReconcileIsRegisteredAsAMutatingOfficialCommand |
| C-annotation | `cmd/tossctl/verify_reconcile.go` | mutating 주석 제거 | CAUGHT | TestMutatingAnnotationOnTradeCommands,TestVerifyReconcileIsRegisteredAsAMutatingOfficialCommand |
| C-symbol-flag | `cmd/tossctl/verify_reconcile.go` | --symbol 플래그 추가 | CAUGHT | TestVerifyReconcileOffersNoIdentifierSymbolOrBypassFlag |
| C-env-creds | `cmd/tossctl/verify_reconcile.go` | 자격을 환경 변수에서도(preflight 가 덮음) | SURVIVED | — |
| C-adapter-embed | `cmd/tossctl/verify_reconcile.go` | 클라이언트 임베딩(쓰기 메서드 승격) | CAUGHT | TestReconcileNarrowReaderNeverYieldsAWriteMethod |
| V-cap-loop-end | `internal/verifylive/reconcile_check.go` | 상한 도달을 수락으로(조건주문 그룹) | CAUGHT | TestReconcileRefusesPaginationFaultsInEitherRead |
| V-cap-loop-end-plain | `internal/verifylive/reconcile_check.go` | 상한 도달을 수락으로(일반 주문 그룹) | CAUGHT | TestReconcileRefusesPaginationFaultsInEitherRead |
| GO2-static-after-approval | `internal/verifylive/reconcile.go` | 정적 거절(Q1·Q3)을 승인 뒤로(A-GREEN 리뷰어 변이 GO2) | CAUGHT | TestReconcileStaticRefusalsComeBeforeTheHumanApproval |
| MJ10-marshal-held-until | `internal/verifylive/reconcile_record.go` | 대사 확정형이 held_until 을 싣게(HeldUntil 중복 대입 삭제 뒤 A-GREEN P2-b 확인) | CAUGHT | TestReconcileLineCarriesNoHoldEvenWhenTheOutstandingLineHadOne |
| CG1-second-rows | `internal/verifylive/reconcile.go` | 둘째 스냅숏 행 검사 제거(CG-1) | CAUGHT | TestReconcileValidatesTheSecondSnapshotToo |
| CG2-no-flock | `internal/verifylive/record.go` | 기록 파일 flock 무시(CG-2) | CAUGHT | TestReconcileRefusesWhileAnotherReconciliationHoldsTheRecordFile |
| CG2-locked-code | `internal/verifylive/reconcile.go` | 잠김 거절 코드 뭉갬(CG-2) | CAUGHT | TestReconcileRefusesWhileAnotherReconciliationHoldsTheRecordFile |
| CG2-path-reread | `internal/verifylive/reconcile.go` | 잠근 fd 대신 경로로 재읽기(CG-2 ABA) | CAUGHT | TestReconcileFinalGateSitsAtTheWriteBoundary |
| CG2-size-check | `internal/verifylive/record.go` | 쓰기 직전 크기 검사 제거(CG-2) | CAUGHT | TestLockedAppendRefusesWhenTheRecordGrewAfterTheLockedRead |
| CG3-two-writes | `internal/verifylive/record.go` | 줄과 개행을 두 번에 씀(CG-3) | CAUGHT | TestReconcileAppendsInOneWriteEndingWithANewline |
| CG3-partial-ignore | `internal/verifylive/record.go` | 부분 쓰기 무시(CG-3) | CAUGHT | TestReconcileReportsAPartialAppendLoudly |
| CG3-sync-ignore | `internal/verifylive/record.go` | fsync 실패 무시(CG-3) | CAUGHT | TestReconcileReportsAFailedSyncLoudly |
| CG3-readback | `internal/verifylive/record.go` | read-back 대조 제거(CG-3) | CAUGHT | TestReconcileDetectsAWriterInterleavingAtTheAppend,TestReconcileVerifiesTheAppendByReadingItBack |
| CG4-display-ignore | `internal/verifylive/reconcile.go` | 승인 표시 실패 무시(CG-4) | CAUGHT | TestReconcileRefusesWhenTheApprovalCannotBeShown |
| CG4-eof | `cmd/tossctl/verify_reconcile.go` | 불완전 줄 승인(CG-4) | CAUGHT | TestReconcileApprovalNeedsACompleteLine |
| CG4-ctx | `cmd/tossctl/verify_reconcile.go` | ctx 취소 무시(CG-4) | CAUGHT | TestReconcileApprovalHonoursCancellation |
| CG4-question-err | `cmd/tossctl/verify_reconcile.go` | 질문 표시 실패 무시(CG-4) | CAUGHT | TestReconcileApprovalRefusesWhenTheQuestionCannotBeShown |
| CG4-channel | `cmd/tossctl/verify_reconcile.go` | 승인 표시를 stdout 으로(CG-4) | CAUGHT | TestReconcileShowsTheAccountsOnTheSameChannelAsTheQuestion |
| CG5-gate-before-lock | `internal/verifylive/reconcile.go` | 순서 교란: 최종 게이트를 잠금 앞으로 추가(CG-5) | CAUGHT | TestReconcileRechecksTheWindowAtBothContractPoints |
| CG5-gate-before-encode | `internal/verifylive/reconcile.go` | 순서 교란: 게이트를 직렬화 앞으로(CG-5) | CAUGHT | TestReconcileFinalGateSitsAtTheWriteBoundary |
| CG6-no-recheck | `cmd/tossctl/verify_reconcile.go` | 신원 재확인 생략(CG-6) | CAUGHT | TestReconcileReaderRechecksIdentityBeforeItsFirstScopedRead,TestReconcileNarrowReaderBindsTheValidatedSequence |
| CG6-plain-unchecked | `cmd/tossctl/verify_reconcile.go` | 일반 주문 읽기 앞 재확인 생략(CG-6) | CAUGHT | TestReconcileReaderRechecksIdentityBeforeItsFirstScopedRead |
| CG6-seq-ignored | `cmd/tossctl/verify_reconcile.go` | seq 재확인 생략(CG-6) | CAUGHT | TestReconcileReaderRechecksIdentityBeforeItsFirstScopedRead |
| CG6-count-ignored | `cmd/tossctl/verify_reconcile.go` | 복수 계좌 재확인 생략(CG-6) | CAUGHT | TestReconcileReaderRechecksIdentityBeforeItsFirstScopedRead |
