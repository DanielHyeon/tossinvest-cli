# Review

## 1단계 기록 (구현 팀메이트 Opus, 2026-11-17)

### 기억 회고

- ai-memory(workspace `default`, project `junghoonghae-tossinvest-cli`): 「missing-order reconcile block release」「adjusted reconcile credit a083」
  「a110 missing-order promotion permanent」 3회 조회. 걸린 것: `gotchas/toss-single-order-read-stays-pending-after-app-cancel.md`(a191 의 사고 요약 — 새 사실
  없음), 나머지는 세션 기록 잡음. 이 change 의 설계 결정·함정에 대한 기존 기록 없음.
- Hindsight(bank `tossos-dev-episodes`): 같은 세 문구로 1회 조회. 관련 사실은 a191 종결 기록의 「missing-order 심볼 차단은 수렴 조정 크레딧이 없어 저절로
  안 풀린다」 하나 — HEAD 로 재확인함(design Q2·Q3).
- 파일 기억 중 적용한 것(현재 HEAD 로 재확인): `universal-check-passes-on-an-empty-sample`(D3 C2 와 변이 m13 — 결속 0 개 행), `two-judgements-cover-for-each-other`
  (판정은 Go 함수 하나, DB 트리거 없음), `a-new-guard-unpins-the-guards-behind-it`(§5 가드 위치), `fail-closed-must-name-what-it-rejects`(design §4 N1~N18),
  `missing-tool-reports-clean`(`codegraph affected` 기본값 불신 — 1 대 1317), `flm-before-claiming-not-before-editing`(인용 함수 53개 AST 를 문서보다 먼저),
  `tossos-branch-behind-main-schema`(D10 — 스키마 37 의 롤백 비용), `tossos-engine-stop-removes-stoploss`(Q8 P2 근거), `existence-check-is-not-a-role-check`
  (C2 — 결속이 「있다」가 아니라 「행을 만든 트랜잭션에서 쓰였다」).

### 번호 배정

- 확인 범위: `git branch -a --list '*a192*'`(이 브랜치만), `git worktree list`(a192 는 이 worktree 하나), `git log --all --oneline -5000` 제목의 `a192` 0건
  (해시 `fa1922586` 의 부분 문자열 1건뿐), `~/tossos-work`(이 worktree 만), `openspec/changes`·`archive`(a192 0), `docs/pm/portfolio`(Story a191 까지).
- a192 를 씀(Manager 배정과 같음). 이름: `a192-a-missing-order-block-is-released-by-the-orders-terminal-proof`(Manager 안 그대로 — 더 나은 이름 제안 없음).

### Story·Feature

- `STORY-TOS-a192` → `FEAT-TOS-005`(Position adoption and common exit policy). 근거: 같은 대사 차단 해제 계열 a083(크레딧)·a110(같은 분쟁만 영구)이
  이 Feature 에 있다. 대안 FEAT-TOS-002(a191 이 있는 실행 검증)는 원인 쪽이고 이 change 는 차단 상태기계라 고르지 않음. **Manager 확인 대상(M10).**

### 증거

- CodeGraph 1.6.0: `analysis/code-context/`(worktree 색인, 원문 `codegraph-callers.txt`·`codegraph-affected.txt`).
- CodeGraphContext·GBrain: not-applicable(사유 `codegraphcontext-context.md`).
- AST: 인용 53개 `analysis/cited-ast/`(전부 rc 0). 손으로 읽은 것은 `analysis/cited-ast/README.md` 표와 design 본문에 표시.
- 운영 장부·브로커: 읽지 않음(지시상 금지). 운영 사실은 a191 기록·Manager 서술 인용.

### Pre-Edit 선언 (확정 2026-11-17 — M14 결정·3회차 반영 승인 뒤, proposal-freeze)

