# a194-a-holding-left-unprotected-by-a-reconcile-block-is-reported Review

## 계약 작성 (2026-11-23)
- 작성: Manager(Fable). 근거 조사: 읽기 전용 조사원(Opus) — 분기 주장의 AST 는 세션 스크래치 산출물 인용이며 change 안 사본·HEAD 재확인은 tasks 1.3.
- `openspec validate a194-a-holding-left-unprotected-by-a-reconcile-block-is-reported --strict`: valid (rc 0, 2026-11-23 HEAD 1ea261100).
- Function Logic Map: 편집 대상 기존 함수의 번들은 tasks 1.5 에서 만든다.

## 1단계 계약·증거 (팀메이트 Opus, 2026-11-23, 격리 worktree `~/tossos-work/a194-impl`, 브랜치 `fix/a194-unprotected-holding-report`)

- 1.1 `base-commit.txt` = `064f59fd99c11ea3a19adcdb804bbcdfed2470cb`(`capture_change_base.py` rc 0). `generate_master_tracker.py --check` rc 0. Story `STORY-TOS-a194` → feature `FEAT-TOS-005`(docs/pm/portfolio/stories/STORY-TOS-a194.yaml:4).
- 1.2 `analysis/code-context/` 세 파일(CodeGraph 1.6.0, worktree 에서 `codegraph init .`; CodeGraphContext·GBrain 은 `not-applicable` 사유 기록).
- 1.3 `analysis/cited-ast/` 17개(HEAD `064f59fd9`). 대조는 아래 표.
- 1.4 편집 표면은 아래 「1.4」.
- 1.5 `analysis/function-logic/internal-app-engine--reconciledriver.judgeholdings/`(pre-edit). `check_analysis.py --change a194-…` rc 0(required 0 — Go 편집 전).
- 1.6 strict validate valid. proposal-freeze 리뷰는 아래 「proposal-freeze 리뷰」.

### 1.3 주장 대조 (proposal·design 의 코드 주장 → AST)

