# a192 design — missing-order 종목 차단의 종결 증명 해제

> 1단계 설계. 분기·early return·side effect 를 근거로 쓰는 문장은 `tools/logic-map` AST 열거에서 옮겼다(좌표 `B<n>(kind@line)`, `call@line`,
> `return@line`). AST 가 열거하지 못하는 것(복합 리터럴의 필드 값, SQL 문자열, 상수, 대입문)은 **「손으로 읽음」**으로 표시했다.
> 근거 종류 표기: **[AST]** = AST 열거 · **[손]** = 손으로 읽음(SQL·리터럴·상수) · **[미실측]** = 운영·브로커에서 재지 않음.
> AST 원문 `analysis/cited-ast/*.ast.json`, 요약 `analysis/cited-ast/summary.txt`, 손으로 읽은 목록 `analysis/cited-ast/README.md`.
> 기준 HEAD `6277ec34e`(= base). 운영 장부·브로커는 읽지 않았다(지시상 금지) — 운영 사실은 a191 기록과 Manager 서술을 인용만 한다.
>
> **개정 2(Manager 결정 2026-11-17 반영)**: M1·M13 = (d) 프로세스 기억 결속(스키마 변경 없음), M11 = 체결 id 워터마크, M12 = 무장-후-해제, M2 = P3.
> 앞 판본의 결속 표·스키마 37·W2·`bindingUnwritten` 서술은 지웠다(백업은 change 밖). 리뷰 1·2회차가 연 구멍 A·B 는 C12·C13 이 닫는다. 리뷰 3회차 처분은 `review.md`.

## 1. 지금의 기전 — 설계 질문 Q1~Q6

### Q1. missing-order 종목 차단은 정확히 언제 생기고 장부에 무엇이 적히나

- **첫 blocking 비교에서 즉시** 생긴다. 유예 0. [AST] `Tracker.Observe` B2(if@637 `!diff.BlocksEntry()`)가 거짓 → B3(else@684) →
  `blocksFor`(call@685). `blocksFor` 는 수량 불일치 B1(range@1404) 다음에 missing-order B2(range@1418)마다 블록 하나를 만든다. 새 키면
  B17(if@720) → `pending` 으로 메모리·게이트에 먼저 걸고(`syncGate` call@759), `persist`(call@767) B2(range@869) → `EnterReconcile`(call@880).
- 블록 값 [손](mismatch.go:1419~1430): `Scope: symbol`, `Market: missing.Market`, `Cause: QUANTITY_MISMATCH`, `Reason: reconciliation_mismatch`,
  `Release: adjusted_reconcile`, `Detail: "order <현재 번호> on <종목> is not in the account's open list"`.
- `persist` 가 장부에 넘기는 것 [손](mismatch.go:880~885): `AccountRef`·`Symbol`·`Cause`·`Evidence(=Detail)` **뿐** — `ScopeMarket`·`ID` 없음.
  그래서 행은 `scope_market = NULL`(시장 정보 소실), `id = reconcileStateID(계좌, 종목, cause, 시각, "")`(call@212).
- 운영 행의 열 값(Manager 가 나열한 열 기준): `id`(파생 해시) · `account_ref` · `symbol` · `cause = QUANTITY_MISMATCH` · `evidence = 위 Detail` ·
  `entered_at = 장부 시계` · `released_at/release_cause = NULL` · `broker_zero_observation_id/digest = NULL` · `scope_market = NULL`.
- **같은 종목에 이미 활성 행이 있으면 아무것도 안 쓴다.** [AST] 메모리: `Block.Key` 는 종목 범위에서 `symbol|계좌|종목` 이다(B4 case@251 →
  return@252) — 원인·시장·주문이 키에 없으므로 같은 종목의 수량 블록과 missing-order 블록, 주문 둘의 블록이 **한 키**다; `blocksFor` 가 수량을
  먼저 만들기 때문에 같은 비교에 둘 다 있으면 수량 블록이 이긴다(B17 if@720 `!exists`). 장부: `EnterReconcile` B7(switch@202) → B8(case@203)
  기존 활성 행 → return@204(첫 관측 유지). 투영 훅의 `enterReconcileScopeInTx` 도 B3(if@401 활성) → return@405.
- 결론: **차단을 만든 주문 번호는 `evidence` 문구에 첫 번째 하나만** 남고, 같은 `cause = QUANTITY_MISMATCH` 를 세 생산자가 쓴다 — 추적기의 수량 블록,
  추적기의 missing-order 블록, 투영 거절(과매도, `reconcileCauseFor` B1 if@429 → return@430, 상수 값 [손] position_projection.go:428~433).
  **cause 열로는 행의 출처를 가를 수 없다.**

### Q2. 오늘 그 차단을 푸는 경로 전부

| 경로 | 근거 | missing-order 만 있는 종목 |
|---|---|---|
| ① 조정 크레딧 + 나중 비교 일치 → `ADJUSTMENT_APPLIED` | [AST] `Observe` B2 → B6(range@653) → B7(if@654) → B8(if@662 `creditUsableBy`) 통과 → B9(if@670) 통과 → `Cleared`(append@682) → `persist` B6(range@899) → `ReleaseReconcile`(call@900, cause [손] mismatch.go:903) | **불가.** 크레딧은 `ConvergeQuantities` 의 defer `AdjustmentApplied`(call@189) 하나뿐(CodeGraph callers: production 1)이고 그 함수는 `diff.Quantities` 가 비면 B1(if@144) → return@145 |
| ② 운영자 `reconcile-resolve` | [AST] `runEngineReconcileResolve` 엔진 잠금(call@133) → 3회 안정 스냅숏 → B21(if@198 `diff.BlocksEntry()`) → B22(if@201 validate) → `Tracker.Resolve`(call@204) → `ReleaseReconciles`(call@1343) 원자 해제, `OPERATOR` | 가능(엔진 정지가 전제) |
| ③ 다른 프로세스가 행을 해제한 뒤 `Refresh`/`Restore` 가 투영에서 지움 | [AST] `Refresh` B9(range@1020)는 활성 행만 다시 담음 | 해제 자체가 아님 |

- `reconcile_states.released_at` 을 쓰는 곳은 저장소 전체에서 `ReleaseReconcile`·`ReleaseReconciles` 두 곳뿐이다(grep `SET released_at` — 다른 결과는
  `risk_bucket_owners`·exit snapshot 표). 그 둘의 production 호출자는 `persist`(①)와 `Resolve`(②)뿐이다(CodeGraph 1.6.0).
- 따라서 **missing-order 만 있는 종목에 자동 해제 경로는 0** — 단서 하나: 같은 종목에 **별도로** 수량 불일치가 생겨 수렴(조정)이 기록됐다면
  ①이 그 키의 블록(= missing-order 블록과 같은 키)을 풀 수 있다(`hasBlockFor` B2 if@547 은 cause·release 만 본다). 즉 오늘도 「missing-order 블록은
  조정이 있으면 풀린다」는 길이 우연히 있다.

### Q3. a191 로 종결이 적힌 다음 비교에서 그 주문은 빠지나, 깨끗한 비교에서 `Observe` 는 무엇을 하나

- **빠진다(조건부).** [AST] `LocalStateFromJournal` 은 `TrackedFillOrders`(call@201)의 결과만 B4(range@212)로 담고, 추적 SQL 은 `terminal = 0`
  만 돌려준다 [손](fills.go:1604·1627). a191 Q5(a)의 예외(빈 범위 legacy 행)는 그대로 남는다.
- 깨끗한 비교(`BlocksEntry` 거짓)에서 [AST]: B2(if@637) 참 → `clearStreaks`(call@642 — 모든 승격 streak 소멸) → B4(if@643)·B5(if@646) 영구 보류 정리
  → B6(range@653) 블록마다: B7(if@654) cause ≠ `QUANTITY_MISMATCH` 또는 release ≠ `adjusted_reconcile` 이면 건너뜀(영구 블록은 `operator_only` 라
  여기서 빠짐) → **B8(if@662) 크레딧이 없으면 `AwaitingAdjustment`(append@667) 후 continue** → (크레딧이 있어도) B9(if@670) 그 종목이 분쟁 중이면
  `AwaitingAdjustment`(append@679) → 아니면 `Cleared`(append@682).
