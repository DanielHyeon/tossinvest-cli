# a192 Tasks

> **High-risk(reconciliation).** 1단계 = 문서·증거만, production Go·시험 편집 0. Manager 결정 M1~M13 확정(2026-11-17 — design §6). 2단계는 Manager 지시 뒤.
> 실계좌 주문을 내는 자동 시험 금지. 픽스처는 합성 값만(실제 종목·수량·가격·주문번호·계좌 금지). 전체 시험·전체 게이트는 사용자 명시 요청 시에만.
> 장부 스키마 변경 없음(36).

## 1. 1단계 — 계약과 증거

- [x] 1.1 Story `STORY-TOS-a192`(FEAT-TOS-005) 등록, registry·Feature 링크.
- [x] 1.2 change 생성, `base-commit.txt` = `6277ec34e`(`tools/sdd/capture_change_base.py`).
- [x] 1.3 CodeGraph 1.6.0 hard evidence(worktree 색인) — `analysis/code-context/`.
- [x] 1.4 CodeGraphContext·GBrain — not-applicable 사유 기록(`codegraphcontext-context.md`).
- [x] 1.5 `tools/logic-map` AST: 인용 함수 53개(`analysis/cited-ast/`, 전부 rc 0).
- [x] 1.6 proposal·design·spec delta·tasks·review 작성, `openspec validate --strict` 통과.
- [x] 1.7 proposal-freeze 적대 리뷰(다른 계열, 2회) — `review.md`.
- [x] 1.8 Manager 결정 M1~M13(2026-11-17) 반영 개정 + 결정 반영분 리뷰 1회(3회차) — `review.md`.
- [x] 1.9 Manager 의 개정판 검증과 3회차 「Manager 결정 대기」 처분 — M14 = 현재 문서대로, 3회차 반영 4건 승인(2026-11-17). proposal-freeze.

## 2. Pre-Edit

- [x] 2.1 Pre-Edit 선언을 `review.md` 에 확정(2026-11-17).
- [x] 2.2 편집 대상 기존 함수마다 FLM 번들(`analysis/function-logic/<pkg>--<fn>/` — ast·FLM·BTM·risk): `Tracker.Observe`, `Tracker.persist`, `blocksFor`,
      `ValidReconcileReleaseCause`, `buildGateway`, `ReconcileDriver.RunOnce`, `Tracker.Restore`. 원본은 `analysis/cited-ast/` 를 `pre-edit` 로 재사용.
- [x] 2.3 `go test -coverprofile` 로 편집 함수들의 기존 분기 블록 카운트를 잼(BTM 「미확인」 행 채움) — 대상 패키지만.

## 3. RED → GREEN (새 시험 파일 `internal/reconcile/a192_order_terminal_release_test.go`, `internal/journal/a192_order_terminal_release_test.go`)

- [x] 3.1 RED: `TestA192MissingOrderBlockArmsThenReleasesAfterTheTerminalProof`(실 journal — `RecordFill` 로 종결, 비교 1 무장·유지, 비교 2 해제, cause `ORDER_TERMINAL_CONFIRMED`, 게이트 열림).
- [x] 3.2 `TestA192ARestartedTrackerNeverReleasesThisWay`, `TestA192ARowAnotherProducerInsertedFirstIsNeverBound`(필수 조건 1 — `Refresh` 와 `Observe` 사이 끼워 넣기),
      `TestA192AnOverlappingMissingOrderOnAnExistingRowIsNeverBound`, `TestA192AnUnprovenInsertIsNeverBound`.
- [x] 3.3 `TestA192AComparisonNotAfterTheTerminalCommitKeepsTheBlock`, `TestA192ANonCleanTerminalKeepsTheBlock`(비종결·`terminal ∧ fail_closed`·없음·읽기 실패). — **이름 차이**: 같은 이름의 e2e 시험 대신 `TestA192VerdictHoldsForEveryFailedCondition` 의 하위 시험(comparison at/before the commit·undated·not terminal·terminal but fail closed·no snapshot·read failure)과 e2e `TestA192AFailClosedFlipBetweenVerdictAndReleaseKeepsTheBlock` 이 맡음.
- [x] 3.4 `TestA192AQuantityDisputeDisqualifiesTheBinding`, `TestA192EveryBoundOrderMustBeTerminal`, `TestA192AForeignAccountObservationNeverAuthorisesRelease`.
- [x] 3.5 `TestA192AnAmendmentAfterBindingKeepsTheBlock`, `TestA192AnOpenOrderWithTheSameIDKeepsTheBlock`, `TestA192AFillAfterTheBlockKeepsIt`(워터마크 읽기와 삽입 사이 체결 포함).
- [x] 3.6 `TestA192ABindingChangeRearms`, `TestA192AFillCommittedBetweenVerdictAndReleaseKeepsTheBlock`, `TestA192AnAuthorityErrorStillRecordsTheTaint`(외부 포지션만 있는 비교 포함),
      `TestA192ARetriedInsertIsNeverBound`, `TestA192ASameObjectRestoreDropsBindings`.