| 주장(출처) | AST·근거 | 판정 |
|---|---|---|
| 차단 종목 보유는 if@152 → continue@153 으로 빠지고 range@188 의 `alertUnmanaged` 에 안 닿음(proposal F1) | `cited-ast/judgeHoldings.ast.json` B10 if@152 · B16 range@188(continue 는 logic-map 이 세지 않아 소스로 확인) | 참 — 표시만 AST 로 바꿈 |
| 그 침묵은 정본 요구(proposal F2) | openspec/specs/exit-policy/spec.md:137 「전이 상태(RECONCILE 대기·…)만 무알림」 · :159 Scenario | 참(문서 근거) |
| `QUANTITY_MISMATCH` 만 자동 해제, 나머지 `ReleaseOperatorOnly`·`Permanent`(proposal F5) | `cited-ast/blockFromReconcileState.ast.json` B1 if@1411 · B2 if@1414 · B3 if@1417, `cited-ast/ReconcileReasonFor.ast.json` B2 case@159 · B3 case@163 | **정정함** — SNAPSHOT_UNAVAILABLE·SNAPSHOT_STALE 은 operator_only 이지만 Permanent 거짓(생산 쓰기 0). 결론(운영자 전용 차단이 있다)은 불변 |
| `reconcile_states` 를 쓰는 8 자리에 알림 없음(proposal F6) | grep — journal 은 obs 를 import 하지 않음, reconcile 의 알림 호출은 acked_boot.go:98 하나 | 결론 참, 「8 자리」는 재현 못 함 → 숫자 빼고 근거로 바꿈 |
| `reconcile.mismatch_permanent` 생산 0(proposal F7) | grep — event.go:95·:370 정의, obs 시험 2파일만 | 참 |
| critical 미전달 → 진입 잠김(proposal Impact) | alertdelivery.go 머리 주석(a124): 차단 + 운영 모드 승격 | **정정함** — 운영 모드 승격을 더 적음 |
| `TestAdoptionIsSilentUnderReconcile` 그대로 통과(proposal Impact) | reconcileloop_test.go:358-381 — QUANTITY_MISMATCH, 한 주기, `Unmanaged`·`exit.position_unmanaged` 0 단언 | 참(10 분 유예 안) |
| a166 census·a163 핀 밖(proposal Impact) | `analysis/code-context/codegraph-baseline.md` 「구조 고정·census」 | 참 |
| `Tracker.Permanent()` = 계좌 전체(design D1) | `cited-ast/Permanent.ast.json`(return@1317), `cited-ast/Refresh.ast.json`, mismatch.go:1065 `hasPermanentQuantityAccountBlock` | **정정함** — 계좌 전체 **수량** 영구 승격만. 다른 원인 계좌 행은 `EntryAllowed` 의 계좌 범위 `Covers`(mismatch.go:264-267)가 잡음 |
| `EntryAllowed` 사유가 `ReasonReconcilePermanent` 로 운영자 전용을 가름(design D1) | `cited-ast/EntryAllowed.ast.json` B1 range@1327 → B2 if@1328 → return@1329 `&RejectedError{Reason: block.Reason, Detail: block.Detail}` | 사유는 실림. 오늘 생산 원인에서는 판정이 맞음. **결정 필요**: 반환에 Cause·Since 가 없어 D6 본문(원인·시각)을 못 만듦, SNAPSHOT_* 는 Release 와 사유가 갈림 |
| 대사 주기 약 62 초(design D2) | reconcileloop.go:95 `DefaultReconcilePeriod = 60s` + 안정화 2s(snapshot.go:376) | 참 |
| 운영 장부 수량 불일치 133 행 전부 해제(design D2) | 운영 장부 접근 금지(지시) | 미검증 |
| key = 타입·`reconcile`·사실·포지션(design D3) | `cited-ast/reconcileUnmanagedKey.ast.json` return@543 | 참 |
| critical 은 메모리 래치 앞에서 반환(design D3) | `cited-ast/alertUnmanaged.ast.json` B1 if@451 → return@453, 래치 B2 if@455 | 참 |
| `exit.position_adoption_failed` critical(design D3) | internal/obs/event.go:386 `criticalEvents` | 참 |
| 편입 꺼짐∧미지정·제외 판정 재사용(design D4) | `cited-ast/judgeHoldings.ast.json` B12 if@163 `Adoption.Excludes` · B13 if@171 `!Enabled ∧ !Included` — 순수 함수(config/engine.go:135-142), 거부된 블록은 0 으로 만들어져 B13 에 듦(config/engine.go:127-131) | 참 — B10 안에서 같은 술어를 부를 수 있음 |
| `reconcile.mismatch_permanent` critical(design D5) | event.go:370 | 참 |
| 재알림 창 0 으로 하루 한 번(design D5) | 드라이버 옵션 `Alerts ExitAlerter` 는 `Notify` 만(exitloop.go:147-149). `Notify` 는 `claimAndDeliver` 에서 `n.remindAfter()`(기본 1시간, notifier.go:60·:301) | **결정 필요** — 지금 배선으로는 창 0 불가(같은 날 시간마다 재무장) |
| `label` 로 종목 표시(design D7) | reconcileloop.go:566 | 참 |
| `judgeHoldings` 는 `Observe` 성공 주기의 마지막 단계(design §2) | `cited-ast/RunOnce.ast.json` B7 if@463 → return@467, call@470 | 참. 그 밖에 B1 if@422 · B2 if@428 · B5 if@451 도 그 전에 반환 — design §2 에 더 적음 |
| a090 시험이 workingSet 이탈 11 개 고정(design §2 B) | a090_workingset_exits_test.go:22-35 | 참 |

### 1.4 편집 표면