```text
Pre-Edit Gate:
- change id / task id: a192-a-missing-order-block-is-released-by-the-orders-terminal-proof / 3.9
- 대상 심볼(패키지.함수): reconcile.Tracker.Observe(1차 판정·무장 호출, 결속 갱신 호출 2), reconcile.Tracker.persist(워터마크 읽기·삽입 증명 붙잡기·
  종결 증명 해제 갈래), reconcile.blocksFor(신원 비공개 필드), journal.ValidReconcileReleaseCause(case 1), engine.buildGateway(Terminal 배선),
  engine.ReconcileDriver.RunOnce(로그 호출), reconcile.Tracker.Restore(결속·무장 지우기). journal.Journal.EnterReconcile 무편집. 새 파일 internal/reconcile/order_terminal_release.go,
  internal/journal/reconcile_order_terminal_release.go.
- CodeGraph definition/callers/callees/impact: analysis/code-context/codegraph-baseline.md(Observe·persist·blocksFor 호출자 각 1, EnterReconcile 3,
  ValidReconcileReleaseCause 2; affected -d 1 = 24 파일)
- CodeGraphContext 후보와 evidence reconciliation: not-applicable(사유 기록), evidence-reconciliation.md
- 기존 동작 파악 근거: design §1(Q1~Q6), cited-ast 53, 기존 시험 TestMissingOrderRaisesASymbolBlock·TestACoincidentalAgreementStillDoesNotRelease·a083/a110 시험
- Function Logic Map / Branch Test Map: 2단계 2.2(편집 대상 6~7개, pre-edit 원본 = cited-ast)
- upstream 상속 테스트 영향: no(internal/reconcile·reconcile_states 는 TossOS 계층)
- 실패 테스트 선행 작성: yes(tasks 3.1~3.3)
- 설정·DB·journal 변경과 rollback: 스키마 변경 없음(36). 새 release cause 값 하나(쓰기 검증 상수). 롤백 = 이전 이미지. 설정·토글 없음.
- 안전 불변식 §0 위반 여부 검토: 통과 — 청산·손절 경로 무변경(§0.3), 브로커 호출 0(§0.4), 스키마 무변경(§0.6), 토글 없음(§0.7),
  해제는 증거 요건이 있는 방향으로만 넓힘(§0.9 의 손절·사이징 대상 아님).
```

### 리뷰 보이스 계획

- 작성자: Claude(Opus). 바깥 보이스: Codex(다른 계열), `codex exec -s read-only`, 적대적 Eng 관점 필수(High-risk). 회차 상한 2.
- 리뷰가 제안한 새 가드·새 정책·범위 확대는 설계에 넣지 않고 아래 「Manager 결정 대기」에 적는다. 사실 오류(인용·기전)만 문서에 반영한다.

### 리뷰 1회차 (2026-11-17, Codex gpt-6-astra, codex-cli 0.165.0, `codex exec -s read-only`, 적대적 Eng)

원문 `analysis/review/review-1-codex.md`. 판정: 「NOT READY」. 9건(BLOCKER 3·MAJOR 4·MINOR 1·NIT 1). 내가 코드로 확인한 뒤의 처분:

