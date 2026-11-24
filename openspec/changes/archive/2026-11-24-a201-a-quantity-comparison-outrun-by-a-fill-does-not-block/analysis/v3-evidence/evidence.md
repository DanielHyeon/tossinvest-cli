# a201 v3(해제 쪽) 증거 조사

- 시작 HEAD `7fec501f17566a2be7049d3a2510381968635d64`(main) · 끝 HEAD 같음. 메인 저장소 `git status --short` 0줄(시작·끝).
- a201 브랜치 `388f2023c`: production Go 는 main 과 동일(`git diff --stat 7fec501f1 388f2023c -- '*.go'` = 재현 시험 1파일만). 줄 번호는 main 기준이 a201 브랜치에도 그대로.
- AST: `go build ./tools/logic-map` → `ast/*.ast.json` 20개, 요약 `ast-summary.txt`(`sum.py`). RunOnce · Observe · persist · Refresh · orderTerminalCandidate · ConvergeQuantities 는 a201 `analysis/cited-ast/` 와 **바이트 동일**(`cmp`).
- 좌표: `B<n>(kind@줄)` · `call@줄` · `return@줄` = AST. **[손]** = AST 가 열거하지 않는 리터럴·SQL·구조체·주석을 손으로 읽음.
- CodeGraph **1.6.0**: `callers` 10개 심볼(`codegraph-callers.txt` — 이름 기반이라 `Observe` 는 잡음이 많아 production 호출 자리는 grep 으로 확정), `affected reconcileloop.go mismatch.go order_terminal_release.go --filter '*_test.go' -d 1` = 19 파일(`codegraph-affected.txt`).
- 탐침: scratch 사본 worktree(`388f2023c`, 끝에 제거) — `probe.diff` · 실행 원문 `probe-run.txt` · `probe-r11-keep.txt` · `probe-r11-offset.txt`.

## Q1. 표지를 만들 수 있는 자리와 그때 손에 있는 것

### (a) RunOnce → Observe 통로

| 사실 | 근거 |
|---|---|
| 「로컬 읽기 뒤 체결」의 두 재료를 동시에 쥘 수 있는 곳은 `RunOnce` 뿐 — 로컬 읽기 call@434 의 앞뒤, 수렴 보고 `converged`(call@451, 같은 함수 지역 변수), `diff.AsOf` | RunOnce AST; `d.opts.Journal` 은 구체 `*journal.Journal`(reconcileloop.go:120 [손]) 이라 `MaxFillEventID`(reconcile_order_terminal_release.go:75) · `FillWatermark`(position_adjustments.go:190) 둘 다 바로 부를 수 있음 |
| `Observe(ctx, diff)` 인자는 둘뿐, `Diff` 구조체(compare.go:292-304 [손])에 표지 칸 없음 | Observe 서명 · Diff 리터럴 |
| `Tracker` 에 RunOnce(또는 그 밖)가 종목별 프로세스 기억을 넣는 공개 메서드는 `AdjustmentApplied(comparison, symbols...)`(mismatch.go:484) 하나 — 그 생산 호출자는 `ConvergeQuantities` 의 defer(converge.go:189) 하나 | grep 비시험 호출 1 · CodeGraph |
| `Refresh` 는 장부 활성 행으로 `t.blocks` 를 다시 만들 뿐 표지를 받을 통로 아님 | Refresh AST B6 range@1048 · B9 range@1057 |
| **결론**: 표지 통로는 지금 없음. 최소 추가 = (i) `RunOnce` 편집(로컬 읽기 전 `MaxFillEventID` 1회, 수렴 뒤 수량 불일치 종목마다 `FillWatermark`, `Converged` 대조, 표지 호출 1) + (ii) `Tracker` 의 새 메서드 1(새 파일이면 기존 함수 편집 아님) + (iii) 표지를 읽는 쪽(Q2). 탐침 `probe.diff` 가 이 모양(RunOnce +25줄, 새 파일 `probe_v3.go`) | 탐침 |
| 참고(사실만): 기존 통로 `AdjustmentApplied` 를 RunOnce 가 Observe **뒤**에 같은 `diff.AsOf` 로 부르면, 그 크레딧은 같은 비교라 이번 주기에 쓰이지도 반박되지도 않고(`creditUsableBy` 엄격히 나중, mismatch.go:509-519) 다음 깨끗한 비교에서 `ADJUSTMENT_APPLIED` 로 풂 — Tracker 무편집. 단 원인 기록이 「조정 없음」인데 `ADJUSTMENT_APPLIED`(정본 :77 「구분되는 release cause」), broker-zero 경로는 `ADJUSTMENT_APPLIED` 행에 값 0 조정 기록을 요구해 거절(risk_bucket_owner.go:297-309 · :684-) | 정적 추론 — **이 형태로는 탐침 안 함**(탐침의 표지 수명 KEEP=0 과 같은 기전) |