- 고칠 기존 함수: `ReconcileDriver.judgeHoldings`(확정). 결정에 따라 더: `Context.ReconcileDriver`(D5 를 `RecordCritical` 배선으로 하면 — reconcileloop.go:357-388, census 밖), `NewReconcileDriver`(새 맵 필드를 생성자에서 만들면 — 새 파일에서 늦게 만들면 불필요).
- census·구조 고정: a166 0건, a163 0건, a090 은 exit 루프만. 새 파일이 지켜야 할 소스 스캔 둘 — `TestExitEligibilityHasOneSpelling`(eligibility_drift_test.go, `EntryDecisionID`·`AdoptionID` 류 철자 금지), `TestProtectionReadinessIsAbsentFromExitReconcileAndFillLoopSources`(a071_readiness_assembly_test.go:27, reconcileloop.go 대상).
- 알림 수·셈을 단언하는 기존 driver 시험(grep 전수 — reconcileloop_test.go 11곳, adoption_include_test.go 5곳, a085_alert_text_test.go 4곳, a095 8곳 중 차단이 끼는 것):
  - `TestAdoptionIsSilentUnderReconcile`(reconcileloop_test.go:358): `Adopted`·`Unmanaged`·`exit.position_unmanaged` 0 단언. 수량 불일치·한 주기 → 깨지지 않음.
  - `TestActiveForeignCauseIsCountedAndBlocksBeforePriceRead`(:383): `Blocked == 1`·`Adopted == 0`·시세 읽기 0 만 단언(알림 수 단언 없음). IDENTIFIER_CONFLICT 는 운영자 전용이라 a194 뒤 이 주기에 새 보고가 나가지만 단언은 안 깨짐(새 보고가 시세를 읽지 않는 한).
  - **`TestUSIncludedSymbolWaitsUnderAccountWidePermanentReconcile`(adoption_include_test.go:109)**: 계좌 전체 영구 + include 종목 → 133 행 `cycle.Unmanaged != 0` 이면 Fatal. 새 보고가 `cycle.Unmanaged` 를 올리면 **깨진다**.
  - `exit.position_adoption_failed` 사실 값을 열거하는 switch·시험·골든: 없음. 사실 값은 adoption.go:95-103 상수와 `unmanagedFact` 의 생산뿐이고, a095 시험은 특정 값의 key 를 리터럴로 찾을 뿐 전수를 단언하지 않음. obs 시험(a091_stop_sold_nothing_test.go:40-44, obs_test.go:141)은 타입 단위라 새 사실 값과 무관.

## proposal-freeze 리뷰

### 1회차 (2026-11-23) — codex-cli 0.165.0, 읽기 전용(`--sandbox read-only`, `</dev/null`), 적대적 Eng 관점

- 원문: `analysis/review/proposal-freeze-r1-codex.md`. 판정 「계약 freeze 보류 — P0 0, P1 4, P2 4」.
- 처분 규칙(위임 지시): 사실 오류 정정은 반영, 새 가드·새 정책·범위 확장·결정 변경은 **Manager 결정 대기**.