| # | 등급 | 요지 | 확인 | 처분 |
|---|---|---|---|---|
| 1 | BLOCKER | 행이 생긴 뒤의 투영 거절(과매도·귀속 실패)이 기존 행에 안 남아 C2/C3 를 넘는다 | 맞음 — `enterReconcileScopeInTx` B3(if@401) → return@405, 거절을 따로 적는 표 없음 | 사실은 design R4 수정·「구멍 A」로 반영. **닫는 방법(새 가드)은 Manager 결정 대기 M11** |
| 2 | BLOCKER | W2 가 `Observe` 의 조기 반환(영구 보류 권위 읽기 실패) 앞에서 사라지고, missing-order 첨부 실패는 기억 보호가 없음 | 맞음 — B10(if@689) → call@711 → return@712 가 W2 앞 | 반영(내 설계의 내부 불일치): D2 「먼저 기억에 쌓고 쓴 뒤 지움」, C3', 시험 추가 |
| 3 | BLOCKER | 판정 읽기와 해제 커밋 사이에 감지기의 체결·거절·`fail_closed` 갱신이 끼어듦 | 맞음 — `ReleaseReconcile` 은 범위·cause 만 재확인(B7·B9), `RecordFill` 은 추적기 잠금 밖에서 커밋(ledger.go call@80 → call@85) | 반영: D5 tx 안 재판정(판정 함수는 하나, 콜백). R8 |
| 4 | MAJOR | C5 가 결속 시점의 번호만 비교 — 결속 뒤의 정정을 모름 | 맞음 | 반영: C5 에 판정 시점 `ResolveCurrentOrderIDScoped` |
| 5 | MAJOR | 계좌 권위를 행에만 확인 — 다른 계좌 관측이 해제 근거가 될 수 있음 | 맞음 — `blocksFor` 가 추적기 계좌로 블록(1418~1429), a110 시험이 그 fail-closed 를 고정 | 반영: C4 계좌·종목 동일성 |
| 6 | MAJOR | `CommittedAt` 은 tx 시작 전 시각이라 「as-of > CommittedAt」 이 인과 순서가 아님 | 맞음 — fills.go:333·373·556. 짝지어진(matched) 주문은 C10 이 못 봄 | 사실은 「구멍 B」·R7 로 반영. **닫는 방법은 Manager 결정 대기 M12** |
| 7 | MAJOR | 기존 회귀 시험이 `Terminal == nil` 로만 돌면 켠 상태의 안전을 증명 못 함 | 맞음 | 반영: §5 에 켠 상태 재실행·교차 시험 추가(기존 시험 무편집) |
| 8 | MINOR | 기존 `evidence` 의 버전 봉투(스키마 무변경) 대안을 평가 안 함 | 맞음 | 반영: Q7 (e). 선택은 **Manager 결정 대기 M13** |
| 9 | NIT | `logCycle call@545` 인용 오류 | 맞음 — call@545 는 `Log.Event`, `logCycle` 호출은 `Run` call@402 | 반영(Q3·Q5 ⑨) |

리뷰가 맞다고 확인해 준 것: Q8 반증 ①(키 순서), Q5 ③(하한 왕복 2회·실패 시 0), `ExitAllowed` production 호출자 0, broker-zero 가 새 cause 를 거부(M7).

### 리뷰 2회차·최종 (2026-11-17, Codex gpt-6-astra, 같은 방식)