### (b) a192 결속 생성 자리 · 조건 · 모양

| 사실 | 근거 |
|---|---|
| 자리: `persist` B2 range@890 → `enterReconcileWithBinding` call@902 → 그 안 `MaxFillEventID` call@351(삽입 **전**) → `EnterReconcile` call@358 → B3 if@359(`err == nil && entered && bindable && state.ID != ""`) 일 때만 결속 생성 | enterReconcileWithBinding AST |
| `bindable = t.Terminal != nil && block.missing != nil && !t.terminal.seen[block.Key()]`(:348 [손]) — `block.missing` 은 `blocksFor` 의 missing-order 리터럴에만 실림(mismatch.go:1468 [손]). **수량 블록은 결속 대상 아님** | blocksFor AST B2 range@1455 |
| `seen` = 이번 Observe 시작 때 이미 있던 블록 키(beginOrderTerminalObservation :305-308) → 재시도 삽입·기존 키 겹침은 결속 안 함 | beginOrderTerminalObservation |
| `terminal.bindings map[string]*terminalBinding` 키 = **종목**(정규화). 칸: rowID · symbol · orders(map[boundOrder]) · tainted · fillWatermark · armed · armAsOf · armDigest · reason(:56-67 [손]) | 구조체 |
| 수량 블록용으로 같은 맵에 넣으면: `updateOrderTerminalBindings` 가 같은 종목이 `diff.Quantities` 에 있을 때 `tainted = true`(B6 range@331 → B7 if@332) — 표지가 서는 주기 자체가 수량 불일치라 즉시 실격; 판정 C3(B4 if@189) · C4 빈 orders(B5 if@194 → `identity_incomplete`)에서 영원히 유지. **판정·갱신 함수는 missing-order 전용**. 모양 중 재사용 가능한 칸은 rowID · fillWatermark · arm 계열뿐 | updateOrderTerminalBindings · orderTerminalReleaseVerdict AST |
| 같은 종목의 QM 블록과 missing-order 블록은 **한 키**(`Block.Key` 종목 범위 = `symbol|계좌|종목`, mismatch.go:256-257 [손]); `blocksFor` 가 수량을 먼저 만들어 같은 비교에 둘 다 있으면 QM 이 이김(Observe B17 range@738 → B18 if@739). W2 재현의 그 주기 diff 에는 QM(31 vs 0)과 missing-order(o-stop)가 같이 있음 → 블록은 QM, 결속 없음 | Observe AST, 재현 |

### (c) 이미 활성인 수량 블록(entered=false 모양)

| 사실 | 근거 |
|---|---|
| 장부에 이미 활성 행이 있으면 그 블록은 `Refresh`(call@458)가 `t.blocks` 에 넣어 둠 → Observe blocking 갈래 B17 range@738 → B18 if@739 `!exists` 거짓 → **`out.Added` 에 안 들어감** → `pendingBlocksForPersistence`(pending 만, mismatch.go:1140-1156)에도 없음 → `persist` · `enterReconcileWithBinding` · `EnterReconcile` 가 **그 종목에 대해 불리지 않음** | Observe · persist AST |
| 따라서 「앞 크레딧 위 W2」·R1-1 모양(QM 이 이미 활성)의 표지는 `enterReconcileWithBinding` 에서 만들 수 없음 — 다른 자리(Observe 안, 또는 Observe 뒤 Tracker 메서드)가 필요. 탐침은 Observe 뒤 새 메서드가 `hasBlockFor(t.blocks, 종목)`(QM ∧ adjusted_reconcile, mismatch.go:556-567)일 때만 표지 | 탐침 |
| `Block` 에 행 id 칸이 없음(mismatch.go:222-247 [손]); `blockFromReconcileState` 는 ID 를 옮기지 않음(:1398-1424 [손]). 행 id 에 고정하려면 `ActiveReconcileStates`(ID 포함) 재읽기 또는 `Block`·`blockFromReconcileState` 편집 필요 | 구조체 |
| a192 는 entered=false 를 의도적으로 제외(N14 — 투영 거절 행·이전 프로세스 행·다른 생산자 행). QM cause 생산자: 추적기 수량 · 추적기 missing-order · 투영 거절(position_projection.go:401-415, 활성 행 있으면 `return nil`) · **위험 버킷 orphan fill**(`latchReleasedOwnerLateFillInTx` risk_bucket_owner.go:595, `scope_market` 채움 — a192 Q1 은 셋으로 셈, 생산 도달성 미확인) | grep `INSERT INTO reconcile_states` |

