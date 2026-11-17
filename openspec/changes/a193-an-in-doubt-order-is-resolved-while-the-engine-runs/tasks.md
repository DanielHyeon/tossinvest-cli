# a193 Tasks

> **High-risk(주문 해소·알림).** 1단계 = 문서·증거만, production Go·시험 편집 0. 2단계는 Manager 지시 뒤.
> 실계좌 주문을 내는 자동 시험 금지. 픽스처는 합성 값만(실제 종목·수량·가격·주문번호·계좌 금지). 브로커는 httptest/가짜만. 전체 시험·전체 게이트는 사용자 명시 요청 시에만.
> 원장 스키마 변경 없음. 설정·토글 없음.

## 1. 1단계 — 계약과 증거

- [x] 1.1 Story `STORY-TOS-a193`(FEAT-TOS-009 — Manager 확인 Q-D) 등록, registry·Feature 링크.
- [x] 1.2 change 생성, `base-commit.txt` = `791fb6ef1`(`tools/sdd/capture_change_base.py`).
- [x] 1.3 CodeGraph 1.6.0 hard evidence(worktree 색인) — `analysis/code-context/`.
- [x] 1.4 CodeGraphContext·GBrain — not-applicable 사유 기록(`codegraphcontext-context.md`).
- [x] 1.5 `tools/logic-map` AST: 인용 함수 40개(`analysis/cited-ast/`, 전부 rc 0) + 편집 대상 2개 번들(`analysis/function-logic/`, pre-edit).
- [x] 1.6 proposal·design·spec delta·tasks·review 작성, `openspec validate --strict` 통과.
- [x] 1.7 proposal-freeze 적대 리뷰(다른 계열, 최대 2회) — `review.md`(1회차 BLOCK P1 5·P2 3, 2회차 BLOCK P0 1·P1 2·P2 3 — 사실 정정 반영, 판단은 결정 대기).
- [x] 1.8 Manager 결정 Q-A~Q-I 처분(2026-11-18) 반영 — design D0, proposal, delta(처음 봄·세션 park 통지·원장 못 읽음), review.md. 리뷰는 2회차로 끝(Manager).
- [x] 1.9 Manager 1단계 검증 통과(HEAD `0a48e0138`) — Q-G 는 실측으로 닫음, proposal-freeze.
- [x] 1.10 (Manager, 읽기 전용 — 2026-11-18 완료: IN_DOUBT 6회 전부 park, 지속 11.9~46.3시간, 동시 최대 3; 일회 측정·스크립트 미커밋 — design D0.1) 운영 저널 사본에서 IN_DOUBT PLACE attempt 의 기간별 수와 **동시 최대 수**를 잰다(계좌·종목·수량·가격 미기록) — 정본이 요구한
      「주기와 동시 실행 상한의 실측」의 운영 쪽. 결과를 design D2 표 「실측 2」에 적는다.

## 2. Pre-Edit

- [x] 2.1 Pre-Edit 선언을 `review.md` 에 확정.
- [x] 2.2 편집 대상 기존 함수 번들 확정 — `cmd/tossctl--engineruntime`, `cmd/tossctl--testproductionruntimestartsexactlyonealertdeliverer`(1단계 pre-edit 판 있음).
      편집 직전 `rg 'AuxiliaryNames\(\)|in-doubt-settle'` 와 a166 `shapes.json` 키를 다시 세어 편집 표면이 늘지 않았는지 확인.
- [x] 2.3 `go test -coverprofile`(대상 패키지만: `./cmd/tossctl` 의 engine_runtime_branch·a098 시험)로 `engineRuntime` 의 기존 분기 블록 카운트를 잼(BTM 「미측정」 행 채움).

## 3. RED → GREEN

새 시험 파일 `internal/app/engine/a193_in_doubt_settle_test.go`(실 원장 `journal.Open` 임시 경로 + 가짜 목록·계좌 판독기 + `clock.NewFake`), `cmd/tossctl/a193_in_doubt_settle_wiring_test.go`.
기존 시험 파일 무편집(단, 3.10 의 기대 목록 한 줄은 예외 — FLM 번들 있음).

- [x] 3.1 RED: `TestA193AnInDoubtPlaceFoundInTheListIsConfirmedWhileTheEngineRuns` — 세션 중 5xx 로 IN_DOUBT 된 매도 PLACE, 목록(OPEN)에 같은 지문 1건 → 한 스윕 뒤
      원장 CONFIRMED + broker_order_id, 발의 무장 유지(`ExitIntentAttempts` ≠ Unaccepted 해제), 진입 게이트 래치 0, critical 0.
- [x] 3.2 `TestA193AnInDoubtSellWithoutBaselineParksAndTheParkCauseAlertsOnce` — 목록 부재, 기준선 없음 → 안정 3회·45초(가짜 시계) 뒤 park(사유에 「without a
      pre-dispatch account baseline」), `ReasonUnresolvedInDoubt` 래치, 그 뒤 exit 관측 판정 2주기에 `recordParkCause` critical **1회**(key `order.unresolved_in_doubt|<position>|attempt:<id>`),
      실행자 자신의 critical 0(중복 없음).