- missing-order 블록은 늘 B8 에서 멈춘다. 그 수는 `RunOnce` 의 `cycle.Awaiting`(len@456)으로 `reconcile.clean` 줄의 `awaiting` 필드에만 남는다
  (`Run` → `logCycle`(call@402) → `Log.Event`(call@545); 리뷰 1 #9 로 인용 수정). 즉 사고 뒤 운영 로그는 매 분 「clean, awaiting ≥ 1」 이다.
- 비교가 다른 종목 때문에 막히는 동안(B3 else@684)에는 해제 판단 자체가 없다 — 오늘의 자동 해제는 **비교 전체가 진입을 막지 않을 때만** 일어난다.

### Q4. 영구 승격 — streak, 영구 행, 재기동

- streak [AST]: blocking 관측마다 `promotionKeys`(call@634)가 이 비교의 분쟁 키 집합을 만들고(missing-order 는 `missingOrderPromotionKey` B1(if@411)
  — 계좌·시장·거래일·종목·방향·주문번호 중 하나라도 비거나 계좌가 다르면 키 없음 → return@414), `advanceStreaks`(call@716) 가 **직전 비교에도 있던 키만**
  +1 하고 나머지를 버린다(B1 range@1055 — `next` 맵 교체). 깨끗한 비교는 `clearStreaks`(call@642). `thresholdKey`(call@727)가 3 이상인 키를 찾으면
  B18(if@727) → 계좌 범위 영구 블록(값 [손] mismatch.go:728~737: `Scope: account`, `Cause: QUANTITY_MISMATCH`, `Reason: reconciliation_mismatch_permanent`,
  `Release: operator_only`, `Detail: permanentPromotionDetail(3, key)`).
- 운영 주기 [손]: `ReconcileDriver.Run` 은 `RunOnce`(call@401) 뒤 `DefaultReconcilePeriod = 60s`(reconcileloop.go:95) 를 잔다(call@403). 한 주기 안에
  스냅숏 2회(2s 간격). `Tracker.MinInterval`(30s)을 읽는 `Due` 는 production 호출자 0 이다(grep). 그러니 영구 승격은 **같은 분쟁이 약 2~3분**
  이어지면 일어난다.
- 영구 행이 담는 것: `symbol = NULL`, `cause = QUANTITY_MISMATCH`, `evidence = "same canonical dispute reached permanent threshold: threshold=3
  kind=missing-order account=… market=… day=… symbol=… side=… orderID=…"`(`permanentPromotionDetail` B3 case@432 → return@433, 문자열 [손]).
  **결속은 이 문구뿐이다** — 구조화된 열 없음.
- 재기동 [AST]: `Restore` B3(range@943)·B4(if@944)로 이 계좌의 활성 행을 `blockFromReconcileState`(call@948)로 블록에 되돌린다. 그 함수: B1(if@1374)
  `QUANTITY_MISMATCH` 면 `adjusted_reconcile`, B2(if@1377) 계좌 범위면 B3(if@1380) 영구·`operator_only`. **`Block.Market` 은 되살리지 않는다**(리터럴에
  필드 없음 [손] mismatch.go:1363~1373). 지워지는 것 [손](mismatch.go:959~969): `streaks = nil`, `pendingPermanent = false`, `pendingPermanentKey` 빈 값,
  `adjusted = nil`(크레딧은 일부러 복원 안 함), `failures = 0` → 영구 행이 있으면 B5(if@967) `failures = 3`. 게이트는 B6(if@970)
  `RebuildReconcileProjection`.
- 결론: 재기동 뒤 남는 것은 **행(차단)과 문구**뿐이다. 「이 차단을 만든 분쟁 키」·「수량 분쟁이 겹쳤는지」·크레딧은 사라진다.

### Q5. 블록을 읽는 소비자와 보호를 줄이는 자리

| # | 소비자 | 근거 | 범위 | 보호 영향 |
|---|---|---|---|---|
| ① | 진입 게이트 | [AST] `syncGate` B4(range@1467) 종목 행 → `BlockSymbol`(call@1472), B9(range@1486) 계좌 행 → `Gate.Block`(call@1491); 재기동 `RebuildReconcileProjection` | 종목 행 = 그 종목 신규 진입, 계좌 행 = 전부 | 진입만 |
| ② | 편입 `judgeHoldings` | [AST] B7(if@140) `ExitEligible` 이면 먼저 continue(관리 중 포지션은 영향 없음) → **B10(if@152) `d.blocked` → continue — 로그·경보·`Unmanaged` 계수 0** | `blocked` = `Permanent()` ∨ `EntryAllowed(market, symbol) ≠ nil`(return@200). `Block.Covers` 는 종목 범위에서 **시장을 보지 않는다**(B4 case@265 → return@273 `EqualFold(symbol)`), 빈 종목이면 전부(B5 if@270 → return@271). 영구 행이면 **모든 종목** | **무관리 보유분이 자동 손절을 못 받는다**(사고에서 실제로 일어난 일) |
| ③ | 청산 확정 하한 | [AST] `ExitObserver.submit` → `applyFloor`(유일 호출자, CodeGraph) → `reconcileFloor.ConfirmedFloor` B2(if@197): 덮는 블록 또는 영구 행이면 적용 → `Holdings`(call@208)·`SellableQuantity`(call@232) 브로커 읽기 2회 → `localOpenSells`(call@243) → `min(보유, 매도가능) − 로컬 미체결 매도`(riskcalc) | 블록이 덮는 종목, 영구면 **모든 종목** | **손절·익절 앞에 브로커 왕복 2회가 붙고, 읽기 실패는 `applyFloor` B2(if@1761) → `reportZeroFloor` → `"0"`(return@1768) — 한 주도 안 나감.** 취소됐지만 종결이 안 적힌 매도는 `LiveOrdersForSymbol`(종결 증거 = `terminal = 1` [손])에 남아 하한을 깎는다(사고 당시 a191 이전 상태) |
| ④ | 포지션 정책 명령 `Runtime` | [손] position_policy_command.go:133~137 `Blocks()` → 화면 | 표시 | 없음 |
| ⑤ | `reconcile-resolve` validate | [AST] `buildReconcileResolveRuntime` B12(if@304) 블록 0 → 거절, B13(range@307)·B14(if@308) cause ≠ `QUANTITY_MISMATCH` → 거절 | — | 해제 경로 |
| ⑥ | 열린 포지션 상한 | [손] open_positions.go:152~180 — 활성 종목 RECONCILE 행을 슬롯으로 셈 | 계좌 | 진입만 |
| ⑦ | 위험 버킷 진입 | [손] risk_bucket.go:295~312 — 그 종목·계좌 범위 활성 행이 있으면 진입 거절 | 종목 | 진입만 |
| ⑧ | 위험 버킷 broker-zero | [손] risk_bucket_owner.go:265~281·676~677 — 그 종목의 **가장 최근 해제 행**의 release cause 가 `RECHECK_MATCHED`/`ADJUSTMENT_APPLIED` 일 때만 근거로 받음 | 종목 | 새 release cause 는 근거가 못 됨(fail-closed — M7) |
| ⑨ | 주기 로그 | [AST] `RunOnce` `cycle.Blocked`(len@454)·`Awaiting`(len@456) → `Run` 의 `logCycle`(call@402) → `Log.Event`(call@545) | — | 없음 |

- 「청산·손절은 블록에 안 막힌다」의 검증: **반만 참이다.** `Tracker.ExitAllowed` 는 분기 0 으로 `true` 를 돌려주지만(return@1304) production 호출자가
  **0** 이다(grep) — 장식이다. 실제 청산 경로는 ③에서 블록의 영향을 받는다: 막히지는 않지만 **늦어지고(왕복 2회), 읽기가 실패하면 0 주가 된다.**
  관리 중 포지션의 손절 발의 자체(②의 B7)와 수동 flatten(exitwiring.go 주석, flatten 은 하한 호출자가 아님 [손])은 블록과 무관하다.
- 이것이 사용자 판단 「차단 하는게 더 위험」을 코드로 뒷받침한다: 영구 행 하나가 **계좌 모든 종목**의 편입을 멈추고 모든 청산에 왕복 2회와 실패 시 0 을 붙인다.

### Q6. 엔진 안 순서 — 「종결 기록」과 「대사의 첫 비교」

- 두 루프는 **따로 도는 goroutine** 이다. [AST] `engineRuntime` 이 `filldetect`·`reconcile`·`exit` 를 `SupervisedLoop` 로 넘기고(call@709),
  `Runtime.Run` B3(range@304) 마다 `go`(go_statements @306) → `loop.Run`(call@308). 서로를 기다리지 않는다. 고정 순서 없음.
- 감지기는 3s 마다(`PollInterval` [손] detect.go:130), a191 의 종결 목록 조회는 그 주문당 첫 조회 즉시·이후 30s 간격(`closedLookupInterval` [손]).
  대사는 60s 마다 스냅숏 2회(2s 간격) → 안정이면 `LocalStateFromJournal`(call@423) → `Compare`(call@428).
- **종결이 먼저 적히면 차단은 안 생긴다**: 대사의 로컬 상태 읽기(call@423)가 종결 기록 뒤면 그 주문은 추적 집합에 없어(Q3) missing-order 가 아니다.
  차단은 「브로커 OPEN 에서 사라짐 → 대사 스냅숏 → 대사 로컬 읽기 → … → 종결 기록」 순서일 때만 생긴다. 창의 길이는 a191 첫 조회가 성공하면
  감지기 한 주기(≈3s + 요청 시간), 종결 목록 반영이 늦어 첫 조회가 `not_found` 면 30s 이상이다 [미실측 — 종결 목록 반영 지연은 a191 N5 미측정].
  또 취소가 대사의 두 스냅숏 사이에 일어나면 두 스냅숏의 `Digest`(OPEN 목록 포함 [손] snapshot.go:174~178)가 달라 그 주기는 비교 없이 끝난다
  (`RunOnce` B1 if@418 → return@419).
- 감지기의 종결 기록은 `JournalLedger.Apply` B7(if@84) → `Refresh`(call@85)를 부르지만 `Refresh` 는 장부의 활성 행을 다시 투영할 뿐 해제하지 않는다.
- 결론: a191 이후에도 차단은 **확률적으로** 생긴다(대개 짧은 창). 생기면 오늘은 사람이 풀어야 하고, 이 change 가 그 경우를 닫는다.

## 2. 설계 질문 Q7~Q11

### Q7. 차단과 그것을 만든 주문을 묶는 방법 — **결정 M1·M13 = (d) 프로세스 기억**(Manager 2026-11-17)

| 안 | 읽을 값이 코드·장부에 있나 | I1 원인 | I2 수량 지배 | 재기동 | 판정 |
|---|---|---|---|---|---|
| (a) 장부에 결속 표 추가(스키마 37) | 새로 쓴다 — 쓸 값은 `Diff.MissingOrders[i]`(`LocalOrder`)에 다 있음 [AST] `Compare` B23(range@477)·B24(if@478) | 충족 | 충족 | 판정 유지 | **기각(Manager)** — 아래 |
| (b) 기존 열만으로 유도 | 없음 — `cause` 는 세 생산자가 공유(Q1), 행에 주문·출처 열 없음 | **반증**: 「그 종목의 엔진 주문 중 행 진입 뒤 종결된 것」으로 유도하면 수량 블록 행이 무관한 주문의 종결로 풀린다. 「행 진입 전 살아 있던 주문 전부」면 표본 0 인 수량 행이 전칭 판정으로 자동 참 | 불가 | — | 기각 |
| (c) evidence 문구 해석 | 첫 주문 하나만(Q1), 투영 거절 문구에도 `[order …]`(position_projection.go:374) | 첫 주문만 | 흔적 없음 | — | 기각(문구는 계약이 아님) |
| **(d) 프로세스 기억 결속** | 삽입 증명은 이미 있다: `EnterReconcile` 은 `(ReconcileState, entered bool, error)` 를 돌려주고, 삽입 갈래만 `entered = true` 와 그 tx 가 쓴 행 id 를 돌려준다 [AST] B8(case@203) → return@204(기존 행 — 리터럴 `false` [손]), 삽입(call@214) → commit(call@222) → return@226(리터럴 `true` [손]) | 충족(D2·D3) | 충족(기억의 수량 흔적 = 실격) | **재기동 = 결속 없음 → 운영자 해제**(fail-closed). 크레딧과 같은 원칙(`Restore` 가 일부러 복원 안 함 [손] mismatch.go:963~966) | **채택** |
| (e) `evidence` 버전 봉투(리뷰 1 #8) | 자유 TEXT [AST] `EnterReconcile` B5(case@186) | 모든 생산자가 원자 갱신해야 충족 | 같음 | 판정 유지 | 기각(Manager) — 문장 칸이 계약을 겸함 |

**(a) 를 기각한 이유(기록으로 남김):** 실계좌 장부 마이그레이션(36→37)과 그 롤백 비용 — v36 이진(a191 release·a187 rollback 핀)은 v37 장부를 `ErrSchemaTooNew` 로
거부하므로 롤백이 「마이그레이션 전 자동 백업 복원」(backup.go:68 [손])이 된다. 그리고 그것이 끌고 오는 표면(`EnterReconcile` 안의 결속 쓰기, 이후 첨부 쓰기 W2,
쓰기 실패를 기억하는 장치, 「쓰기 실패 + 재기동」 잔여 R5)을 통째로 없앤다. (d) 의 대가는 「재기동 전 행은 이 경로로 안 풀린다」 하나이고, 차단부터 해제까지가
1~2주기(≈1~2분)라 그 창에 재기동이 겹칠 일은 드물다. 재기동이 겹치면 오늘처럼 운영자 해제다 — 안전 방향.

**바뀐 I3(Manager 2026-11-17)**: 프로세스 기억은 해제를 늦추거나 포기하게만 한다. 해제 순간의 증명(행 신원·종결 스냅숏·현재 lineage·체결 워터마크)은
해제 트랜잭션 안에서 장부로 다시 읽는다. 이 프로세스가 직접 만들지 않은 행(재기동 전 행, `Restore`·`Refresh` 로 들어온 블록, 다른 생산자가 만든 행)은
이 경로로 풀지 않는다.

### Q8. 계좌 전체 영구 차단의 해제 방식

| 안 | 근거 | 위험 | 범위 | spec 영향 |
|---|---|---|---|---|
| P1 승격 키가 missing-order 이고 그 주문이 종결 확인되고 계좌 전체 비교에 차단 차이 0 → 자동 해제 | 영구 승격의 근거(「다시 봐도 안 고쳐진다」)가 a191 이후 missing-order 에서는 「브로커 단건 조회 결함」으로 설명됨 | 아래 반증 ①~④ | 영구 행 결속(승격 시점 키 집합) 장부화 + `Observe` 깨끗한 갈래의 B7 우회 + 새 cause | 「영구 불일치의 해제는 운영자 확인뿐이다(SHALL)」(spec 3곳)·scenario 「영구 차단의 운영자 해제」·a110 Story limits 를 **뒤집음** |
| P2 엔진을 세우지 않는 운영자 해제(콘솔·엔진 내 명령) | 엔진 정지 = 손절 없음(기억) — 해제 절차 자체가 위험 | 엔진 루프와 같은 프로세스에서 3회 안정 스냅숏·잠금·감사 기록을 다시 짜야 함; 콘솔은 확인 문구 금지(기억) | 별도 Story(UI·명령·권한) | `Reconcile release is durable before it is visible` 의 「engine exclusion」 문장 수정 |
| P3 지금대로 운영자 전용 | spec 그대로 | 엔진 정지 창 | 0 | 0 |

**Manager 잠정안(P1 을 missing-order 종류 승격에 한해)에 대한 반증:**

- **① 승격 키는 하나만 기록되고, missing-order 가 수량보다 먼저 뽑힌다.** [AST] `thresholdKey` B1(range@1077)·B2(if@1078)는 문턱에 닿은 키 중
  `promotionDisputeKeyLess`(call@1078)로 **가장 작은 것 하나**를 고르고, 그 비교는 B1(range@1091)·B2(if@1092)로 배열 원소를 차례로 보는데 **첫 원소가
  `kind`** 다 [손](mismatch.go:1087). `"missing-order" < "quantity"`(바이트 비교). 그래서 같은 비교들에서 missing-order 분쟁과 수량 분쟁이 **함께** 3회
  이어지면(예: 부분 체결 뒤 앱 취소 — 그 종목에 수량 차이와 missing-order 가 같이 생김) 영구 행 문구는 `kind=missing-order` 로 적히고 수량 분쟁은
  문구에 없다. 「기록된 키가 missing-order 면 자동 해제」는 **3회 지속된 수량 분쟁을 숨긴 영구 행을** 주문 종결과 우연한 일치로 푼다 — a083/a110 의
  「원인으로만」 원칙 위반. 고치려면 승격 시점의 **문턱 키 전부**를 장부에 남기고 「전부 missing-order」를 요구해야 한다(P1').
- **② 결속은 재기동에서 끊긴다.** `pendingPermanentKey` 는 `Restore` 가 지우고(Q4), 영구 행에는 문구뿐이다(Q7 (c) 기각 사유와 같음). P1 을 재기동을 넘겨 유지하려면 승격 키 집합을 장부에
  남겨야 하고(스키마 비용), 프로세스 기억이면 재기동 뒤 운영자 해제로 돌아간다.
- **③ 문턱 아래 형제 분쟁**: 승격은 「그 키」로 결정되지만 영구 행은 계좌 전체를 덮어, 문턱에 안 닿은 다른 분쟁들의 진입 차단도 함께 대신한다. 다만
  그 형제들은 각자 종목 행을 가지므로(Q1) 영구 행이 풀려도 그 종목 행은 남는다 — 「계좌 전체 비교 차이 0」 조건이 지금 남은 형제를 막는다. 이것은
  반증이 아니라 P1' 의 필요 조건이다.
- **④ 빈도와 근거의 질**: a191 뒤 missing-order 영구 승격은 a191 이 **2~3분 동안 종결을 확인하지 못했을 때만** 생긴다(Q4·Q6). 그 상황은 정의상
  a191 의 사유 코드(`closed_list_not_found`·`ambiguous`·`row_unreadable` 등)가 지속된 상황이다 — 나중에 종결이 확인되더라도 그 주문은 a191 의 잔여
  위험(a191 design Q3 「거부하지 않을 때 통과하는 것」 (a)~(c))에 가장 가까운 표본이다.
- **⑤ spec·Story**: 영구 차단의 운영자 전용 해제는 spec 문장 3곳(불일치 시 진입 차단 ×2, 조정 credit 요구사항 말미)·scenario 1·a110 Story 의 acceptance
  「Existing permanent blocks never auto-release」로 **명시적으로 결정된** 안전 규칙이다.

**추천**: 이 change 는 **P3 유지**(영구 행은 운영자 전용, 종목 행만 자동). P1 은 반증 ①·②를 닫는 P1'(승격 시점 문턱 키 전부를 장부에 결속하고
「전부 missing-order ∧ 전부 깨끗한 종결 ∧ 그 뒤 계좌 전체 비교 차이 0」)로 **별도 Story** 에서. P2 도 별도 Story. 근거: a191 + 이 change 로 영구 승격까지
가는 경우가 드물어지고(④), P1' 는 Requirement 수준 반전이라 자기 proposal-freeze 리뷰를 받아야 한다. 결정은 M2.

### Q9. 첫 관측 유예(차단을 늦게 거는 것)

Manager 잠정안대로 **채택하지 않는다**(I4 — 차단 조건·시점 불변). 반대 근거는 있다: Q5 ③ — 블록이 걸린 동안 그 종목의 손절은 브로커 왕복 2회를
더 하고 읽기가 실패하면 0 주가 된다. 즉 「짧게 걸렸다 풀리는 차단」도 공짜가 아니다. 그러나 유예는 진입 쪽의 안전망을 같은 창만큼 빼며(그 창에 그
종목 신규 진입 허용), 비용(하한)을 줄이는 정석은 차단을 늦추는 것이 아니라 **원인이 해소되면 빨리 푸는 것**(이 change)이다. 하한의 적용 범위
자체(missing-order 블록에도 하한이 필요한가)는 후속 후보로 적는다(§7).

### Q10. 해제 조건이 거부할 정상 입력과 통과시킬 위험 입력 — §4 의 표

### Q11. 관측 가능성

- 장부: 새 release cause `ORDER_TERMINAL_CONFIRMED`(D5) + evidence 꼬리 `released: missing-order terminal proof — N bound order(s) terminal,
  latest terminal commit <시각>, comparison as-of <시각>`. D5 의 새 메서드가 기존 `ReleaseReconcile` 과 같은 모양(`active.Evidence + " | released: " + …` [손]
  reconcile_states.go:299~301)으로 붙인다.
- 로그: 해제 하나마다 이벤트 `reconcile.block_released`(새 `obs.EventType` 리터럴 — a191 선례대로 `internal/obs` 를 고치지 않음) 필드 `symbol`·`market`
  (결속의 시장)·`release_cause`·`bound_orders`(수). **싣지 않는 것**: 계좌번호, 주문번호, 수량, 가격. 위치 결정은 M6.
- 편입 건너뛰기가 조용한 문제(`judgeHoldings` B10): **후속으로 뺀다.** 근거 — (i) 다른 계약(exit-policy 편입·경보 등급, a095 의 사실 칸·래치 규칙)을
  건드리는 경보 설계 문제이고, (ii) 이 change 가 차단 창을 줄이면 그 침묵의 비용도 같이 준다, (iii) 같은 change 에 넣으면 `judgeHoldings` 본문 편집
  (FLM)과 경보 정책 결정이 High-risk 리뷰 범위를 두 배로 만든다.

## 3. 설계 (Manager 결정 2026-11-17 반영판)

### D1 범위

- 대상: **이 프로세스의 missing-order 관측이 직접 삽입한** 종목 범위·`QUANTITY_MISMATCH`·`scope_market = NULL` 활성 행.
- 제외(전부 블록 유지 → 오늘처럼 조정 규칙 또는 운영자): 재기동 전에 생긴 행, 첫 시도에 삽입되지 않은(재시도로 삽입된) 블록, `Restore`·`Refresh` 로 들어온 블록, 다른 생산자(투영 거절·
  `IDENTIFIER_CONFLICT` 계열)가 만든 행, 이미 있던 행에 겹친 missing-order, 계좌 범위 영구 행(P3), 다른 cause·시장 범위 행.

### D2 결속 — 프로세스 기억, 삽입 증명으로만 생긴다

결속 = `Tracker.bindings[종목]` 한 칸(새 비공개 필드, 새 파일의 타입):

| 칸 | 내용 | 언제 |
|---|---|---|
| `rowID` | 삽입된 `reconcile_states.id` | 생성 시 — `EnterReconcile` 이 `entered = true` 로 돌려준 `state.ID` |
| `orders` | 그 행이 살아 있는 동안 그 종목에서 관측된 missing-order 정규 신원 전부(6요소 + 원래 번호) — 집합에 더하기만 | 생성 시 + 이후 blocking 관측마다 |
| `tainted` | 수량 불일치(또는 외부 포지션)가 그 종목에서 한 번이라도 관측됨 → 그 행이 풀릴 때까지 실격 | blocking 관측마다 |
| `fillWatermark` | 삽입 **호출 전에** 읽은 `max(fill_events.id)` | 생성 직전 |
| `arm` | 무장 상태 `{as-of, digest(rowID, orders, 결속 주문 종결 스냅숏)}` | D4 |

- **생성(삽입 증명)**: `blocksFor` 가 missing-order 블록에 그 신원을 비공개 필드로 싣는다(B2 range@1418 의 리터럴 1필드). `persist` 는 그 블록을 쓸 때
  `t.Terminal`(새 선택 필드)이 있으면 ① `MaxFillEventID` 를 먼저 읽고 ② `EnterReconcile`(call@880)을 부르고 ③ **`entered == true` 일 때만** 결속을
  `{rowID: state.ID, orders: {그 블록의 신원}, fillWatermark: ①}` 로 만든다. 지금 `persist` 는 `entered` 를 B5(if@890 `!entered && cause 다름`)에서만 읽고
  삽입 여부는 버린다 — 그 반환값을 붙잡는 것이 편집의 전부다. **`EnterReconcile` 의 반환을 넓힐 필요가 없다**(필수 조건 2).
- **재시도 삽입은 결속하지 않는다(리뷰 3 #1 반영)**: 결속은 그 블록이 **처음 만들어진 같은 `Observe` 의 첫 `persist` 시도**에서 `entered == true` 일 때만 생긴다.
  쓰기가 실패해 pending 으로 남았다가(B17 if@720 이 블록을 지키고, `pendingBlocksForPersistence` call@766 이 깨끗한 비교에서도 재시도) 나중에 삽입된 블록은
  결속이 없다 — 그 사이의 관측(둘째 주문·수량 흔적)이 결속에 안 들어 있기 때문이다. 블록의 비공개 표시 하나(「첫 시도」)로 가르고, 첫 시도 뒤 지운다.
  대가: 첫 삽입이 실패한 드문 경우 운영자 해제(N20).
- `entered == false`(기존 활성 행 — `EnterReconcile` B8 return@204)이면 결속을 만들지 않는다. 그 행의 출처가 무엇이든(투영 거절·이전 프로세스·다른 생산자)
  이 경로의 대상이 아니다.
- **추가·실격**: `Observe` 가 **갈래를 나누기 전**(분쟁 집합을 만든 call@633·@634 뒤, B2 if@637 앞 — 깨끗한 갈래의 권위 읽기 조기 반환 B5 if@646 → return@647
  과 blocking 갈래의 B10 if@689 → return@712 둘 다보다 앞; 리뷰 3 #2 반영 — 외부 포지션만 있는 비교는 `BlocksEntry` 가 거짓이라 깨끗한 갈래로 간다)과 `persist`
  뒤에 같은 멱등 함수를 부른다: 결속이 있는 종목마다 이 비교의 missing-order 신원을 `orders` 에 더하고, 그 종목이 `diff.Quantities` 또는 `diff.ExternalPos`
  에 있으면 `tainted = true`. 두 번 부르는 이유: 같은 비교에서 처음 생긴 결속은 `persist` 안에서 생기므로, 같은 종목의 둘째 missing-order(블록 키가 같아
  `Observe` B17 이 버린 것)를 `persist` 뒤 호출이 더한다. 장부 쓰기가 없으니 쓰기 실패 처분도 없다.
- **소멸**: 그 행이 이 추적기의 해제로 커밋되면(`persisted.released`) 지움. `Resolve`·운영자 해제 등 다른 길로 행이 사라지면 다음 판정에서 「활성 행 id ≠
  `rowID`」로 실격되고 그때 지움(새 파일 함수 안 — `Resolve`·`Refresh` 무편집).
- 재기동: 새 프로세스의 `Tracker` 는 `bindings == nil` → 남은 행은 전부 유지(운영자 해제). `reconcile-resolve` 의 tracker 는 `Terminal == nil` 이라 결속을 만들지 않는다.
- **같은 객체의 `Restore`(리뷰 3 #3 반영)**: `Restore` 는 streak·크레딧을 지우지만(mismatch.go:954~966 [손]) 새 칸은 모른다. `Restore` 잠금 안에서 결속·무장을 지운다
  (대입 1줄 — 기존 함수 편집, D11). `Refresh` 는 결속을 **만들지 못할 뿐** 이미 있는 결속을 지우지 않는다 — 운영 순서가 매 주기 `Refresh`(call@447) 다음
  `Observe`(call@453)이므로, 지우면 결속이 한 주기도 못 산다. 같은 행인지는 C2 가 행 id 로 가른다.

#### 필수 조건 1 — 「투영 거절이 행을 먼저 만들고, `Refresh` 전에 `Observe` 가 같은 종목 missing-order 로 기억에 새 블록을 만든다」는 실제로 가능한가

**가능하다(오늘).** [AST] `RunOnce` 는 `Refresh`(call@447) 와 `Observe`(call@453) 사이에 잠금을 쥐지 않는다. 체결 감시는 다른 goroutine 이고(`Runtime.Run` go@306),
`JournalLedger.Apply` 는 `RecordFill`(call@80)을 **추적기 잠금 없이** 커밋한 뒤에야 `Refresh`(call@85)를 부른다. 그 `RecordFill` tx 안에서 투영 거절이
`enterReconcileScopeInTx` 의 삽입(B4 if@408 의 `tx.Exec`)으로 종목 행을 만들 수 있다. 그러면 `Observe` 의 `t.blocks` 에는 그 행이 없어 B17(if@720)이 새 키로
pending 블록을 넣고, `persist`(call@767) → `EnterReconcile`(call@880) → B8 기존 행 → return@204(`entered = false`). 투영 과매도의 cause 도
`QUANTITY_MISMATCH` 라 `persist` B5(if@890)를 지나 그 블록은 `durable`(append@897)이 되고 `Observe` B20(range@768)이 `pending = false` 로 만든다 — **오늘의
추적기는 그 행을 자기가 만든 줄 안다**(귀속 실패면 cause 가 달라 B5 → return@893 오류).
**닫는 법**: 결속은 오직 `entered == true` 에서 생기므로 이 순서에서는 결속이 생기지 않는다 → 이 경로로 안 풀린다. 시험
`TestA192ARowAnotherProducerInsertedFirstIsNeverBound`(투영 거절 행을 `Refresh` 와 `Observe` 사이에 끼워 넣음) + 변이 m1.

#### 필수 조건 2 — 삽입 증명의 모양과 편집 자리 수

- 증명 = 같은 `EnterReconcile` 호출이 돌려준 `entered == true` 와 `state.ID`. 둘 다 삽입 tx 의 결과다(`return@226` 은 commit call@222 뒤). evidence 문구는 보지
  않는다.
- 자리 수 [AST]: `EnterReconcile` 의 반환을 바꾸지 않으므로 그 호출자 3곳(`persist` call@880, `guardTrackedFillIdentity`, `recordScopedLineageConflict` — CodeGraph)은
  무편집. 편집은 `persist` 안 **1자리**(call@880 의 결과를 붙잡음)이고, `entered` 를 읽는 곳은 오늘 B5(if@890) 1곳뿐이다. 새 파일의 「삽입 여부 보고」
  메서드는 `EnterReconcile` 의 규칙(한 범위 한 행·첫 관측 유지)을 복제하게 되므로 **추천하지 않는다** — 반환이 이미 그 보고다.
- 남는 모양: `EnterReconcile` 이 커밋한 뒤 전송 오류를 돌려주면(`entered` 를 못 봄) 블록은 pending 으로 재시도되고, 재시도는 B8 → `entered = false` →
  결속 없음 → 유지(fail-closed).

#### 필수 조건 3 — 결속에 담는 것

위 표 그대로: 삽입된 행 id, 관측된 missing-order 신원 전부(집합), 수량·외부 포지션 흔적(실격 — 풀릴 때까지), 체결 워터마크(+ 무장).

#### 필수 조건 4 — 해제 트랜잭션 → D5

### D3 판정 함수 — 하나, 새 파일, 순수

`internal/reconcile/order_terminal_release.go` 의 순수 함수(초안 `orderTerminalReleaseVerdict`). 입력 = **얼린 비교 입력**(추적기 계좌, `diff.AccountRef`·`AsOf`,
그 종목의 분쟁 여부, `diff.ExternalOrd` 의 주문번호 집합, `BlocksEntry`) + **결속**(rowID·orders·tainted·fillWatermark·arm) + **장부 증명**(그 종목 활성 행,
결속 주문마다 종결 스냅숏 또는 오류, 결속 주문마다 대체 흔적(대체 간선·그 번호를 겨눈 AMEND 시도) 존재 여부, `fill_events.id > fillWatermark` 인 그 종목(계좌 무관)·종목 빈 행 존재 여부, 귀속 못 한 행 존재 여부).
출력 `{verdict ∈ 유지·무장·해제, reason, digest, boundOrders, latestTerminalCommit}`. 처음 실패한 조건의 사유로 「유지」:

| # | 조건 | 거부 사유(초안) |
|---|---|---|
| C1 | 결속이 있다(`bindings[종목]` — 이 프로세스의 삽입 증명) | `not_inserted_by_this_process` |
| C2 | 그 종목의 활성 행 id = `rowID`, 행이 종목 범위·`QUANTITY_MISMATCH`·`scope_market` 빈 값 | `row_changed` |
| C3 | `tainted == false` | `quantity_dispute_seen` |
| C4 | `orders` 가 하나 이상이고 모든 신원의 6요소가 비지 않으며, 신원 계좌 = 행 계좌 = 추적기 계좌 = `diff.AccountRef`, 신원 종목 = 행 종목 | `identity_incomplete` / `account_mismatch` |
| C5 | 모든 신원에서 현재 번호 = 원래 번호, **그리고** 장부에 그 원래 번호를 겨눈 AMEND 시도나 대체 간선이 하나도 없음(범위·확정 여부 무관 — M15) | `lineage_moved` |
| C6 | 모든 신원의 종결 스냅숏 읽기가 오류 없이 레코드를 돌려줌 | `terminal_read_failed` / `terminal_not_found` |
| C7 | 모든 레코드가 `Terminal ∧ !FailClosed` | `not_clean_terminal` |
| C8 | `AsOf`·모든 `CommittedAt` 이 시각으로 읽히고 `AsOf > max(CommittedAt)` | `comparison_not_after_terminal` |
| C9 | 그 종목이 이 비교의 분쟁(수량·missing·외부 포지션)에 없음 | `symbol_disputed` |
| C10 | `diff.ExternalOrd` 에 결속 번호와 바이트가 같은 주문이 없음(M4 — 포함 결정) | `terminal_order_still_open` |
| C11 | `BlocksEntry` 거짓(M3 — 호출 자리가 보장) | — |
| C12 | `fill_events.id > fillWatermark` 인 그 종목의 행(계좌 무관)도, 종목이 빈(귀속할 수 없는) 행도 없음(M11 (ii'), M16) | `fill_after_block` |
| C13 | 무장: `arm` 이 없거나 digest 가 다르면 → **무장**(이번엔 해제 안 함). 같은 digest 이고 `AsOf > arm.as-of` 면 → 해제(M12 (iii)) | `armed` |

- C6 의 스냅숏 읽기는 C4 를 통과한 신원만 넘긴다 — 범위가 불완전하면 legacy 빈 범위 행을 읽기 때문이다 [AST] `lookupFillSnapshotScoped` B1(if@988) → return@989.
  조회 번호는 원래 번호(감지기가 스냅숏을 쓴 번호).
- C7: 실패 갱신은 `terminal` 을 두고 `fail_closed = 1` 만 세운다 [손] fills.go:834~857.
- C12 의 근거: 투영 거절은 체결 delta 에서만 생기고 그 delta 는 거절돼도 같은 tx 에서 `fill_events` 에 남는다(fills.go:464~472 가 훅 :551 앞; 활성 행이면 거절은
  nil — position_projection.go:101~103·401~405 — 리뷰 2 확인). SQLite 쓰기는 직렬이라 AUTOINCREMENT id 순서가 커밋 순서다. 워터마크는 삽입 **전**에 읽으므로
  그 사이의 체결은 「뒤」로 세어진다(보수). 옛 행은 `allFillEventsCTE`(fills.go:1046~1057)로 귀속. 체결 0 인 취소(delta 0)는 C12 가 못 보고 C7·C13 이 맡는다.
- C13 의 근거: 운영 `Run` 은 주기를 직렬로 돌고(reconcileloop.go:397~403) 각 `RunOnce` 는 수집 뒤에 로컬을 읽는다(:417~428 — 리뷰 2 확인). 무장 관측은 종결
  스냅숏을 이미 읽은 뒤이므로, 그보다 나중 as-of 의 비교는 수집도 로컬 읽기도 그 커밋 뒤다. digest 가 바뀌면(결속 주문이 늘거나, 스냅숏이 다시 쓰이거나,
  행이 바뀌면) 무장을 버리고 새로 무장한다. 재기동이면 결속 자체가 없다.
- C10 은 a191 잔여 위험 (a) 의 한 모양(장부는 종결인데 브로커 OPEN 에 남음 → 외부 주문으로 보임 `Compare` B21·B22)을 잡는다.

### D4 호출 자리 — 오늘의 자동 해제 자리 그대로(M3)

- `Observe` 깨끗한 갈래 B8(if@662) 안: 크레딧이 없어 `AwaitingAdjustment` 로 가기 **직전**에 `t.Terminal != nil` 이고 결속이 있으면 1차 판정을 묻는다. 1차
  판정의 장부 읽기는 `t.Terminal` 의 공개 읽기(tx 밖 — `j.db`)다. 결과가 「무장」이면 `arm` 을 갱신하고 `AwaitingAdjustment` 로, 「해제」면 블록에 해제
  후보 표시(얼린 비교 입력·digest)를 달아 `Cleared` 에 넣는다. 크레딧이 있는 블록은 오늘 경로 그대로 — 수량 규칙이 먼저(I2).
- `Terminal == nil` 또는 결속 없음 = 오늘과 같은 행동.
- blocking 갈래에서는 해제를 묻지 않는다(오늘 규칙과 같음).

### D5 해제 — 행 id 고정 + `released_at IS NULL` + 같은 tx 안 재판정(필수 조건 4)

- 새 journal 메서드 하나(새 파일 `internal/journal/reconcile_order_terminal_release.go`): 쓰기 트랜잭션 안에서 ① `rowID` 로 행을 읽어 활성·범위·cause 확인,
  ② 결속 신원마다 종결 스냅숏(**tx 를 받는** 기존 비공개 `lookupFillSnapshotScoped(ctx, tx, …)` — `queryRower` 인자라 재사용 가능), ③ 대체 흔적(대체 간선·AMEND 시도 — M15) 존재(쓰기 없는
  tx 읽기 — 공개 `ResolveCurrentOrderIDScoped` 는 `j.db` 를 쓰고 충돌 시 `EnterReconcile` 을 부르므로 금지, lineage.go:334·419~420·442~454), ④ `fill_events.id >
  워터마크` 존재(tx 읽기)를 읽어 **같은 판정 함수**(콜백 — 순수, `Observe` 가 얼린 비교 입력·결속·무장을 값으로 붙잡음)에 넘기고, 「해제」이고 digest 가 1차
  판정과 같을 때만 같은 tx 에서 `UPDATE reconcile_states SET released_at, release_cause = 'ORDER_TERMINAL_CONFIRMED', evidence = … WHERE id = ? AND released_at IS NULL`.
  `RowsAffected == 1` 을 확인한다.
- tx 안에서 공개 `Journal` 메서드 호출 금지(연결 1개 — journal.go:174 — 자기 대기), 콜백 안에서 `Tracker` 메서드 호출 금지(`Observe` 가 `t.mu` 를 쥔 채
  `persist` 를 부름 — mismatch.go:623·767). 시험·변이로 고정(m11·m12).
- tx 안 판정이 「유지」면 오류가 아니라 「해제 안 함」 — 블록·게이트 유지, 사유 기록, 주기 성공(M5).
- `persist` B6(range@899) 은 블록별로 갈린다: 크레딧 해제는 오늘의 `ReleaseReconcile`(call@900, `ADJUSTMENT_APPLIED`), 종결 증명 해제는 새 메서드. 새 메서드가
  「해제 안 함」이면 `persisted.released` 에 넣지 않는다(오늘 B8 if@911 의 「확인 못 함 = 오류」와 다름 — 그 갈래는 크레딧 경로에만 남음).
- `ValidReconcileReleaseCause` B1(switch@112) 에 새 상수 한 칸(상수 집합이 제약이라는 선례).
- 해제 쓰기 실패(tx 오류) 처분은 오늘과 같다: `persistErr` → 블록·게이트 유지 → `RunOnce` B7(if@457) → return@461(편입 없음).

### D6 재기동

결속·무장은 프로세스 기억이다. 재기동한 엔진은 결속이 없어 남은 missing-order 행을 이 경로로 풀지 않는다 — 운영자 해제(`reconcile-resolve`)로 돌아간다.
`Restore` 는 결속·무장을 지운다(같은 객체 재사용 대비). `Refresh` 는 무편집이다(결속을 만들지 않고, 지우지도 않으며, 같은 행인지는 C2 가 판정한다).

### D7 로그 이벤트(M6)

`RunOnce` 의 `Observe` 성공 뒤(call@453 다음) 새 파일의 함수 하나가 `outcome.Cleared` 중 새 cause 블록마다 `reconcile.block_released` 를 쓴다(로그만, 알림 없음).
필드 `symbol`·`market`·`release_cause`·`bound_orders`. 계좌번호·주문번호·수량·가격 없음. `logCycle` 무편집.
같은 함수가 결속이 남은 종목의 유지 사유가 **바뀔 때만** `reconcile.block_release_held`(필드 `symbol`·`reason`)를 씀 — M5 의 구현(2026-11-17 수용).
바뀜 기억은 `ReconcileDriver.terminalHoldLogged`(RunOnce 직렬이라 잠금 없음).

### D8 바꾸지 않는 것

차단을 거는 조건·시점(`blocksFor` 의 블록 집합·키·값·개수 — 비공개 필드 하나만 더 실음), 수량 불일치 해제 규칙, 영구 승격·영구 행·운영자 해제(P3),
`EnterReconcile`, `reconcile-resolve`, `Comparer.Compare`, `LocalStateFromJournal`, `Refresh`, 청산 확정 하한·exit 루프, a191 체결 감시, 진입 게이트,
**장부 스키마(36 그대로)**.

### D9 upstream·토글

`internal/reconcile`·`reconcile_states` 는 TossOS 계층이고 엔진을 켤 때만 돈다 — 토글 OFF = upstream 동일(불변식 3)에 닿지 않는다. 새 토글 없음.
`Tracker.Terminal == nil` 이 사실상의 비활성 상태다.

### D10 스키마·배포·롤백

스키마 변경 없음(36 유지). 마이그레이션·백업 복원 롤백 없음 — 롤백은 이전 이미지로 교체. 배포 순간 활성 missing-order 행은 결속이 없어 운영자 해제(N13).
새 release cause 값이 들어간 행을 이전 이진이 읽을 때: `ValidReconcileReleaseCause` 는 **쓰기** 검사라 읽기에는 영향이 없고, 해제된 행은 투영(`ActiveReconcileStates`
— `released_at IS NULL`)에 안 들어온다. 이전 이진의 broker-zero 는 그 행을 근거로 안 받는다(M7 과 같음).

### D11 편집 표면(2단계 — gate ⑤ FLM 번들 대상) — 결정 뒤 다시 셈

| 기존 함수 | 편집 | 결정 전 판본 대비 |
|---|---|---|
| `reconcile.Tracker.Observe` | B8 안 1차 판정·무장 호출(분기 1), 갈래 나누기 전 결속 추가·실격 호출 1, `persist` 뒤 같은 호출 1 | 같음(W2 쓰기 호출 → 기억 갱신 호출로 바뀜) |
| `reconcile.Tracker.persist` | missing-order 블록에서 워터마크 읽기 1 + `EnterReconcile` 결과로 결속 생성(call@880 결과 붙잡기 — 자리 1), 종결 증명 블록은 새 journal 메서드로 갈래 1 | 같음(결속 실어 보내기 → 결과 붙잡기) |
| `reconcile.blocksFor` | missing-order 블록 리터럴에 신원 비공개 필드 1 | 같음 |
| `journal.ValidReconcileReleaseCause` | case 값 1 | 같음 |
| `engine.buildGateway` | `Tracker{… Terminal: in.journal}` 필드 1 | 같음 |
| `engine.ReconcileDriver.RunOnce` | D7 호출 1줄 | 같음(M6 확정) |
| `reconcile.Tracker.Restore` | 잠금 안 결속·무장 지우기 대입 1(리뷰 3 #3) | **늘어남** |
| ~~`journal.Journal.EnterReconcile`~~ | **편집 없음** — 반환이 이미 삽입 증명 | **줄어듦** |

- 없어진 것: 새 표·`schemaV37`·`SchemaVersion` bump·`migrations` 줄, W2 첨부 쓰기 메서드, `bindingUnwritten`, `EnterReconcileRequest.Disputes` 필드.
- 새 파일: `internal/reconcile/order_terminal_release.go`(결속·무장 기억, 판정 함수, digest, 로그 함수), `internal/journal/reconcile_order_terminal_release.go`
  (`MaxFillEventID`, tx 해제 메서드와 tx 묶인 읽기 — 대체 흔적 존재·워터마크 뒤 체결 존재), 시험 파일 2.
- 타입 필드: `Tracker.Terminal`(인터페이스 — `MaxFillEventID`·`LookupFillScoped`·해제 메서드; production `*journal.Journal`), `Tracker.bindings`,
  `Block` 비공개 필드 2(신원·해제 후보). 상수: release cause `ORDER_TERMINAL_CONFIRMED`.
- 패키지 문서의 상태표 주석(mismatch.go:35~41)은 함수 밖. `BlockRules()` 데이터는 바꾸지 않음.

## 4. Q10 — 거부할 정상 입력과 통과시킬 위험 입력(결정 반영판)

**거부하는(블록을 유지하는) 정상 입력:**

| # | 정상 입력 | 처리 | 비용 |
|---|---|---|---|
| N1 | 같은 종목에 수량 불일치(또는 외부 포지션)가 같이 있었음(부분 체결 뒤 취소 포함) | C3 실격 → 조정 규칙만 | 오늘과 같음 |
| N2 | 같은 종목 주문 둘 이상, 하나만 종결 | 나머지는 missing → C9 → 다 종결되면 해제 | 그만큼 늦음 |
| N3 | 둘째 주문이 일시적으로 OPEN 에서 빠졌다 다시 보임(종결 아님) | `orders` 에 남음 → C6/C7 | 운영자(다시 보임은 「읽음」이지 원인 아님) |
| N4 | 비교 as-of ≤ 종결 기록 시각 | C8 → 다음 주기 | 1주기 |
| N5 | 다른 종목의 분쟁 때문에 비교 전체가 막힘 | D4(깨끗한 갈래만) | 그 분쟁이 끝날 때까지 |
| N6 | 종결이지만 `fail_closed` | C7 | 운영자 |
| N7 | 종결 스냅숏 읽기 실패·없음 | C6(주기 성공, M5) | 다음 주기 |
| N8 | a191 이 아직 종결을 못 씀 | missing → C9 | a191 이 확인할 때까지 |
| N9 | 정정 lineage 이동(a167 영역), 또는 그 번호를 겨눈 AMEND 시도·대체 간선이 하나라도 있음(실패한 정정 포함 — M15) | C5 | 운영자. 운영 장부 집계(2026-11-17, 개수만): AMEND 시도 0건 → 지금 막는 정상 입력 0 |
| N10 | 신원 요소가 빈 값·다른 계좌의 missing-order | C4 | 운영자 |
| N11 | 같은 번호의 외부 주문(브로커 OPEN 에 남음) | C10 | 운영자 |
| N12 | 처음 통과한 비교 | C13 무장만 | 1주기(≈60s) |
| **N13** | **재기동 뒤 남은 missing-order 행** | 결속 없음(C1) | 운영자 해제(`reconcile-resolve`) |
| **N14** | **이미 있던 활성 행에 겹친 missing-order**(투영 거절 행, 이전 프로세스 행, 다른 생산자 행) | `entered = false` → 결속 없음(C1) | 조정 규칙 또는 운영자 |
| **N15** | **삽입 증명이 없는 블록**(`EnterReconcile` 이 커밋 뒤 오류를 돌려 재시도가 B8 로 감, `Terminal == nil` 인 tracker) | 결속 없음(C1) | 운영자 |
| N16 | 차단 뒤 그 종목에 체결이 하나라도 커밋(결속 주문의 늦은 부분 체결, **다른 계좌**의 같은 종목 체결, 종목이 빈 legacy 체결 포함 — M16) | C12 | 운영자 또는 조정. 운영 장부 집계(2026-11-17): 체결 이벤트 중 종목이 빈 행 0건 |
| N17 | 영구 행이 있음 | 종목 행만 풀리고 계좌는 막힘(P3) | 운영자 |
| N18 | 종결 기록이 차단 삽입보다 **이전**(같은 주기 경합) | **거부하지 않음** — C8 은 비교와 종결의 순서만 봄 | — |
| N19 | 해제 사이에 `Resolve`·운영자가 행을 풀고 같은 종목에 새 행이 생김 | C2(행 id 다름) | — |
| N20 | 첫 삽입 시도가 실패해 pending 으로 남았다가 재시도로 삽입된 블록(리뷰 3 #1) | 결속 없음 | 운영자 |
| N21 | 같은 `Tracker` 객체에 `Restore` 가 다시 불림(리뷰 3 #3) | 결속·무장 지움 | 운영자 |

**거부하지 않을 때 통과하는 위험 입력:**

| # | 위험 입력 | 결과 | 안전망 |
|---|---|---|---|
| R1 | 브로커가 살아 있는 주문을 CLOSED 에 깨끗한 종결로 싣고 OPEN 에서도 빠뜨림(a191 잔여 (a)) | 해제 → 진입 재개 | 없음 — a191 만 있을 때 차단이 안 생긴 경우와 같은 계열. OPEN 에 보이면 C10 |
| R2 | 단건·CLOSED 가 둘 다 체결 0 으로 낡았는데 실제는 부분 체결(a191 잔여 (c)) | 보유 반영이 늦으면 우연한 일치 → 해제 → 반영되면 새 차단 | 보유 대사(지연 [미실측]) |
| R3 | 같은 정규 범위 안 번호 재사용(a191 잔여 (b)) | 모델 밖 | 없음 |
| R4 | 행 뒤의 투영 거절 | C12 가 닫음(거절된 delta 도 `fill_events` 에 남음). delta 0 거절은 없음 | — |
| R6 | 판정 읽기와 해제 사이의 체결·거절·`fail_closed` 갱신 | D5 tx 재판정 + digest 일치로 닫음 | — |
| R7 | 종결 시각 캡처와 커밋 사이에 수집이 시작된 비교 | C13 무장-후-해제로 닫음 | — |

(R5 는 (d) 결정으로 소멸 — 장부 결속 쓰기가 없다.)

## 5. 시험·변이 계획(결정 반영판)

픽스처는 합성 값만(`SYN001`·`SYNX`, `o-a192-…`, 수량 1~10, 가격 없음). 실 journal(`openJournal` 선례) + 가짜 시계. 새 시험 파일
`internal/reconcile/a192_order_terminal_release_test.go`, `internal/journal/a192_order_terminal_release_test.go`. 기존 시험 무편집.

RED(★ = base 에서 실패해야 함):

- ★ `TestA192MissingOrderBlockArmsThenReleasesAfterTheTerminalProof` — missing 관측(삽입) → `RecordFill` 로 종결 → 나중 깨끗한 비교 1(무장, 유지) → 더 나중 깨끗한
  비교 2(해제) → 장부 `ORDER_TERMINAL_CONFIRMED`, 게이트 열림.
- `TestA192ARestartedTrackerNeverReleasesThisWay`(N13 — 새 `Tracker` + `Restore`).
- `TestA192ARowAnotherProducerInsertedFirstIsNeverBound`(N14·필수 조건 1 — 투영 거절 행을 `Refresh` 와 `Observe` 사이에 삽입).
- `TestA192AnOverlappingMissingOrderOnAnExistingRowIsNeverBound`(N14 — 기존 행 위의 missing-order).
- `TestA192AnUnprovenInsertIsNeverBound`(N15 — 가짜 store 가 커밋 뒤 오류).
- `TestA192AComparisonNotAfterTheTerminalCommitKeepsTheBlock`(C8), `TestA192ANonCleanTerminalKeepsTheBlock`(C6·C7),
  `TestA192AQuantityDisputeDisqualifiesTheBinding`(C3 — 흔적 뒤 크레딧이면 `ADJUSTMENT_APPLIED`), `TestA192EveryBoundOrderMustBeTerminal`(N2·N3),
  `TestA192AForeignAccountObservationNeverAuthorisesRelease`(C4), `TestA192AnAmendmentAfterBindingKeepsTheBlock`(C5),
  `TestA192AnOpenOrderWithTheSameIDKeepsTheBlock`(C10), `TestA192AFillAfterTheBlockKeepsIt`(C12 — 워터마크 읽기와 삽입 사이에 체결 끼워 넣기 포함),
  `TestA192ABindingChangeRearms`(C13 — 무장 뒤 결속 주문 추가·스냅숏 재기록), `TestA192AFillCommittedBetweenVerdictAndReleaseKeepsTheBlock`(D5),
  `TestA192AnAuthorityErrorStillRecordsTheTaint`(D2 — 영구 보류 권위 읽기 실패 주기의 수량 흔적, **외부 포지션만 있는 깨끗한 갈래 비교 포함** — 리뷰 3 #2),
  `TestA192ARetriedInsertIsNeverBound`(N20 — 첫 시도 실패 → 둘째 주문·수량 흔적 → 재시도 삽입 → 두 깨끗한 비교에도 유지), `TestA192ASameObjectRestoreDropsBindings`(N21), `TestA192APermanentBlockStaysOperatorOnly`(N17),
  `TestA192WithoutATerminalReaderTheTrackerBehavesAsToday`, `TestA192TheReleaseLogCarriesNoAccountOrderQuantityOrPrice`.
- **판정을 켠 상태**에서 기존 안전 시나리오 재실행: `TestACoincidentalAgreementStillDoesNotRelease`·a083 크레딧·a110 승격 흐름을 `Terminal` 이 실 journal 인 tracker 로
  (리뷰 1 #7).
- 기존 시험(선택 실행, 2단계): `./internal/reconcile/...`, `internal/journal` 의 `reconcile_states_test`·`risk_bucket_owner_test`, `internal/app/engine` 의
  `reconcileloop_test`·`engine_test`, `cmd/tossctl` 의 `engine_reconcile_test`. 상한 `codegraph affected --filter '*_test.go' -d 1` 24파일.

변이 — 각 변이에 **양성 대조군**(그 변이가 대상 코드에 닿는지 먼저 확인하는 장치)을 둔다(기억 「뮤테이션은 대상에 닿아야 한다」):

| 변이 | 바꾸는 것 | 잡을 시험 | 양성 대조군(닿음 확인) |
|---|---|---|---|
| m1 | **삽입 증명을 건너뜀**(`entered` 무시, 결속을 늘 생성) | `…AnotherProducerInsertedFirst…`·`…OverlappingMissingOrder…` | 그 시험의 `EnterReconcile` 이 `entered=false` 를 돌려줬음을 가짜 store 계수로 단언(무변이에서 1) |
| m2 | **재기동 뒤에도 결속이 있다고 침**(`Restore` 뒤 행마다 결속 생성) | `…ARestartedTrackerNeverReleases…` | 재기동 전 같은 흐름이 해제됨(무변이 대조 — 판정 경로가 산다) |
| m3 | **워터마크 비교를 뒤집음**(`>` → `<=`) | `…AFillAfterTheBlockKeepsIt…` | 끼워 넣은 체결의 id > 워터마크임을 시험이 단언 |
| m4 | **무장 없이 첫 통과에서 해제** | `…ArmsThenReleases…`(비교 1 뒤 유지 단언) | 비교 1 에서 판정이 「통과」까지 갔음을 사유 코드 `armed` 로 단언 |
| m5 | 무장 digest 무효화 삭제 | `…ABindingChangeRearms…` | 무변이에서 재무장 사유가 찍힘 |
| m6 | C3 삭제 | `…AQuantityDisputeDisqualifies…` | 흔적이 실제로 세워졌음을 결속 상태로 단언 |
| m7 | C5 삭제(lineage) | `…AnAmendmentAfterBinding…` | 대체 흔적 읽기가 참을 돌려줌 |
| m8 | C7 `!FailClosed` 삭제 | `…ANonCleanTerminal…`(fail_closed 행) | 그 행의 `terminal=1 ∧ fail_closed=1` 을 장부에서 단언 |
| m9 | C8 엄격 → 이상 | `…NotAfterTheTerminalCommit…`(같은 시각) | 같은 시각 입력임을 단언 |
| m10 | C10 삭제 | `…OpenOrderWithTheSameID…` | `ExternalOrd` 에 그 번호가 있음을 단언 |
| m11 | 해제 tx 안에서 공개 `LookupFillScoped`(`j.db`) 호출 | `…FillCommittedBetweenVerdictAndRelease…`(시한 안 끝남) | 무변이가 시한 안에 끝남 |
| m12 | tx 재판정 생략(1차 판정만 믿음) | 같은 시험 | 끼워 넣은 체결이 커밋됐음을 단언 |
| m13 | 「전부」 → 「하나라도」 | `…EveryBoundOrder…` | 결속 주문 2개임을 단언 |
| m14 | `orders` 0 개일 때 참(빈 표본) | `…AnUnprovenInsert…` 변형 | 결속이 0 신원으로 만들어졌음을 단언 |
| m15 | 해제 cause 를 `ADJUSTMENT_APPLIED` 로 | `…ArmsThenReleases…`(cause 단언) | — |
| m16 | 판정을 크레딧 판정 앞으로 | `…AQuantityDisputeDisqualifies…`(크레딧 경로가 이겨야 함) | 크레딧이 있음을 단언 |
| m17 | `Terminal == nil` 에서도 판정 | `…WithoutATerminalReader…` | — |
| m18 | 로그에 주문번호 | `…TheReleaseLogCarries…` | 로그 줄이 1개 찍혔음을 단언 |
| m19 | 갈래 나누기 전의 결속 갱신 호출 삭제(조기 반환 뒤로 밀림) | `…AnAuthorityErrorStillRecordsTheTaint…` | 그 주기가 권위 읽기 오류로 끝났음을 단언 |
| m20 | 재시도 삽입도 결속 | `…ARetriedInsertIsNeverBound…` | 첫 시도가 실패했음을 가짜 store 계수로 단언 |
| m21 | `Restore` 의 결속 지우기 삭제 | `…ASameObjectRestoreDropsBindings…` | `Restore` 전 결속이 있었음을 단언 |

사본 하네스는 무변이 대조군이 GREEN 이고 시작 sha 가 맞을 때만 변이를 돈다.

## 6. Manager 결정(2026-11-17 확정)과 남은 결정

| # | 결정 | 확정 |
|---|---|---|
| M1·M13 | 결속 방식 | **(d) 프로세스 기억**. 스키마 36 유지, 결속 표·`evidence` 봉투 없음. I3 개정(Q7) |
| M2 | 영구 차단 | **P3** — 운영자 전용 그대로. P1'·P2 별도 Story. Q8 반증 ① 수용 |
| M3 | 해제 자리 | 깨끗한 갈래에서만 |
| M4 | C10 | 포함 |
| M5 | 판정 읽기 실패 | 유지 + 사유, 주기 성공 |
| M6 | 로그 | `RunOnce` 뒤 새 함수, 로그만 |
| M7 | broker-zero | 새 cause 를 근거로 받지 않음(fail-closed) |
| M8·M9 | 배포 롤백·R5 | (d) 로 **소멸** |
| M10 | Feature·이름 | 그대로(FEAT-TOS-005) |
| M11 | 행 뒤 투영 거절 | **(ii') 체결 id 워터마크**(삽입 호출 전 읽기 허용), 판정은 해제 tx 안 |
| M15 | C5 의 넓이 | **수용(2026-11-17)** — 그 번호를 겨눈 AMEND 시도나 대체 간선이 하나라도 있으면 유지(범위·확정 무관). 유지 쪽 초집합, lineage 규칙 복제 없음. 운영 AMEND 0건 |
| M16 | C12 의 넓이 | **수용(2026-11-17)** — 그 종목(계좌 무관) 또는 종목이 빈 체결이면 유지. 운영 체결 이벤트 중 종목 빈 행 0건 |
| M5 구현 | 유지 사유 기록 | **수용(2026-11-17)** — 결속마다 마지막 판정 사유(tx 재판정은 `tx_` 접두), 사유가 바뀔 때만 로그 `reconcile.block_release_held`(종목·사유만) |
| M12 | 커밋 시각 순서 | **(iii) 무장-후-해제**, digest 무효화 |

**남은 결정 M14**(리뷰 3 #1): 첫 시도에 삽입되지 않은(재시도로 삽입된) 블록 — 지금 문서는 「결속 안 함」(보수). 대안은 pending 동안의 신원·흔적을
따로 쌓아 삽입 증명 뒤 넘기기. 리뷰 3회차 반영 4건은 재리뷰 없이 문서에 들어갔다 — `review.md`.

## 7. 후속 후보(이번 change 밖) — 보호를 직접 줄이는 자리를 앞에

1. **청산 확정 하한의 적용 범위**(Q5 ③) — 블록이 덮는 동안 손절은 왕복 2회를 더 하고 읽기가 실패하면 0 주다. missing-order 블록(수량 분쟁 아님)에도 그 하한이
   필요한지. exit 경로 변경이라 별도 High-risk.
2. **편입 건너뛰기 가시성**(`judgeHoldings` B10) — 차단으로 건너뛴 무관리 보유분을 사실 칸으로 보고.
3. P1' — 계좌 영구 차단의 자동 해제(승격 시점 문턱 키 전부 결속, 「전부 missing-order ∧ 전부 깨끗한 종결 ∧ 그 뒤 계좌 전체 비교 차이 0」). spec 반전.
4. P2 — 엔진을 세우지 않는 운영자 해제.
5. `Tracker.ExitAllowed`(production 호출자 0) 장식 메서드 정리.
6. `reconcile.clean` 이벤트의 의미(a191 후속).
7. 정정 옛 번호(C5) — a167 과 함께.
8. 영구 행 `Detail` 이 문턱 키 하나만 적는 것(Q8 ①).

## 8. Manager 가 손으로 읽은 사실의 검증

| 사실 | 판정 | 근거 |
|---|---|---|
| `blocksFor` 1402~1433, missing-order 마다 종목 블록, cause `QUANTITY_MISMATCH`, release `adjusted_reconcile`, 수량 블록과 원인 같음 | **맞음** + 보충: `Market` 도 싣지만 `persist` 가 장부에 안 넘김 → 행 `scope_market = NULL`. 같은 종목이면 블록 키 하나 | [AST] B2 range@1418, [손] 1419~1430·880~885 |
| `DefaultReconcileInterval = 30s` | 값은 맞음. **운영 주기는 아님** — `Due` 호출자 0, 실제 주기는 `DefaultReconcilePeriod = 60s` + 안정화 2s | [손] mismatch.go:281, reconcileloop.go:95 |
| `DefaultMaxFailures = 3`, 승격 키 6요소, streaks 프로세스 로컬, `permanentPromotionDetail`, `Resolve` 운영자 해제 | 맞음 + 보충: 문턱 키는 하나만 기록되고 `kind` 가 첫 비교 원소라 missing-order 가 수량보다 먼저 뽑힘(Q8 ①) | [AST] `thresholdKey`·`promotionDisputeKeyLess` |
| `ExitAllowed` 는 항상 true | 문자 그대로 맞음. **그러나 production 호출자 0** 이고 실제 청산 경로는 확정 하한(Q5 ③)으로 블록의 영향을 받음 — 「청산·손절이 블록에 안 막힌다」는 반만 참 | grep, [AST] `ConfirmedFloor` B2·`applyFloor` B2 |
| 해제 크레딧 `adjusted map[symbol]asOf`, 수렴은 수량 차이가 없으면 크레딧 없음 | 맞음 + 보충: 같은 종목에 **별도로** 수량 수렴이 있으면 오늘도 missing-order 블록이 `ADJUSTMENT_APPLIED` 로 풀릴 수 있음(Q2 단서) | [AST] `ConvergeQuantities` B1, `hasBlockFor` B2 |
| `reconcile_states` 열 목록, 주문 번호 구조화 열 없음(evidence 에만) | 맞음 + 보충: evidence 에는 **첫 주문 하나만**, 그리고 같은 cause 를 세 생산자가 씀 | [AST] `EnterReconcile` B8, `enterReconcileScopeInTx` B3 |
| journal `user_version` 36 | 코드 `SchemaVersion = 36` 은 맞음. **운영 장부 값은 확인 안 함**(지시상 운영 디렉터리 금지) | [손] schema.go:6 |
| `judgeHoldings` 가 `blocked` 면 continue, 로그·경보 없음 | 맞음 + 보충: `ExitEligible`(관리 중) 검사가 먼저라 관리 중 포지션은 영향 없음, `Covers` 는 시장을 안 봄, 영구면 모든 종목 | [AST] B7 if@140, B10 if@152, `Covers` B4·B5 |
| a191 `closedLookupInterval = 30s`, 주기당 1건 | 상수 맞음 [손] closed_lookup.go:82. 수용 조건은 a191 design D3 인용(이번에 AST 재생성 안 함 — 안 봄) | — |