- [x] 3.7 `TestA192APermanentBlockStaysOperatorOnly`, `TestA192WithoutATerminalReaderTheTrackerBehavesAsToday`, `TestA192TheReleaseLogCarriesNoAccountOrderQuantityOrPrice`.
- [x] 3.8 판정을 켠 tracker 로 기존 안전 흐름 재실행(`TestACoincidentalAgreementStillDoesNotRelease`·a083 크레딧·a110 승격 — 새 시험 파일, 기존 시험 무편집). — 실제: `TestA192TheAdjustmentRulesHoldWithTheReaderOn`(조정 없는 일치·크레딧 해제), `TestA192ACreditWinsOverTheTerminalProof`, a110 승격은 `TestA192APermanentBlockStaysOperatorOnly` 가 판정을 켠 채 돎. 기존 a083/a110 시험 파일 자체를 켠 상태로 다시 돌리지는 않음.
- [x] 3.9 GREEN: 새 파일 `internal/reconcile/order_terminal_release.go`, `internal/journal/reconcile_order_terminal_release.go` + design D11 의 기존 함수 편집.
      편집 뒤 AST·FLM·BTM 재생성(옛/새 ast 정렬로 재번호).

## 4. 검증

- [x] 4.1 변이 원장(`analysis/mutation/`) — design §5 의 m1~m21 + 배관·로그 변이 m22~m33 전부 CAUGHT(최종 33/33 — `analysis/mutation/README.md`), 변이마다 양성 대조군(닿음) 먼저, 사본 하네스는 무변이 대조군 GREEN·시작 sha 단언 후.
- [x] 4.2 선택 시험: `./internal/reconcile/...`, `internal/journal` 의 reconcile_states·risk_bucket_owner 시험, `internal/app/engine` 의 `reconcileloop_test`·`engine_test`,
      `cmd/tossctl` 의 `engine_reconcile_test`. `-race` 는 reconcile 패키지만. 미실행 범위를 기록.
- [x] 4.3 `make lint`(진입점), `openspec validate a192-… --strict`.
- [x] 4.4 gstack review(코드) — 적대적 Eng 보이스 포함, 작성자와 다른 계열, 최대 2회차. 새 가드·범위 확대 제안은 Manager 결정 전 채택 금지.

## 5. 종결

- [x] 5.1 Manager 독립 검증(diff·선택 시험 재실행·변이 표본) — 2026-11-17: 기존 함수 편집 7곳 = D11, production 3파일 sha256 일치, reconcile 209/0·journal 선택 160/0·engine 843/6/1(실패 6 은 `-trimpath` 계측 결함, 옵션 없이 6/6 PASS)·cmd 선택 rc 0, 변이 표본 10/12 → 시험 추가 뒤 12/12, M15·M16·유지 사유 로그 수용; 전체 시험 미실행(권한 규칙) — review.md 「Manager 독립 검증 5.1 기록」.
- [x] 5.2 (사람) 배포 — 이미지 교체(스키마 변경 없음). 배포 순간 활성 missing-order 행은 운영자 해제(결속 없음). 운영 장부 수리 없음. — 2026-11-17 Manager 실행·사용자 승인: 핀 `tossos:a192-…-f269c68f8`(release)·rollback a191, SchemaVersion 36 = 운영 user_version 36, 엔진 부재 약 5초, 교체 뒤 `reconcile.clean` `blocked` 0·미해제 행 0, 배포 순간 활성 missing-order 행 0(운영자 해제 불필요). 새 해제 경로의 운영 동작은 **미관측** — review.md 「배포 기록」.
- [x] 5.3 archive·Story 경로·PM `--check`, 검증된 교훈 retain(ai-memory·Hindsight) — gate ② 선행 조건이라 선체크(선례 a187·a189·a190·a191); 실제 수행 결과는 archive 커밋과 종결 보고.