- [x] 3.3 `TestA193AnInDoubtBuyProvenAbsentIsFailedConfirmed` — 기준선 있는 PLACE(전략 진입 모양), 부재 + 보유·매수가능 무변 → FAILED_CONFIRMED(증명 규칙 무변경의 회귀 확인).
- [x] 3.4 `TestA193ACancelledResolutionLeavesTheAttemptInDoubt` — 가짜 시계의 Sleep 대기 중 ctx 취소 → 원장 상태 IN_DOUBT·settled_at 빈 값, 진입 게이트 래치 0, critical 0,
      결과 로그 0, `Run` 반환 `context.Canceled`. 하위 시험: 스캔(HTTP) 도중 취소 · 부재 교차 확인 읽기 도중 취소 — `Resolve` 에서 park 를 부르는 네 return 자리(299·325·348·351) 중 취소가 닿는 셋을 덮음.
      하위 시험(리뷰 R1-6): 10분 넘은 attempt 의 스윕 중 취소 → `noteStale` 미호출·`ReasonAlertUndelivered` 래치 0.
- [x] 3.5 `TestA193AShutdownMidResolutionIsAGracefulAuxiliaryStop` — `Runtime.Run` 에 실행자를 보조로 넣고 해소 중 취소 → `OnStop` 미호출, `Run` nil, 원장 IN_DOUBT.
- [x] 3.6 `TestA193TheExitObserverDoesNotWaitForAResolution` — 막혀 있는 가짜 해소기(채널로 붙잡음) 동안 다른 포지션의 exit 관측 주기가 끝나고 손절 제출이 나감;
      `TestA193AttemptsAreResolvedOneAtATime` — IN_DOUBT 둘(다른 종목), 동시에 진행 중인 `Resolve` 최대 1(계측);
      `TestA193TheSameAttemptIsNotResolvedTwice` — 종결 뒤 다음 스윕에서 그 attempt 로 `Resolve` 호출 0.
- [x] 3.7 `TestA193AnAttemptInDoubtPastTheBoundAlertsOnce` — 해소기가 매번 오류(브로커 이전 오류) → 10분 전 critical 0, 10분에 critical 1(타입 `order.in_doubt`,
      key `order.in_doubt|stale:<attempt id>`, `remindAfter` 0), 그 뒤 스윕 반복에도 1, 새 실행자(재시작 흉내)도 같은 key; `DispatchStartedAt` 빈 attempt 는 `RecordedAt` 기준;
      기록 실패면 래치 안 서고 다음 스윕이 다시 시도. `TestA193TheStaleAlertCarriesNoAccountQuantityOrPrice`(제목·본문·필드에 합성 계좌·수량·가격 문자열 0).
- [x] 3.8 `TestA193OnlyInDoubtPlaceAttemptsAreTouched` — RECORDED·DISPATCH_STARTED·ACKED PLACE, IN_DOUBT CANCEL·AMEND → 상태 무변·`Resolve` 호출 0·브로커 호출 0;
      `TestA193TheSettlerNeverReplays` — 재생 진입점 호출 0(가짜 Replayer 계측); `TestA193NoInDoubtMeansNoBrokerCall` — 미결 0 이면 스윕 N회에 브로커 요청 0.
- [x] 3.9 `TestA193ResolutionRequestsAreBounded` — 가짜 브로커로 요청 수를 셈: 발견 경로, 부재 증명 경로(≈22), 기한 소진 경로(정상 1쪽 ≤122), 페이지 상한.
      여러 쪽·느린 응답 경로의 **벽시계 길이**도 잼(리뷰 R1-2 — 강제 상한이 없음을 수치로). 결과를 design D2 「실측 1」에 적음.
- [x] 3.10 배선: `TestA193TheProductionRuntimeStartsTheSettlerAsAuxiliary`(보조 이름에 있고 `LoopNames` 에 없음) + 기존 `TestProductionRuntimeStartsExactlyOneAlertDeliverer`
      기대 목록에 `"in-doubt-settle"`(FLM 번들). `TestA193ASettlerStopRecordsACritical` — Run 의 패닉·nil 배선 → `OnStop` critical 1(key `order.in_doubt|settler-stopped:<실행 id>`).
      `TestA193TheSettlerStartsAfterRecovery` — Recover 가 막혀 있는 동안 스윕 0.
- [x] 3.12 알림 수(Q-A·Q-E 처분): `TestA193FirstSightIsAlertedOnce`(스윕 3회에 key `order.in_doubt|attempt:<id>` 1, 기록 실패면 다음 스윕 재시도),
      `TestA193AParkOutsideAnExitProposalIsAlertedOnce`(무장 발의 없는 PLACE park → `order.unresolved_in_doubt|attempt:<id>` 1),
      `TestA193AParkOfAnArmedExitProposalIsNotAlertedTwice`(무장 발의의 attempt park → 실행자 park 알림 0, `ExitIntentAttempts.Parked` 에 그 attempt — 기존 `recordParkCause` 의 입력),
      `TestA193AnArmedProposalReadFailureStillAlerts`, `TestA193AnEpisodeRaisesFirstSightAndParkOnly`(정상 경로 알림 = 처음 봄 1 + park 1, 10분 0).