## Q2. 해제 판정을 걸 자리

| 사실 | 근거 |
|---|---|
| 자리: Observe B2 if@650(`!BlocksEntry`) → B6 range@666 → B7 if@667(cause QM ∧ release adjusted_reconcile 아니면 건너뜀 — 영구 블록은 operator_only 라 여기서 빠짐) → **B8 if@675 `!creditUsableBy(t.adjusted[종목], AsOf)`** → B9 if@682 `orderTerminalCandidate` call@682 → 거짓이면 `AwaitingAdjustment` append@686. 크레딧 있으면 B10 if@689(disputed → Awaiting@698) 아니면 Cleared@701 | Observe AST(사전 증거와 같은 좌표) |
| `orderTerminalCandidate` 는 결속이 없으면 B2 if@380 → return@381(거짓). 표지된 QM 블록은 결속이 없으므로 **a192 후보 함수 안에서 끝내려면 그 함수를 고쳐야 함**(B2 앞/뒤). a192 이 새 파일로 만든 함수라 a192 FLM 번들 없음 — 이번에 고치면 기존 함수 편집(gate ⑤ 번들 대상) | orderTerminalCandidate AST, a192 function-logic 목록(7개: buildGateway · RunOnce · ValidReconcileReleaseCause · blocksFor · Observe · persist · Restore) |
| 둘 중 하나는 반드시 편집: Observe(B9 뒤에 둘째 후보 — 탐침이 이 자리) 또는 orderTerminalCandidate | — |
| **해제 원인 기록 때문에 persist 도 편집 필요**: persist B6 range@921 → B7 if@922(`terminalRelease != nil` → `releaseOnOrderTerminal` call@925) 아니면 `ReleaseReconcile` call@935 의 Cause 가 리터럴 `ADJUSTMENT_APPLIED`(:938 [손]). `RECHECK_MATCHED` 를 쓰는 갈래 없음. `releaseOnOrderTerminal` → 장부 `ReleaseReconcileOnOrderTerminalProof` 는 원인을 `ORDER_TERMINAL_CONFIRMED` 로 박음(reconcile_order_terminal_release.go:106 · :128 [손]) | persist · releaseOnOrderTerminal AST |
| a192 재판정이 하는 일: 쓰기 tx 안에서 `readOrderTerminalProof` 를 다시(행 id 로 행 · 결속 주문 스냅숏 · 대체 흔적 · `fill_events.id > 워터마크 ∧ (종목 같음 ∨ 종목 빈 행)`, 계좌 무관 :188-198 [손]) → 순수 콜백 `orderTerminalReleaseVerdict` → 참일 때만 `UPDATE … WHERE id=? AND released_at IS NULL`, `RowsAffected == 1` 확인. 콜백은 잠금·공개 Journal 메서드 금지(연결 1개) | journal 파일 [손], releaseOnOrderTerminal AST call@416 · call@418 |
| 표지된 QM 블록이 재판정에서 다시 볼 것(사실상 남는 것): 행이 그 id·활성·QM·종목 범위인가(C2 — 행 id 필요, Q1(c)), 표지 뒤 그 종목 체결 유무(C12 유사 — 워터마크를 **표지 때** 읽어야 함, Q3). 주문 스냅숏 · lineage 는 대상 없음. 기존 `ReleaseReconcile` 은 범위(계좌·종목·시장)+`ExpectCause` 로 활성 행을 tx 안에서 다시 찾을 뿐 행 id·워터마크를 안 봄(reconcile_states.go:276-310 [손]) | — |

## Q3. a192 C1~C13 을 표지된 수량 블록에 옮기면

