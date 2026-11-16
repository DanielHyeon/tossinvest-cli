# a191 Tasks

> **High-risk(체결 감지).** 1단계 = 문서·증거만, production Go 편집 0. 2단계는 Manager 의 1단계 검증과 M1~M10 결정(특히 **M8 필수**) 뒤에만 시작한다.
> 실계좌 주문을 내는 자동 시험 금지. 픽스처는 합성 값만(실제 종목·수량·가격·주문번호 금지).

## 1. 1단계 — 계약과 증거

- [x] 1.1 Story `STORY-TOS-a191`(FEAT-TOS-002) 등록, registry·Feature 링크.
- [x] 1.2 change 생성, `base-commit.txt` = `a392308e8`(`tools/sdd/capture_change_base.py`).
- [x] 1.3 CodeGraph 1.6.0 hard evidence(worktree 색인) — `analysis/code-context/`.
- [x] 1.4 CodeGraphContext·GBrain — not-applicable 사유 기록(`codegraphcontext-context.md`).
- [x] 1.5 `tools/logic-map` AST: 편집 대상 `Detector.collect` 번들 + 인용 함수 19개(`analysis/cited-ast/`).
- [x] 1.6 proposal·design·spec delta·tasks·review 작성, `openspec validate --strict` 통과.
- [x] 1.7 proposal 단계 적대 리뷰(작성자와 다른 계열 보이스, 최대 2회차) — `review.md`.
- [x] 1.8 Manager 1단계 검증(2026-11-16 통과) + 결정 M1(경보 통로)·M2(UNKNOWN 단건 포함)·M3(간격·보고 시한 값)·M4(수량 불일치 처분)·M5(Feature)·M6(execution 존재 요구)·M7(빈 커서 거절)·**M8(지연·총량 상한 — 착수 전 필수)**·M9(freshness)·M10(잘못된 종결 잔여 위험). 결정을 design 에 반영하고 바뀐 Requirement 는 재리뷰.
- [x] 1.9 (선택, 사람 승인) design Q9 의 읽기 전용 조회 1회 — **콘솔 대안으로 수행 — 정확한 조회 형태는 미수행**(Manager 실측 2026-11-17, 사용자 승인 2026-11-17). 결과로 Q3 의 미실측 행을 갱신(지지됨, 한계 ①~⑤ 는 design Q9 결과).

## 2. Pre-Edit

- [x] 2.0 결정 반영분(Requirement 변경)에 한정한 리뷰 1회(다른 계열) → `openspec validate --strict`.
- [x] 2.1 Pre-Edit 선언을 `review.md` 에 확정(1단계 초안 있음 — M1 결과로 대상 심볼 목록 갱신).
- [x] 2.2 `go test -coverprofile` 로 `collect` B2·B3·B5·B15·B16·B18 의 기존 블록 카운트를 잼(branch-test-map 「미확인」 행 채움).

## 3. RED → GREEN (`internal/filldetect/a191_closed_lookup_test.go` 새 파일, 상태 질의를 가르는 새 가짜 pager)