| # | 등급 | 발견 요약 | 처분 |
|---|---|---|---|
| R1-1 | P1 | 알림 꺼짐 → 일반 등급으로 내리는 경로가 없다. 두 타입(`exit.position_adoption_failed`·`reconcile.mismatch_permanent`)은 표에서 항상 critical(event.go:370·:386)이고 `Notify` 는 타입으로만 등급을 정함(notifier.go:149-157) → 꺼진 엔진에도 outbox 행·진입 래치·모드 승격(alertdelivery.go:315-325·:446-471). 꺼짐·topic 잔존·topic 없음 각각에서 새 행·래치·전이 0 시험 요구 | **Manager 결정 대기** — 꺼짐일 때의 타입·기록 경로(D3 는 기존 `exit.position_unmanaged` 갈래로 내려갈 수 있으나 D5 는 대응 일반 타입이 없음). 팀메이트 1.3 도 같은 결론 |
| R1-2 | P1 | D5 하루 한 번은 `Notify` 배선으로 불가 — `Notify` 는 `remindAfter()` 기본 1시간(notifier.go:301·:395-399) → 같은 날 시간마다 재무장. `RecordCritical(ctx,e,0)` 배선을 확정해야 함. 공유 Notifier 의 창 변경은 대안이 아님 | **Manager 결정 대기**(배선: 새 옵션 + `Context.ReconcileDriver` 편집 → FLM 번들 추가, census 밖) |
| R1-3 | P1 | D6 본문의 현재형(「걸려 있지 않다」「해제 전까지 직접 관리」)은 해소 뒤 늦게 배달되면 거짓 — 정본 engine-safety(:1905-1907)·기존 `alertAdoptionFailed` 는 관측 시각 + 과거형. `blocked_stale` 에 운영자 해제가 필수인 듯 안내하면 안 됨. PENDING → 해제·편입 → 지연 배달 시험 요구 | **Manager 결정 대기** — D6 문구·본문에 실을 시각(무보호 관측 시각)을 바꾸는 결정. 사실 지적 자체는 참(adoption.go:481-505 확인) |
| R1-4 | P1 | ADDED 만으로는 정본 exit-policy Scenario 「RECONCILE 중 편입 시도 … 무알림」(:159-161 — 모든 RECONCILE 무결정 보유)을 좁힐 수 없다. engine-safety 「무관리 보유 보고의 등급은 사실이 정한다」(:1885)는 「시도했으나 실패」를 critical 로 정의 — 시도하지 않은 두 사실을 critical 로 더하는 것은 등급 계약 확장 | **Manager 결정 대기** — MODIFIED delta·engine-safety delta 추가는 범위 변경. 팀메이트 확인: 정본 reconciliation :83(「수량 불일치 감지 … 알림이 발송된다」)·engine-safety :199(영구 불일치 critical)도 이미 있음 |
| R1-5 | P2 | 성공한 대사 주기에만 보고 → 안정화 미수렴·로컬 상태·Refresh·Observe 실패가 이어지면 영구 차단의 일일 보고도 무기한 침묵(reconcileloop.go:422-470). 미수렴은 `reconcile.clean` 로그. 포지션 조회 실패(adoption.go:129-134)는 차단 분기 앞에서 빠짐, 드라이버는 ingest 알림을 끔(reconcileloop.go:342-347) | **Manager 결정 대기**(범위). design §2 에 다른 반환 셋과 「미수렴은 오류 로그 줄도 없음」은 사실로 적음 |
| R1-6 | P2 | D1 은 `operator_only` 전체를 못 가름(SNAPSHOT_* 는 operator_only + `ReasonReconcileMismatch`), 「생산 쓰기 0」은 복원 행에도 맞는다는 증명이 아님(Refresh 가 활성 행 그대로 복원). `EntryAllowed` 반환에 Cause·Since 없음 | **Manager 결정 대기** — `Blocks()` 의 Release·Cause·Since 를 쓸지 / 지원 원인을 한정할지. 사실은 proposal F5·evidence-reconciliation 「F5 상세」에 정정 반영 |
| R1-7 | P2 | 시험 전략 공백: 창 0 → 1시간 변이는 행 수 1 그대로(outbox UPDATE, journal/outbox.go:334-346) — 실제 발송 수를 봐야 함; 래치 변이에 기록 실패·회복 RED 없음; KST/UTC 자정 분리 없음; 하네스는 `NotificationsEnabled=false`·`fakeAlerts` — 실제 Notifier·outbox·배달 실행자 필요; 선택 정규식이 a095 전체를 못 잡음(`TestA095NotificationsOffNeverRecordsACritical` 빠짐); §4-1 「critical 1」이 어느 타입인지 불명 | **Manager 결정 대기**(시험 전략 변경). 정규식 누락은 사실(`A194|Adoption|ReconcileLoop|Unmanaged` 는 `TestA095NotificationsOff…` 에 안 맞음) |
| R1-8 | P2 | design §3·delta 의 「이미 편입된(ExitEligible) 보유 — 청산이 계속 나간다 / 이미 보호된다」는 과장 — `ExitEligible` 은 자격이고 exit state 개설을 보증하지 않음(position_projection.go:515-518, adoptOne 보호 미개설 경계) | **반영(사실 정정)** — design §3 문장 고침. delta `specs/exit-policy/spec.md` 의 Requirement 괄호 「이미 보호된다」는 Requirement 문장 수정이라 리뷰 재실행 대상 → **Manager 결정 대기** |

- 2회차: **실시하지 않음.** 8건 중 7건이 결정·범위 변경이라 문서가 거의 바뀌지 않았고, 같은 문서를 다시 보내면 같은 발견만 돌아온다. Manager 결정 뒤 requirement 변경 리뷰로 재실시할 것.
- freeze 판정: 보류(1.7 Manager 몫).

### Manager 결정 반영 (2026-11-23)