| # | a192 조건 | 표지된 QM 블록 | 이유(사실) |
|---|---|---|---|
| C1 | 이 프로세스 삽입 증명 | **(iii) 충돌** | 삽입 증명을 요구하면 처음 선 W2·W3 만 닫히고 「앞 크레딧 위 W2」(QM 이미 활성 — Q1(c))는 표지 대상이 못 됨. 그 모양을 닫으려면 기존 행(출처 불명 — 투영 거절·이전 프로세스·orphan fill 포함)에 표지해야 함 |
| C2 | 활성 행 id = rowID, QM, scope_market 빈 값, 계좌 같음 | (i) | 단 entered=false 행은 id 를 다시 읽어야 함(Block 에 ID 없음). ScopeMarket 있는 행은 지금도 `ReleaseReconcile`(시장 빈 값)로 안 풀림 |
| C3 | 수량·외부 포지션 흔적이면 실격 | **(iii) 뒤집힘** | 표지 자체가 수량 불일치 주기에 섬. 대응물은 「표지보다 나중 비교가 그 종목을 분쟁으로 봄 → 표지 소멸」(크레딧 반박과 같은 모양, mismatch.go:841). 탐침 KEEP=0(소멸)·KEEP=1(유지) 둘 다 R1-1 결과 같음(아래 Q4) |
| C4 | 신원 완전·계좌 같음 | (ii) 주문 신원 없음 / (i) 계좌 | `diff.AccountRef` = 추적기 계좌 조건은 적용 가능 |
| C5 | lineage 이동 없음 | (ii) | 주문 대상 아님 |
| C6·C7 | 종결 스냅숏 있음·깨끗한 종결 | (ii), 단 열린 질문 | 같은 주기 diff 에 그 종목 missing-order 가 같이 있음(W2 재현). 다음 비교가 깨끗하다 = 그 주문이 추적 집합에서 빠짐 = `terminal = 1`(fail_closed 포함 — `TrackedFillOrders` 조건, codex r1 :120) |
| C8 | as-of > 종결 커밋 | (i) 변형 | 「표지한 비교 as-of 보다 엄격히 나중」— `creditUsableBy(mark, diff.AsOf)` 그대로 씀(탐침). 주기 직렬 · 수집 뒤 로컬 읽기 · 체결 커밋은 표지 계산(워터마크 읽기) 전 → 다음 비교의 로컬 읽기는 그 체결을 반영 |
| C9 | 그 종목 분쟁 없음 | (i) | `symbolsInDispute` = 수량 ∪ missing ∪ **외부 포지션**(mismatch.go:576-595). 외부 **주문**은 안 셈 |
| C10 | 같은 번호 외부 주문 | (ii) | — |
| C11 | 비교 전체가 진입 안 막음 | (i) | 호출 자리(B2 if@650)가 보장. 다른 종목 분쟁·missing 이 있으면 해제 없음(a192 N5 · 크레딧과 같음) |
| C12 | 워터마크 뒤 그 종목 체결이면 유지 | (i) 조건부 | 워터마크를 **로컬 읽기 전 전역값**(v2 D2')으로 쓰면 표지를 만든 그 체결이 「뒤」로 세어져 언제나 실격. 표지 때(수렴·Observe 뒤) 읽어야 함. 그 사이 체결은 흡수(다음 로컬 읽기에 반영). 활성 행이 있을 때 투영 거절은 새 행 없이 `return nil`(position_projection.go:401-405) — a192 M11 의 가림이 표지 뒤에도 같은 모양으로 있음(탐침 안 함) |
| C13 | 무장-후-해제 | (ii) 구조상 불필요 / 결정 | a192 무장은 `RecordFill` committed_at 이 tx 시작 전 시각이라서(order_terminal_release.go:20-25). 표지는 id(커밋 순서)와 메모리 보고로 서고 그 뒤 비교는 직렬로 나중 — 탐침은 첫 나중 깨끗한 비교에서 해제 |

추가 실격 조건(질문 목록 그대로):

| 경우 | 지금 코드에서 일어나는 일 |
|---|---|
| 표지 뒤 **또** 체결 | 다음 로컬 읽기에 반영. 투영 거절이면 가려짐(C12 유사로만 막힘) |
| 그 종목에 조정 반영 | B8 크레딧 검사가 먼저 → `ADJUSTMENT_APPLIED`(탐침 R1-1 k+2) |
| 외부 보유 | 분쟁(C9) → 유지. 외부 주문만은 분쟁 아님 |
| 같은 종목 missing-order 동반 | `BlocksEntry` 참 → blocking 갈래 → 해제 없음 |
| 그 종목에 a192 결속도 있음(missing-order 로 삽입된 행에 나중 수량 불일치) | 결속은 C3 로 실격(`quantity_dispute_seen`)되지만 행은 같은 키·같은 행 — 표지가 서면 v3 가 그 행을 풀 수 있음. a192 정본 :381 은 「이 경로로」만 금함 — 관계 미정 |
| 계좌 범위·영구 | B7 if@667 에서 operator_only 건너뜀 → 자동 해제 없음. 종목 행은 풀릴 수 있음(a192 N17 과 같음) |
| 재기동 | 표지 = 프로세스 기억이면 소실(크레딧 `Restore` :1001 · 결속 :1003 선례) |
| 「새 스냅숏」의 정의 | `diff.AsOf` 가 표지 비교 as-of 보다 엄격히 나중(`creditUsableBy` — 크레딧·a192 C8 과 같은 시각 비교). 읽을 수 없으면 거짓(유지) |

## Q4. R1-1 · R2-1 · R2-2 와 v3

**탐침 원문**(`probe-run.txt`, `probe-r11-keep.txt`, `probe-r11-offset.txt`). 탐침 = Observe 깨끗한 갈래 B9 뒤에 「표지 ∧ 나중 as-of ∧ 비분쟁 → Cleared(RECHECK)」, persist 가 그 블록을 `RECHECK_MATCHED` 로, RunOnce 가 Observe 뒤에 표지(로컬 읽기 전 `MaxFillEventID` · 종목 `FillWatermark` · `Converged` 대조), 표지 수명 = 나중 분쟁이면 소멸(`PROBE_V3_KEEP=1` 이면 유지)·해제·블록 없음. 행 id 고정·tx 재판정·C12 유사·Restore 는 흉내 안 냄.

| 모양 | 결과 |
|---|---|
| W2 | 표지 주기 QM 활성·진입 거절 → 다음 깨끗한 주기 `released:RECHECK_MATCHED` → 재매수 편입 1 |
| W3-a · W3-b(합성) | 같음(다음 주기 RECHECK) |
| 앞 크레딧 위 W2 | 표지 주기 해제 0(크레딧 반박) → 다음 깨끗한 주기 RECHECK |
| **R1-1**(앞 주기 31→20 수렴·크레딧·QM → 사람 5 주 매도·스냅숏 15·로컬 읽기 20 → 엔진 매도 5 주가 W2 자리에 커밋 → 수량 stale) | 주기 k: `expected quantity 20 but found 15` stale, 해제 0, 표지 `{SYN201: k as-of}`. 주기 k+1(스냅숏 10·로컬 15): **blocking → 해제 0**, 수렴 15→10. KEEP=0 이면 표지 소멸, KEEP=1 이면 남음. 주기 k+2(10·10): 두 경우 모두 `ADJUSTMENT_APPLIED`(크레딧 검사 B8 이 먼저) |
| R1-1 변형(k+1 에 사람이 5 주 되사 스냅숏 15 = 로컬 15) | 주기 k+1 깨끗 → **`RECHECK_MATCHED` 해제**. 오늘 코드는 크레딧이 k 에 반박돼 운영자 해제까지 유지. → 표지가 선 주기에 실제 외부 분쟁이 섞였고 뒤에 외부 거래가 상쇄하면 v3 가 「조정 없는 일치」로 풂(Q6) |
| 판 중단 뒤 B(B 체결이 그 주기 Observe 뒤) | A: 다음 주기 RECHECK. **B: 표지 없음 → 고착 그대로**(Q8) |

- 코드 경로로: R1-1 이 v1 에서 열린 이유는 「다시 읽은 로컬 대 **옛** 스냅숏」이었음. v3 해제는 Observe 의 깨끗한 갈래(B2 if@650)에서만 일어나고 그 갈래는 **새 수집 + 새 로컬 읽기**의 비교에서만 옴 → 주기 k+1 의 실제 차이(10 vs 15)가 비교를 blocking 으로 만들어 해제 갈래에 못 감.
- R2-1(a192 결속 누락 · 비연속 영구 승격) · R2-2(건강 계수): v3 는 Observe 호출(call@464)을 건너뛰지 않음 — 탐침 RunOnce 편집은 읽기 2종 + Observe 뒤 메서드 호출뿐이고 `cycle.Err`·`judgeHoldings`(call@477)·`note`(defer@426) 무변. `beginOrderTerminalObservation`(call@647) · `advanceStreaks`(call@710/@735) · `refuteUsableAdjustmentCredits`(call@716)·크레딧 회계(B27 range@826)는 그대로 돎 → 그 증거가 빠질 구조가 없음. Observe 편집은 깨끗한 갈래에서 **오늘 Awaiting 으로 가던 블록**(append@686)만 Cleared 로 돌림. (시험으로 R2-1·R2-2 입력을 돌리지는 않음 — 구조 논증)

## Q5. `RECHECK_MATCHED` 소비자

| 자리 | 하는 일 | 생산 도달성 |
|---|---|---|
| 상수 정의 reconcile_states.go:65(주석 :58-64 「task 6.3 이후 수량 불일치를 스스로 닫지 않음 … 재조회만 한 생산자가 정직하게 이름 붙이려 남김」) | — | 쓰는 생산 코드 0(grep: 비시험 Go 에서 상수 참조는 정의 · `ValidReconcileReleaseCause` · risk_bucket_owner 2곳 · mismatch.go:865 주석뿐) |
| `ValidReconcileReleaseCause` B2 case@113 — RECHECK_MATCHED · ADJUSTMENT_APPLIED · OPERATOR · ORDER_TERMINAL_CONFIRMED | 쓰기 검사(`ReleaseReconcile` :269 · `ReleaseReconciles` :351 · 종결 해제 :106) | 이미 유효 → **상수 집합 편집 불필요** |
| `recordRiskBucketBrokerZeroObservation` B12 if@279(:280) | 위험 버킷 owner 의 「공식 보유 0 관측」 기록: 그 종목의 **가장 최근 해제** QM 행 원인이 RECHECK_MATCHED·ADJUSTMENT_APPLIED 가 아니면 `released_reconciliation` 으로 막음. ADJUSTMENT_APPLIED 이면 B15 if@297 로 값 0 조정 기록 · 조정 뒤 관측까지 요구 — **RECHECK_MATCHED 는 그 추가 요구 없이 통과** | 생산 호출자 0(함수 주석 :225-227 「production constructor/caller 없음」, CodeGraph callers = 시험 1) |
| `loadRiskBucketBrokerZeroAuthority` :677(`valid` 식) → B5 if@678, ADJUSTMENT_APPLIED 면 B7 if@684 조정 재검 | owner 해제 영수증 검증 | 호출자 `validateRiskBucketOwnerReleaseReceipt`(:838) · `releaseRiskBucketOwner`(:966) — 후자의 생산 호출은 구조 시험이 금지(risk_bucket_fill_test.go:343 `.releaseRiskBucketOwner(` 비시험 파일에 있으면 실패) |
| 시험: reconcile_states_test.go:264-353(유효 집합 열거 :350) · a138 · a066 · risk_bucket_owner_test(픽스처가 RECHECK_MATCHED 행 삽입) · restore_test.go:755(주석 — 예전엔 깨끗한 비교가 RECHECK_MATCHED 로 닫았음) · execgw reconcile_projection_test.go:209 | — | — |
| 콘솔 · cmd/tossctl · a194 · 로그 | release cause 를 읽는 생산 코드 0(grep `ReleaseCause|release_cause|RecheckMatched` — a192 로그가 ORDER_TERMINAL 상수를 쓰는 것뿐) | — |
| 정본 | `RECHECK_MATCHED` 문자열 0(openspec/specs 전체 grep). order-execution spec.md:353 「해제는 재조회 일치와 원인 기록을 요구한다」 | — |

- 비교(사실만): RECHECK_MATCHED 로 풀린 행은 broker-zero 경로(지금 비생산)가 **조정 증명 없이** 근거로 받음. 새 원인이면 그 경로는 `released_reconciliation` 으로 거절(fail-closed — a192 M7 과 같은 처분). 새 원인은 `ValidReconcileReleaseCause` case 1 + 유효 집합 시험(:350) 편집이 따름. 이력: RECHECK_MATCHED 는 task 6.3 전 추적기가 깨끗한 비교마다 쓰던 원인(restore_test.go:755 주석) — 운영 장부에 옛 행이 있을 수 있음(운영 장부 안 읽음 — 미확인).

## Q6. 정본 MODIFIED 대상 원문과 「우연한 일치」 원칙

- :53(불일치 시 진입 차단 본문) — 「허용 오차를 넘는 불일치가 확인되면 신규 진입은 차단되고(SHALL) …」 v3 는 차단 쪽 무변이라 이 문단은 바꿀 필요 없음(사실: v3 가 바꾸는 것은 해제만).
- :77 원문: 「해제 규칙의 정밀화(SHALL): 비영구 차단의 자동 해제는 **조정 이벤트가 반영된 뒤의 재조회 일치**, 또는 「missing-order 종목 차단은 …」 Requirement 가 정한 **결속된 주문 전부의 종결 기록 뒤의 재조회 일치**에만 근거하며, 각각 구분되는 release cause(ADJUSTMENT_APPLIED 계열, ORDER_TERMINAL_CONFIRMED)와 원인 기록을 남긴다. 조정도 종결 증명도 없이 우연히 일치한 단발 관측은 차단을 해제하지 못하고(SHALL NOT), 영구 불일치의 해제는 운영자 확인뿐이다(SHALL).」 → 바꿔야 하는 최소: 근거 목록에 셋째 근거와 그 원인 이름 추가, 마지막 SHALL NOT 의 「조정도 종결 증명도 없이」에 셋째 근거를 넣음.
- :303 Requirement 의 :322-325 원문: 「이 요구사항은 해제의 증거 요건을 약화해서는 안 된다(SHALL NOT). 조정이 기록되지 않은 심볼은 재조회가 아무리 여러 번 일치해도 조정 경로로는 자동 해제되지 않으며, 그 밖의 자동 해제는 「missing-order …」 Requirement 의 조건을 전부 갖춘 경우뿐이다(SHALL).」 → 「그 밖의 자동 해제는 … 뿐이다」에 셋째 경로 추가.
- :347 Scenario 「조정 없는 일치의 반복」: WHEN 「조정이 한 번도 기록되지 않았고 missing-order 종결 증명 해제의 조건도 갖추지 않은 심볼의 재조회가 여러 주기 연속으로 일치한다」 THEN 「어떤 주기에서도 자동 해제되지 않고 운영자 해제만 남는다」 → WHEN 에 「표지 조건도 갖추지 않은」 추가.
- 그 밖에 걸리는 문장: :87 · :333 THEN 의 원인 이름(ADJUSTMENT_APPLIED)은 해당 경로 한정이라 무변. a192 :381 「같은 종목의 수량 불일치가 관측됐으면 **이 경로로** 해제해서는 안 된다」는 경로 한정 — v3 가 같은 행을 다른 원인으로 푸는 것과의 관계를 정본이 말하지 않음. a192 :448 「이 경로는 수량 불일치 해제 규칙 … 을 바꿔서는 안 된다」.
- 표지가 위조 불가능한 증거인가 — 사실:
  - 표지의 두 재료는 프로세스 안 사실: `fill_events.id` 는 AUTOINCREMENT·단일 연결 직렬 쓰기라 커밋 순서(a192 C12 근거, journal.go:174), `ConvergeReport.Converged` 는 메모리 보고. 외부 입력으로 직접 쓸 수 없음.
  - 그러나 표지가 증명하는 것은 「그 종목 체결이 로컬 읽기 뒤에 커밋됐고 그 종목이 조정되지 않았다」뿐이고 **불일치가 그 체결 때문이었다**가 아님: R1-1(실제 외부 매도 5 주 동반)에서도 표지가 섬(탐침 `marks after k`). 그 뒤 외부 거래가 상쇄해 새 비교가 일치하면 RECHECK 해제(탐침 R1-1 변형).
  - (가) 종목 워터마크에 계좌 조건 없음(position_adjustments.go:190-198 [손], W3-b 재현) → 다른 계좌의 같은 종목 체결이 표지를 세움. 엔진은 기동 시 공식 계좌 목록의 첫 레코드 하나로 범위를 정함(정본 :35-:49, `NewReconcileDriver` 가 AccountRef 1개 요구 reconcileloop.go:325) — 장부에 다른 계좌 체결 행이 실제로 있는지는 운영 장부를 안 봐서 **미확인**(a192 M16 집계는 「종목 빈 체결 0」만).
  - (나) 「미반영」에는 수렴 **거절**(`soleLiveMarket` — 살아 있는 인스턴스 0(W2 의 경로) · 복수 시장, converge.go:276-309)이 섞임. 복수 시장 거절 종목은 오늘 조정이 영영 안 오므로 운영자 해제뿐인데, 그 종목에 아무 체결(다른 계좌 포함)이 창에 오면 표지 → 뒤의 일치로 RECHECK.
  - 귀결: v3 의 해제 근거 = 「표지 + 그 뒤 새 수집·새 로컬 읽기의 일치 한 번」. 틀릴 수 있는 것은 오늘 원칙이 막는 그 모양 — 일치가 진행 중 체결·정산 전파 때문에 우연히 보인 경우(mismatch.go:51-57 주석) — 가 표지된 종목에서는 막히지 않음.

## Q7. 편집 표면과 고정 장치

| 기존 함수 | 생산 호출자 · 자리 수 | 필요성 | 고정 장치 · 겹침 |
|---|---|---|---|
| `ReconcileDriver.RunOnce` | `Run` call@412 — 1 | 필수(표지 재료를 쥔 유일한 곳) | a071 금지 문자열(`ProtectionReadiness`·`ReadinessAdapter`·`checkProtection`, a071_readiness_assembly_test.go:28-35) — 쓰지 않으면 비적중. a199 는 같은 파일 `Context.ReconcileDriver`(:368)·패키지 주석(:64-70, 6줄↔6줄)만 — RunOnce 무겹침·줄 번호 불변. a194 FLM 은 `judgeHoldings` 만 |
| `Tracker.Observe` | `RunOnce` call@464 — 1 | 필수(또는 아래 `orderTerminalCandidate`) — 깨끗한 갈래 둘째 후보 + 표지 수명 | a192 FLM 번들 대상 함수(archive …/function-logic/internal-reconcile--tracker.observe). 편집 지점 뒤 분기 id 가 밀림(B9 뒤 전부) |
| `Tracker.orderTerminalCandidate` | Observe call@682 — 1(+시험 2) | Observe 대신 고를 수 있음 — 결속 없음 B2 if@380 → return@381 앞에 넣어야 함 | a192 새 함수, 번들 없음 → 이번이 첫 FLM |
| `Tracker.persist` | Observe call@786 — 1 | 필수(원인 리터럴 :938, 또는 새 tx 메서드 갈래) | a192 FLM 번들 대상 |
| `Tracker.enterReconcileWithBinding` | persist call@902 — 1 | **불필요/불충분** — entered=false 행엔 안 불림(Q1(c)) | — |
| `Tracker.Restore` | gateway.go:279 · cmd/tossctl/engine_reconcile.go:283 — 2 | 표지를 프로세스 기억으로 두고 같은 객체 Restore 규율(a192 D11 선례)을 따르면 대입 1 | a192 FLM 대상 |
| `ValidReconcileReleaseCause` | — | RECHECK_MATCHED 면 불필요, 새 원인이면 case 1 | reconcile_states_test.go:350 열거 |
| `buildGateway` | — | 불필요(기존 `Terminal`/`Journal` 은 `*journal.Journal` — 인터페이스에 메서드를 더해도 배선 무편집) | **a166 census 적중 함수**(shapes.json 키 `internal/app/engine/gateway.go||buildGateway`) — 건드리면 accountbudgetruntime 시험 깨짐. census 에 reconcileloop.go · mismatch.go · order_terminal_release.go 없음 |

- CodeGraph `affected --filter '*_test.go' -d 1` 3파일 → 19 시험 파일(목록 `codegraph-affected.txt`).
- 진행 중 change 겹침(브랜치 diff, merge-base 기준): a199 = reconcileloop.go 의 `Context.ReconcileDriver` + 주석(위). a200 = 시험 2 파일만(internal/app/engine · internal/reconcile). a197 · a087 = 대상 파일 무편집. a167 · a171 = 브랜치 없음, change 문서에 대상 파일 언급 0(grep).
- a192 번들 함수를 다시 고치면: 이번 change 가 함수마다 새 FLM 번들(gate ⑤ — 시험 함수 포함 기존 함수 본문 편집이면 필수), Observe 는 분기 위치 번호가 밀려 Branch Test Map 재번호(기억 「분기 번호는 위치다」), a192 아카이브 번들은 옛 ast 로 남음.

## Q8. v3 로도 남는 모양

| 모양 | 귀결 |
|---|---|
| 재기동(표지 소실) | 오늘과 같음 — 운영자 해제 + a194 알림 |
| 판 중단 뒤 종목 B(체결이 표지 계산 뒤·다음 로컬 읽기 전) | **탐침: B 고착 그대로**(표지 조건 (가)가 거짓 — 체결이 아직 안 옴). 오늘과 같음 |
| 표지 뒤 또 체결 · 투영 거절 가림 | C12 유사를 두면 유지 → 오늘과 같음 |
| 영구 승격 뒤 | 계좌 행은 운영자, 종목 행만 풀릴 수 있음 |
| 계좌 범위(빈 종목 수량 불일치) | 표지 키 없음(탐침은 빈 종목 건너뜀) → 오늘과 같음 |
| 전략 진입 backfill(`backfillConfirmedStrategyFillTx`, fill_events 행 없이 투영만) | 워터마크 안 움직임 → 표지 없음 → 오늘과 같음(claim-check-v2 (p1)) |
| 워터마크 읽기 실패 | 표지 없음 → 오늘과 같음 |
| 다른 종목 분쟁으로 비교가 계속 blocking | 해제 대기(크레딧·a192 와 같음) |
| pending(비영속) 블록에 표지 → 해제 시도 | `ReleaseReconcile` 가 활성 행 없음 → persist 오류(:946-949) — 오늘 크레딧 경로와 같음 |

## AST 목록(20, `ast/`)
RunOnce · Observe · persist · Refresh · Restore · blocksFor · blockFromReconcileState · hasBlockFor · enterReconcileWithBinding · orderTerminalCandidate · releaseOnOrderTerminal · updateOrderTerminalBindings · beginOrderTerminalObservation · orderTerminalReleaseVerdict · ConvergeQuantities · recordRiskBucketBrokerZeroObservation · loadRiskBucketBrokerZeroAuthority · ValidReconcileReleaseCause · ReleaseReconcile · EnterReconcile.

## 정리
- 사본 worktree 제거(`git worktree list` 에 a201-v3-evidence 0), 메인 `git status --short` 0줄, HEAD `7fec501f1` 불변. 실행한 시험: 사본에서 `go test ./internal/app/engine -run 'TestA201V3Probe…' -count=1`(3회) — 전체 스위트·gate·sdd-sync 안 돌림.
- 안 본 것: 운영 장부(계좌 수·옛 RECHECK_MATCHED 행), `latchReleasedOwnerLateFillInTx` 생산 도달성, 표지 뒤 투영 거절 가림의 탐침, R2-1·R2-2 입력의 실제 실행, Hindsight/ai-memory recall(이번 조사에선 안 함), a199·a200 영역.