- [x] 3.1 RED: `TestA191OrderThatLeftTheOpenListAndIsCanceledInTheClosedListIsRecordedTerminal`(실 journal — 종결 CANCELLED, `TrackedFillOrders` 0).
- [x] 3.2 RED: `TestA191TerminalSnapshotLeavesTheReconcileLocalState`(같은 흐름 뒤 `reconcile.LocalStateFromJournal` 의 `OpenOrders` 0 — import 순환이면 `internal/reconcile` 쪽 시험 파일).
- [x] 3.3 `TestA191SingleReadPendingAndClosedListSilentKeepsTodaysSnapshot`(못 찾음 → 오늘과 같음, `Cycle.Unresolved` 사유, FailClosed 0, 게이트 래치 0).
- [x] 3.4 `TestA191ClosedListFillDisagreementNeverRecordsAFill`(CLOSED FILLED vs 단건 PENDING 0 → fill_events 0) + 주문 수량만 다른 경우.
- [x] 3.4b `TestA191AcceptedClosedRowKeepsTheSingleReadExecution`(원장 0, 단건 `PARTIAL_FILLED`/4, CLOSED `CANCELED`/4 → delta 4·단건 가격·종결).
- [x] 3.5 `TestA191ClosedListReadFailureDoesNotFailTheCycle`(CLOSED 429 → 주기 성공, freshness 도장, 같은 주기 재조회 0).
- [x] 3.6 `TestA191ClosedLookupIsThrottledPerOrder`(간격 안 0회, 간격 뒤 1회, 첫 조회 즉시).
- [x] 3.7 `TestA191ClosedRowFromAnotherScopeIsNotAccepted`, `TestA191DuplicateClosedRowsForOneIDAreAmbiguous`, `TestA191ClosedRowIDMustMatchByteForByte`.
- [x] 3.7b `TestA191SameScopeRowsWithDifferentPayloadsAreAmbiguous`(같은 키·다른 status, 페이지 순서 뒤집기), 바이트 동일 중복은 하나로 셈.
- [x] 3.7c 모든 새 시험은 가짜 pager 가 받은 `OrderQuery` 순서를 단언한다(리뷰 1 #11 — 기존 `fakePager` 는 status 를 안 봐 우발 CLOSED 호출을 못 잡음).
- [x] 3.8 `TestA191ReplacedWithoutLineageIsNotAccepted`, `TestA191PartialFilledInTheClosedGroupIsNotTerminal`, `TestA191UnknownSingleReadIsNeverLookedUp`(M2 = 제외 — CLOSED 호출 0, 기존 refusal 그대로).
- [x] 3.9 `TestA191OrderStillInTheOpenListIsNeverLookedUp`, `TestA191TerminalSingleReadIsNeverLookedUp`, `TestA191ExistingTrackedReadGuardsRunBeforeTheLookup`.
- [x] 3.10 `TestA191ClosedQueryNamesTheSymbolAndTheTradingDayWindow`(KR·US, `Status=CLOSED`·`Limit=100`·from/to ±1일), `TestA191ClosedWalkStopsAtThePageCeiling`.
- [x] 3.11 `TestA191UnresolvedIsReportedOnlyAfterItPersists`, `TestA191JustPlacedOrderMissingFromTheOpenListIsNotEscalated`.
- [x] 3.12a `TestA191ClosedRowWithoutExecutionIsNotAccepted`(M6), `TestA191OnlyOneClosedLookupPerCycle`(M8 ②), `TestA191SlowClosedLookupIsCutAtTheDeadline`(M8 ①), `TestA191UnresolvedAlertCarriesOnlyTheInternalIdentifier`(M1), `TestJournalTrackedCarriesTheIntentID`(새 시험).
- [x] 3.13 GREEN: `internal/filldetect/closed_lookup.go` + `collect` 2줄 + `JournalTracked.TrackedOrders` 1줄 + `engineFillDetector` 1줄 + `cmd/tossctl/a191_fill_alerts.go`. 편집 뒤 AST·FLM·BTM 재생성(옛/새 ast 정렬로 재번호).

## 4. 검증

- [x] 4.1 변이 원장(`analysis/mutation/`) — design §5 의 변이 m1~m26 전부 CAUGHT, 사본 하네스는 무변이 대조군 GREEN 확인 후.
- [x] 4.2 선택 시험: `go test ./internal/filldetect/...`(기존 시험 행동 변화 0 확인), journal `TrackedFillOrders`·`RecordFill` 종결, reconcile `LocalStateFromJournal`, cmd/tossctl `engineFillDetector` 호출 시험(M1-A). `-race` 는 filldetect 패키지만.
- [x] 4.3 `make lint`(진입점), `openspec validate a191-… --strict`.
- [x] 4.4 retry matrix 정본 표(`openspec/changes/archive/2026-07-26-harden-execution-base/retry-matrix.md`)·`retry-matrix-slo-2026-10-22.md` 에 CLOSED 확인 조회 행 추가 — 코드와 같은 커밋.
- [x] 4.5 gstack review(코드) — 적대적 Eng 보이스 포함, 작성자와 다른 계열, 최대 2회차.
- [x] 4.6 M8 검증(측정: 멈춘 CLOSED 에서 주기 2.00s·다른 주문 같은 주기 적용 — `TestA191SlowClosedLookupIsCutAtTheDeadline`; 빈 상태(=재시작)에서 후보 3 이 3 주기에 하나씩 — `TestA191OnlyOneClosedLookupPerCycle`. **정상 경로에서 2s** 로 한정 — 토큰 잠금·예산 정리 지연은 M11 = 받아들임): 멈춘 CLOSED 응답에서 같은 주기 다른 주문(보호 매도)의 반영 지연 ≤ 2s + 스케줄 오차, 재시작 직후 후보 N 개가 N 주기에 걸쳐 하나씩 물림을 단언.

## 5. 종결

- [ ] 5.1 Manager 독립 검증.
- [ ] 5.2 a167 에 「종결 확인은 a191 스냅숏을 쓴다」 메모, a100 에 「detect.go 줄 번호 재인용」 메모(그 change 파일은 그 change 의 저자가 고침).
- [ ] 5.3 (사람) 배포·운영 장부 복구 순서 — 엔진 정지 창, `reconcile-resolve` 실행은 사람.
- [ ] 5.4 archive·Story 경로·PM `--check`, 검증된 교훈 retain(ai-memory·Hindsight).