- Q1~Q7·R1-7·정본 공백 문단을 proposal·design(결정 표 「Manager 결정 2026-11-23」)·spec delta·tasks 에 반영. engine-safety MODIFIED 새 파일. 두 MODIFIED 는 `analysis/spec-delta/make_modified.py` 가 정본 블록을 떠서 치환 4건(E1·E2·S1·S2)만 적용 — 생성물(그리고 delta 파일의 MODIFIED 절)에서 편집을 되돌리면 정본 블록과 같음(True·True), diff hunk 각 2, 그 밖 차이 0(`analysis/spec-delta/edits.md`).
- R1-1~R1-8 처분: 위 1회차 표의 「Manager 결정 대기」는 이 결정으로 닫힘. R1-8 의 delta 괄호는 「이미 exit 관리 대상이다」로 고침.
- 편집 대상 기존 함수 재계산: `judgeHoldings` 하나(드라이버 구조체 새 필드는 함수 본문이 아님, 맵은 새 함수 안에서 생성, `NewReconcileDriver`·`Context.ReconcileDriver`·`RunOnce`·`alertUnmanaged` 무편집) — 번들 추가 없음, `check_analysis.py` rc 0.

### 2회차 (2026-11-23, 마지막) — codex-cli 0.165.0, 읽기 전용

- 원문: `analysis/review/proposal-freeze-r2-codex.md`. 판정 「freeze 보류 권고 — P0 0, P1 2, P2 3」. 1회차 처분 중 R1-1 타입 분리·R1-6 `Covers ∧ ReleaseOperatorOnly`·`cycle.Unmanaged` 비증가는 코드와 부합한다고 확인. MODIFIED 는 의도한 E1·E2·S1·S2 외 차이 없음을 독립 확인.

| # | 등급 | 발견 요약 | 처분 |
|---|---|---|---|
| R2-1 | P1 | `d.alert` 는 반환값이 없고 오류를 로그로 삼킴(reconcileloop.go:568-576) → 그 경로 뒤 래치는 기록 실패를 성공으로 셈 | **반영(호출 경로 명확화)** — design D5: 새 함수가 `d.opts.Alerts.Notify` 반환값을 직접 보고 `Alerts == nil` 은 성공으로 세지 않음. 기존 함수 편집 범위 불변 |
| R2-2 | P1 | `Notify` critical 은 동기 전송(notifier.go:354, 보고당 네트워크·대기 최대 약 34 초) — 계좌 전체 차단 아래 보유 N 개면 첫 주기에 N+1 건 직렬(10 개면 약 6 분), 다음 대사·편입이 밀림. PENDING 행은 창과 무관하게 매 관측 재전송(outbox.go:379) → 「보유당 1시간마다 1건」은 정산된 행에 한함 | 사실 부분 **반영**(proposal Impact 알림 수 문장). 동기 지연을 받아들일지 / 기록·전송 분리를 허용할지는 **Manager 결정 대기** |
| R2-3 | P2 | 완성 계약의 모순 셋: (a) engine-safety MODIFIED 의 유지 문장 「critical 무관리 보고의 본문은 편입 시도가 실패한 시각의 사건」이 시도 없는 차단 보고까지 덮음 (b) exit-policy·reconciliation ADDED Scenario 가 알림 켜짐 전제 없이 critical (c) 「10분 넘게」 vs 「10분 이상」 | (b) **반영** — 세 Scenario WHEN 에 알림 켜짐 전제. (c) **반영** — S1 을 「10분 이상」으로(생성기 수정 뒤 재생성·재검증). (a) 그 문장은 Manager 지시상 「한 글자도 바꾸지 않는」 기존 문장이라 **Manager 결정 대기**(S1 이 「본문은 차단을 관측한 시각의 사건」을 더했지만 유지 문장과의 범위 충돌은 남음) |
| R2-4 | P2 | 추가 변이를 죽일 입력 부족: `EntryAllowed` 사유 변이는 SNAPSHOT_* 입력이 있어야 갈림; 보유 보고 래치 변이는 보유 보고의 기록 실패→재관측→성공 시험이 따로 필요; 전달 완료 1시간 뒤 재전송 | **반영(시험 계획 보완)** — design §4 에 4a·9a·9b |
| R2-5 | P2 | 「실패 주기 무보고」 공시가 코드보다 넓음 — ingest·converge 오류는 `cycle.Err` 에 남고 주기 계속(reconcileloop.go:438-449) | **반영(사실 정정)** — design §6·proposal 을 조기 반환 넷으로 좁히고 `cycle.Err` 를 침묵 조건으로 쓰지 않는다고 적음 |