원문 `analysis/review/review-2-codex.md`. 판정: 「현재 추천 그대로는 Manager 선택 준비 안 됨」. 5건(BLOCKER 1·MAJOR 4). 1회차 #9 닫힘, #4·#5 는 논리 누락을 닫음
(#2·#3 조건부), Q8 키 순서·하한·`ExitAllowed`·broker-zero 는 재확인. 처분(회차 상한 도달 — 이후 리뷰 없음):

| # | 등급 | 요지 | 확인 | 처분 |
|---|---|---|---|---|
| 1 | BLOCKER | M11 (ii) 의 `committed_at > entered_at` 도 tx 전 시각이라 「T1 캡처 → 블록 커밋 → 체결 커밋」을 놓침 | 맞음 — fills.go:333·373 | 추천을 (ii') `fill_events.id` 워터마크(행을 만든 tx 안에서 읽음)로 고침, 옛 행은 `allFillEventsCTE` 로 귀속·안 풀리면 유지. **선택은 여전히 Manager 결정 대기 M11** |
| 2 | MAJOR | D5 tx 안 재판정의 읽기·콜백 계약이 없음 — 공개 리더는 `j.db`(연결 1개 → 자기 대기), lineage 리더는 충돌 시 쓰기 | 맞음 — journal.go:174, fills.go:1010~1011, lineage.go:334·419~420 | 반영: D5 「트랜잭션 계약」(tx 묶인 읽기 전용 판본, 쓰기 없는 lineage, 순수 콜백, `Tracker`·공개 `Journal` 재진입 금지), 변이 m19 |
| 3 | MAJOR | 결속 키에 계좌·원래 번호가 없어 `INSERT OR IGNORE` 가 다른 계좌 신원을 삼키고 기억을 지움 | 맞음 | 반영: 키 확장, 「커밋 뒤 같은 tx 읽기로 전부 확인했을 때만 기억 삭제」, 변이 m20 |
| 4 | MAJOR | M12 (iii) 무장에 무효화 계약이 없음(무장 뒤 결속 추가 반례) | 맞음 | 반영: 무장을 (행 id, 결속 digest, 스냅숏 digest)에 묶고 바뀌면 재무장, 변이 m22. **선택은 Manager 결정 대기 M12** |
| 5 | MAJOR | spec delta(재기동 같은 판정·첫 통과 비교 해제)가 추천안과 어긋나고 tasks 가 M1~M10 만 셈 | 맞음 | 반영: spec 재기동 문장을 「장부 근거가 같으면 같은 판정, 기억은 늦추는 데만」으로, design §6 끝에 결정별 spec·시험 변경 표, tasks 1.8 을 M1~M13 으로. 첫 통과 해제 scenario 는 M12 결정 뒤 고침(결정 전 판본임을 명시) |

### Manager 결정 대기 (리뷰 제안·새 가드·범위 — 문서에 채택하지 않음)

- **M11**(구멍 A, 착수 차단): 행 뒤 투영 거절 — 추천 (ii') `fill_events.id` 워터마크.
- **M12**(구멍 B, 착수 차단): `CommittedAt` 순서 — 추천 (iii) 무장-후-해제(digest 무효화 포함).
- M13: 결속 저장 (a) 새 표 vs (e) `evidence` 버전 봉투(리뷰 1 #8).
- M4: C10(같은 번호의 외부 주문이면 유지) — 내가 넣은 새 가드라 확인 요청.
- 리뷰가 그 밖에 제안한 「모든 생산자가 실격을 원자 기록」(1회차 #1 Change)은 M11 (i) 로만 적음.

### 남은 위험(리뷰 뒤)

- R5(결속 쓰기 실패 + 재기동 + 수량 분쟁 소멸 + 종결 증명) — M9.
- R1·R2·R3 — a191 잔여 위험과 같은 계열, 이 change 가 새로 여는 모양 아님.
- 2회차가 지적한 D5·결속·무장 계약은 문서 수준 반영이며 **재리뷰를 받지 않았다**(회차 상한). 2단계 코드 리뷰가 그 계약을 시험·변이로 확인해야 한다.

## Manager 결정 (2026-11-17, 1단계 보고 뒤)

M1·M13 = (d) 프로세스 기억 결속(스키마 36 유지, 결속 표·봉투 없음) + I3 개정, M11 = (ii') 체결 id 워터마크(삽입 호출 전 읽기 허용), M12 = (iii) 무장-후-해제
(digest 무효화), M2 = P3, M3 깨끗한 갈래만, M4 C10 포함, M5 유지 + 사유(주기 성공), M6 `RunOnce` 뒤 새 함수(로그만), M7 broker-zero 는 새 cause 를 안 받음,
M8·M9 소멸, M10 그대로. 필수 조건 1~4(삽입 증명·위조 불가·결속 내용·tx 재판정) — 답은 design D2·D5. §7 후속 순서 변경(하한 범위·편입 가시성 앞).
결정 반영 개정(개정 2) 전 판본은 change 밖 스크래치에 백업, 개정 diff 는 `analysis/review/revision-2.diff`.

### 리뷰 3회차 — 결정 반영분 한정 (2026-11-17, Codex gpt-6-astra, `codex exec -s read-only`, Manager 허용 1회)

원문 `analysis/review/review-3-decision-revision-codex.md`. 판정 「NOT READY — #1~#3 해결 뒤」. 4건(MAJOR 3·MINOR 1). 리뷰가 확인해 준 것:
- 필수 조건 1 닫힘 — `EnterReconcile` 의 실패·기존 행 반환은 전부 `false`, 유일한 `true` 는 삽입·커밋 뒤이며 그 id 를 돌려줌(reconcile_states.go:176~229);
  수량 먼저인 키 충돌은 수량 블록이 남고 그 블록엔 missing-order 신원이 없음(mismatch.go:1404~1429·719~724).
- 워터마크는 거절된 투영을 포함한 양수 체결에 대해 건전(옛 행 귀속·미귀속 거절 조건대로). 스냅숏 거절은 이벤트 없이 커밋됨(fills.go:385~409) → C7·digest 가 필수.
- 직렬 주기라 무장-후-해제가 성립(reconcileloop.go:397~428). D5 는 journal 소유 콜백 타입·tx 읽기·얼린 입력으로 import 순환·교착 없이 구현 가능. AST 좌표 오류 없음.

| # | 등급 | 요지 | 확인 | 처분 |
|---|---|---|---|---|
| 1 | MAJOR | pending 으로 남았다 재시도로 삽입된 블록의 결속이 그 사이 관측(둘째 주문·수량 흔적)을 잃음 | 맞음 — B17(if@720)이 첫 블록을 지키고 `pendingBlocksForPersistence`(call@766)가 깨끗한 비교에서도 재시도 | 문서 반영 — **보수 쪽 선택지**(「첫 시도 삽입만 결속」, N20·m20). 리뷰의 다른 선택지(pending 이력을 따로 쌓아 넘김)는 **Manager 결정 대기 M14** |
| 2 | MAJOR | 외부 포지션만 있는 비교는 깨끗한 갈래로 가고 그 권위 읽기 조기 반환(B5 if@646 → return@647)이 결속 갱신보다 앞 | 맞음 | 문서 반영(내 설계의 내부 불일치): 결속 갱신을 갈래 나누기 전(call@634 뒤·B2 앞)으로, 시험 확장 |
| 3 | MAJOR | 같은 객체 `Restore` 가 결속·무장을 남김 — 「복원된 블록 제외」와 모순 | 맞음 — `Restore` 는 streak·크레딧만 지움(954~966) | 문서 반영: `Restore` 잠금 안 결속·무장 지우기 1줄(기존 함수 편집 +1, D11), `Refresh` 는 만들지 못할 뿐 지우지 않음을 명시, N21·m21 |
| 4 | MINOR | spec 이 워터마크(삽입 **전** 읽기) 구간과 다른 유지 조건을 시나리오에 안 담음 | 맞음 | 문서 반영: ADDED 요구 1 문장·시나리오 수정, 복원 시나리오 추가 |

3회차 반영분은 **재리뷰를 받지 않았다**(Manager 허용 회차 소진).

### Manager 결정 대기 (3회차 뒤)

- **M14**: 재시도 삽입 블록 — 지금 문서는 「결속 안 함(운영자 해제)」. 리뷰 대안 「pending 동안의 신원·흔적을 따로 쌓아 삽입 증명 뒤 넘김」을 고를지.
- 3회차 반영 4건(특히 #1 의 정책 선택, #3 의 `Restore` 편집 추가)의 승인.

## Manager 결정 (2026-11-17, 개정 2 검증 뒤) — proposal-freeze

- M14 = 현재 문서대로(첫 시도 삽입만 결속, 재시도 삽입은 결속 없음 — pending 이력 넘기기 불채택, KISS).
- 3회차 반영 4건 전부 승인. 문서 재리뷰 없음 — 2단계 시험·변이(m19~m21)·코드 리뷰가 확인.
- 이후 Requirement 문장 변경이 필요하면 구현을 멈추고 보고.

## 2단계 기록 (구현 팀메이트 Opus, 2026-11-17)

### 순서

FLM 번들 7(편집 전 AST = `analysis/pre-edit-ast/`) → 기존 분기 커버 측정(`analysis/coverage/base-branch-counts.txt`) → 새 파일만 넣은 골격에서
RED 관측(`analysis/red/README.md` — 단언 실패 10건, 컴파일 실패 0) → 기존 함수 편집(GREEN) → 편집 뒤 AST·재번호(`renumber.py`, (종류, 소스 문장)
정렬)·FLM/BTM(`gen_flm.py`) → 변이(`analysis/mutation/`) → 선택 시험·lint(`analysis/green/`) → 코드 리뷰 2회.

### 코드 리뷰 1회차 (2026-11-17, Codex gpt-6-astra, read-only, 적대적 Eng)

원문 `analysis/review/code-review-1-codex.md`(대상 `code-review-1.diff`). 판정 「C12 수정·변이 증거 부족」. 처분:

| # | 등급 | 요지 | 확인 | 처분 |
|---|---|---|---|---|
| 1 | MAJOR | C12 가 종목 없는 legacy 체결(`legacyUnscoped` — fills.go:249)을 놓침; 다른 계좌 같은 종목 체결은 반대로 막음 | 맞음(관대 쪽 구멍) | **고침**: 워터마크 뒤 「그 종목 또는 종목이 빈」 체결이면 유지(계좌 무관 — 보수 쪽 상위집합). 시험 `TestA192TheProofCountsUnattributableAndForeignFills`, 변이 m30 |
| 2 | MAJOR | 변이 원장이 컴파일 실패를 CAUGHT 로 셈(m10·m12·m24), 살아남은 4 | 맞음 | **고침**: 하네스가 빌드 실패를 NOT_REACHED 로 분류, 잘못 쓴 변이 3개 수정, 살아남은 4(m2·m17·m22·m25)는 시험으로 못 박음(아래 변이 절). fail_closed 만 바뀌는 tx 끼워 넣기 시험 추가 `TestA192AFailClosedFlipBetweenVerdictAndReleaseKeepsTheBlock` |
| 3 | MINOR | 유지 사유를 버림(M5 「유지 + 사유」) | 맞음(결정 미구현) | **고침**: 결속마다 마지막 판정 사유(tx 재판정은 `tx_` 접두), `Tracker.OrderTerminalHoldReasons`, 사유가 바뀔 때만 로그 `reconcile.block_release_held`(종목·사유뿐). 변이 m31·m32 |
| 4 | MINOR | 정정 흔적 검사가 design C5(확정 lineage 자식)보다 넓음(그 번호를 겨눈 AMEND 시도면 결과·범위 무관 유지) | 맞음 | 구현은 보수 쪽 확대 — **Manager 결정 대기 M15** |

### 코드 리뷰 2회차·최종 (2026-11-17, 같은 방식)

원문 `analysis/review/code-review-2-codex.md`. 판정 「새 BLOCKER/MAJOR 없음」(정적 — 샌드박스가 go 빌드 임시 디렉터리를 막아 시험 실행 못 함). 처분:

| # | 등급 | 요지 | 처분 |
|---|---|---|---|
| 1 | MINOR(실결함) | 조정 해제로 블록이 사라져도 결속·「armed」 사유가 남음 | **고침**: `updateOrderTerminalBindings` 가 그 종목 블록이 없으면 결속을 버림(새 파일 함수). `TestA192ACreditWinsOverTheTerminalProof` 에 사유 0 단언, 변이 m33 |
| 2 | MINOR(제안) | C12 의 계좌 무관·종목 없는 체결 거부, C5 의 넓은 정정 거부를 Manager 수용으로 기록 | **Manager 결정 대기 M15·M16** |

리뷰가 지적한 「run-4 가 비어 있었다」는 리뷰 시점에 실행 중이었음 — 완료 원문 `analysis/mutation/mutation-run-4-m11-m12.txt`. 최종 원장은 `mutation-run-5-final.txt`.
리뷰 회차 상한 도달 — 2회차 반영분(#1)은 재리뷰 없음(변이 m33 으로 확인).

### 계획(D11) 대비 구현 차이 — Requirement 문장 변경 없음

- 편집한 기존 함수 7개 = D11 표 그대로(`Observe`·`persist`·`blocksFor`·`Restore`·`ValidReconcileReleaseCause`·`buildGateway`·`RunOnce`). `check_analysis.py` 가
  계산한 required 함수도 7(`analysis/green/check-analysis-*.txt`).
- 타입 필드: D11 대로 + `ReconcileDriver.terminalHoldLogged`(M5 사유 로그의 바뀜 기억 — 타입 편집, 함수 아님), `Outcome` 무변경.
- C12: design 은 「그 계좌·종목」, 구현은 「그 종목(계좌 무관) 또는 종목 없는」 — 더 많이 유지(보수). Requirement 의 SHALL NOT(계좌·종목 체결이면 해제 금지)는 그대로 만족.
- C5: design 은 「대체하는 확정 정정이 없음」, 구현은 「대체 간선(범위·확정 무관) 또는 그 번호를 겨눈 AMEND 시도가 없음」 — 더 많이 유지(보수, 규칙 복제 회피).
- 유지 사유 로그 이벤트 `reconcile.block_release_held` 추가(M5 구현) — D7 의 로그 이벤트 하나에 하나 더함.

### Manager 결정 대기 (2단계 리뷰 뒤)

- **M15**: C5 를 넓게(그 번호를 겨눈 AMEND 시도·대체 간선이 하나라도 있으면 유지) 둔 것의 수용 — 대안은 확정 lineage 자식만 보는 판본(규칙 복제 필요).
- **M16**: C12 를 계좌 무관 + 종목 없는 체결까지 유지로 둔 것의 수용 — 대가: 다른 계좌의 같은 종목 체결·옛 판본 관측이 있으면 운영자 해제.
- 유지 사유 로그(`reconcile.block_release_held`)를 사유가 바뀔 때마다 쓰는 것(M5 구현 방식)의 수용.

## Manager 결정 (2026-11-17, 5.1 검증 1차 뒤)

- **M15 = 수용** — C5 를 넓게(그 번호를 겨눈 AMEND 시도나 대체 간선이 하나라도 있으면 유지, 범위·확정 무관). 근거: 유지 쪽 초집합, lineage 규칙 복제 없음.
  운영 장부 읽기 전용 집계(Manager, 2026-11-17, 값 없이 개수만): `mutation_attempts.kind` 는 PLACE 뿐, AMEND 0건 → 지금 막는 정상 입력 0.
- **M16 = 수용** — C12 를 「그 종목(계좌 무관) 또는 종목이 빈 체결」로. 같은 집계에서 `fill_events` 105행 중 종목이 빈 행 0건.
- **유지 사유 로그 `reconcile.block_release_held` = 수용** — M5 의 구현 방식(사유가 바뀔 때만, 종목·사유만).
- 위 셋은 Manager 가 승인한 Requirement 문장 변경: spec delta 의 C12 문장·시나리오(「그 종목의 체결(계좌 무관) 또는 종목을 알 수 없는 체결」),
  lineage 시나리오(「그 번호를 겨눈 정정 시도나 대체 기록이 하나라도(확정 여부 무관)」), 로그 요구에 유지 사유 로그 한 문장. design D3(C5·C12)·
  N9·N16·§6·D7 을 같이 맞춤. `openspec validate --strict` valid.

### Manager 독립 변이 표본 1회차 — 12개 중 CAUGHT 10 · SURVIVED 2

원문 `analysis/mutation/manager-sample-run-1-before-test.txt`(시험 추가 전). 살아남은 g8(해제 UPDATE 의 `AND released_at IS NULL` 삭제)·g9(`changed != 1` 검사
삭제)는 tx 안 재판정의 C2(`state.Active()`)가 먼저 막아 장부 쪽 고정이 혼자 일한 적이 없었기 때문. 새 가드 없이 시험 하나 추가:
`TestA192TheReleaseNeverOverwritesARowAlreadyReleased`(OPERATOR 로 이미 풀린 행에 판정이 늘 참인 콜백으로 해제 → released=false + 오류, 행의 release_cause·
released_at·evidence 그대로). 12개를 `mutate.py` 의 m34~m45 로 옮겨 그 자리에서 다시 돌림 — 결과는 `analysis/mutation/README.md`.

## 남은 위험 (5.1 뒤 추가)

- **종결 스냅숏 범위와 결속 신원의 일치 — 미측정.** 종결 스냅숏의 범위(시장·거래일·매도/매수)는 감지기가 브로커 응답에서 유도하고, 결속 신원은
  `TrackedFillOrders` 행에서 온다. 둘이 같아야 스냅숏을 찾는다 — 다르면 `terminal_not_found` 로 유지(안전 쪽)이고 운영자 해제로 돌아간다. 감지기 →
  장부 → 대사를 한 시험에서 잇는 실행은 없다(엔진 시험은 `RecordFill` 을 손으로 부름).
- 코드 리뷰 2회차 #1 반영분(사라진 블록의 결속 버리기)은 재리뷰 없이 변이 m33 으로만 확인.