- [x] 3.13 `TestA193AnUnreadableLedgerAlertsAfterTenMinutes`(연속 실패 9분59초 0 · 10분 1 · 계속 1 · 성공 뒤 새 연속은 새 key) ·
      `TestA193StaleAndFirstSightAreCheckedBeforeResolving`(막힌 해소 앞에서도 1단 알림이 먼저 기록됨).
- [x] 3.11 GREEN: 새 파일 `internal/app/engine/in_doubt_settle.go` + `engineRuntime` 원소 1. 편집 뒤 두 번들의 AST·FLM·BTM 재생성(옛/새 ast 정렬로 재번호).
      기존 `TestEngineRuntimeConstructionBranchesFailClosedAndAssembleExactSuccess` 가 영값 원장으로 보조 실행자를 띄워도 패닉·정지 기록이 없음을 확인(design D2 주의).

## 4. 검증

- [x] 4.1 변이 원장(`analysis/mutation/`) — design §5 의 m1~m15 전부 CAUGHT, 변이마다 양성 대조군(닿음) 먼저, 사본 하네스는 무변이 대조군 GREEN·시작 sha 단언 후.
- [x] 4.2 선택 시험(codegraph-baseline.md 「선택 시험」): `./internal/app/engine -run 'A193|Auxiliary|Runtime'`, `./cmd/tossctl -run 'A193|A098|EngineRuntime|ProductionRuntime'`,
      `./internal/execgw -run 'Resolve|InDoubt'`, `./internal/journal -run 'Resolve|Pending'`. `-race` 는 `./internal/app/engine -run A193` 만. 미실행 범위 기록.
- [x] 4.3 `make lint`(진입점), `openspec validate a193-… --strict`, `python3 tools/logic-map/check_analysis.py --change a193-…`.
- [ ] 4.4 gstack review(코드) — 적대적 Eng 보이스 포함, 작성자와 다른 계열, 최대 2회차. 새 가드·범위 확대 제안은 Manager 결정 전 채택 금지.
- [x] 4.5 `docs/operations.md` 에 「세션 중 IN_DOUBT 해소와 park」 절: 로그 줄 이름, 10분 알림의 뜻, park 뒤 계정 전역 진입 래치가 프로세스 수명이라는 것, 해동 절차.
- [x] 4.6 (처리 변경) retry matrix 문서(`openspec/changes/verify-execution-capability/analysis/retry-matrix-slo-2026-10-22.md`)에 세션 해소 조회 빈도 계상 한 줄(정본 §0.4).

## 5. 종결

- [ ] 5.1 Manager 독립 검증(diff·선택 시험 재실행·변이 표본).
- [ ] 5.2 (사람) 배포 — 이미지 교체(스키마 변경 없음). 배포 순간 IN_DOUBT attempt 가 있으면 기동 복구가 먼저 해소(종전과 같음).
- [ ] 5.3 archive·Story 경로·PM `--check`, 검증된 교훈 retain(ai-memory·Hindsight).

## 2단계 기록 (2026-11-18)

- 3.x: RED 원문 `analysis/implementation/red-engine.txt`(26 시험 중 21 FAIL · 5 PASS — 아래)·`red-cmd.txt`(2 FAIL). GREEN `green-engine-a193.txt`(29 PASS — VERIFY 에서 3개 더함)·
  `green-cmd.txt`(9 PASS). 골격(아무것도 안 하는 Run·sweep)에 대해 PASS 한 5개는 「하지 않음」을 재는 시험이라 골격에서 RED 가 날 수 없다:
  `TestA193OnlyInDoubtPlaceAttemptsAreTouched`·`TestA193NoInDoubtMeansNoBrokerCall`·`TestA193TheSettlerNeverReplays`·`TestA193AParkOfAnArmedExitProposalIsNotAlertedTwice`·
  (보강 전) `TestA193AShutdownMidResolutionIsAGracefulAuxiliaryStop` — 마지막 것은 대기자 수를 3 으로·해소 시작을 단언하도록 보강해 RED 를 관측했다. 나머지 넷의 이빨은 변이
  M01·M02·M21·M17 로 증명(`analysis/mutation/`).
- 시험 이름 차이: 3.6 의 「exit 관측 루프」 시험은 실제 exit 관측자가 아니라 런타임 수준(해소가 막혀 있는 동안 다른 감독 루프의 주기가 계속 도는가)으로 쟀다.
  3.2 의 park 원인 알림 1회는 실행자가 park 를 알리지 않음 + `ExitIntentAttempts.Parked` 에 그 attempt 가 있음(기존 `recordParkCause` 의 입력)까지만 잰다 — exit 관측자를
  돌려 key 를 세는 시험은 a094 의 기존 시험이 맡는다.
- 4.6: 다른 change(`verify-execution-capability`)의 분석 파일은 고치지 않고, 계상은 design D0.5 에 적었다(새 엔드포인트 0, 해소당 요청 수 실측).