- 리뷰 회차 상한 도달. freeze 판정은 Manager(1.7).

### Manager 1단계 검증·freeze 2026-11-23
- Manager 1단계 검증·freeze 2026-11-23(커밋 `ba41f0059` 를 Manager 가 직접 읽고 strict validate 재실행). 남은 두 건: R2-2 — 동기 전송 수용 + 보고를 `judgeHoldings` 끝에서·주기당 보유 보고 `Notify` 최대 3(design D3, 시험 13a·13b, 변이 2 추가); R2-3(a) — MODIFIED E5 허용(생성기 치환, `analysis/spec-delta/edits.md`, 되돌리면 정본과 같음 True·True, 그 밖 차이 0). R2-1·R2-4·R2-5·R2-2 사실 반영 승인.

## 2단계 (팀메이트 Opus, 2026-11-23)

### 2.1 Pre-Edit 선언

| 항목 | 내용 |
|---|---|
| 대상 심볼 | 편집: `ReconcileDriver.judgeHoldings`(adoption.go:107-193) 본문 — B10 안 모으기 1줄 + 끝에서 `reportBlocked` 호출 1줄. `ReconcileDriver` 구조체(reconcileloop.go:218-256)에 메모리 필드 셋(선언만 — 함수 본문 아님). 새 파일 `internal/app/engine/reconcile_block_report.go`(새 함수 전부). 무편집: `internal/obs` 타입 열거 · `NewReconcileDriver` · `Context.ReconcileDriver` · `RunOnce` · `alertUnmanaged` · `alertAdoptionFailed` · `unmanagedFact` · `blocked` |
| 호출부 | `judgeHoldings` ← `RunOnce`(reconcileloop.go:470) 1자리, `RunOnce` ← `Run` 1자리(CodeGraph 1.6.0 + grep, `analysis/code-context/codegraph-baseline.md`) |
| 기존 테스트(무편집 통과) | `TestAdoptionIsSilentUnderReconcile` · `TestActiveForeignCauseIsCountedAndBlocksBeforePriceRead`(reconcileloop_test.go) · `TestUSIncludedSymbolWaitsUnderAccountWidePermanentReconcile`(adoption_include_test.go:109, `cycle.Unmanaged` 0) · `TestA095NotificationsOffNeverRecordsACritical` · `TestExitEligibilityHasOneSpelling` · `TestProtectionReadinessIsAbsentFromExitReconcileAndFillLoopSources` |
| 불변식 | 차단 종목 편입 0(B10 continue 유지) · 새 보고는 `adopt`·기존 무관리 보고 **뒤** · 브로커 호출 추가 0 · `cycle.Unmanaged` 불변 · 알림 꺼진 엔진에 critical 타입 `Notify` 0 · 보유 보고 critical 앞 메모리 래치 없음 · 차단 보고 래치는 `Notify` nil 뒤에만 · 계좌 정보 0 |
| 실패 테스트(RED) | `internal/app/engine/a194_reconcile_block_report_internal_test.go` — design §4 의 1~13b(a095 실 Notifier·실 원장 하네스 재사용). 원문 `analysis/tdd/red.txt` |
| 주기 상한의 구현 해석 | 「한 주기 보유 보고 `Notify` 최대 3, 넘는 보유는 다음 주기」를 **차례(round)** 로 구현: 이번 차례에 시도한 포지션은 차례가 끝날 때까지 다시 시도하지 않고, 차례의 모두를 한 번씩 시도하면 차례를 비운다. N≤3 이면 매 주기 전원 시도(D3 「매 관측」 그대로), 5 개면 3·2·3·2… — 어떤 보유도 ceil(N/3) 주기를 넘게 밀리지 않음. 기록 실패도 다음 차례에 다시 시도되므로 억제 래치가 아님 |
| rollback | 새 파일 삭제 + `judgeHoldings` 2줄·구조체 필드 3개 되돌림. 원장 스키마·설정·토글 변경 없음 |
| 설정 영향 | 없음(새 설정 0). 동작은 기존 `notifications.enabled`·`adoption.*` 를 읽기만 함 |
