# a201 1단계 증거 — 사전 증거의 현재 HEAD 재확인(1.3) · design 전제(1.5)

- HEAD `86a131a51`(base `f9b818a0e` = origin/main `7fec501f1` + 등록 문서; 그 뒤 커밋은 base 기록 · 재현 시험 · 문서뿐). production Go 는 `7fec501f1` 과 동일.
- 사전 증거: Manager scratch `a201-evidence/evidence.md`(HEAD `7fec501f1`, AST 20) · `a200-evidence/evidence.md`(F1~F14, W0~W6). 이 문서는 그 좌표를 다시 뽑아 대조한 결과이고, 원문을 옮기지 않음.
- AST: `go build ./tools/logic-map` → `cited-ast/*.ast.json` 36(사전 증거 20 + 새 16), 요약 `ast-summary.txt`(사전 증거의 `sum.py` 그대로). 좌표 `B<n>(kind@줄)` · `call@줄` · `return@줄`. **[손]** = AST 가 열거하지 않는 SQL · 리터럴 · 구조체를 손으로 읽음.
- 운영 집계: `ops-aggregates.md`(Manager 제공, 읽기 전용, 건수만).

## 1. AST 재추출 — 사전 증거와 다른 곳 0

| 사전 증거 20(파일명) | 결과 |
|---|---|
| ApplyPositionAdjustment · Compare · ConvergeQuantities · FillWatermark · IngestExternalPositions · LocalStateFromJournal · MaxFillEventID · Positions · RunOnce · Tracker.{AdjustmentApplied, Observe, Permanent, Refresh, persist} · blockFromReconcileState · blocksFor · creditUsableBy · hasBlockFor · soleLiveMarket · symbolsInDispute | **20 바이트 동일**(`cmp`) — sha · 시작/끝 줄 · 분기 · 호출 좌표가 사전 증거 그대로 |

새로 뽑은 16: `HeldBySymbol`, `quantitiesAgree`, `canonicalPresentQuantity`, `promotionKeys`, `Tracker.{advanceStreaks, clearStreaks, refuteUsableAdjustmentCredits, beginOrderTerminalObservation, orderTerminalCandidate, updateOrderTerminalBindings}`, `ReconcileDriver.{Run, logOrderTerminalReleases}`, `Notifier.{Notify, publishBestEffort}`, `notifierAlerter.ManagedPositionClosedExternally`.

## 2. 사실 재확인(사전 증거의 번호 유지)

| # | 사실 | 근거(HEAD) |
|---|---|---|
| F1 | 주기 순서 안정화 → 로컬 읽기 → 비교 → 접기 → 수렴 → Refresh → Observe → 편입 | RunOnce call@428 · @434 · @439 · @445 · @451 · @458 · @464 · @477 |
| G1 | stale 은 판 전체 중단 | ConvergeQuantities B11 if@228 → return@229 — **실제 RunOnce 로 재현**(repro 「판 중단 뒤 종목」: converged 0, A · B QM 2) |
| G2 | `Refused` 버림, `Observe` 는 같은 diff | RunOnce call@451 뒤 `converged` 에서 `len(Converged)` · `Closed` 만(@455-456), call@464 인자 `diff` |
| F7 | 조정 tx 재검증 (a) 수량 → (b) 워터마크 순 | ApplyPositionAdjustment B15 if@253 → return@254, B17 if@267 → return@268 — **재현: 조정 tx 직전 체결도 (a)로 걸림** |
| Q3 | 종목 워터마크는 계좌 조건 없음 | position_adjustments.go:190-198 [손 SQL] — **재현: 다른 계좌의 같은 종목 체결이 (b) stale 을 냄**(W3-b) |
| Q1-2 | 「외부 청산」 알림은 수렴 루프 안 동기 호출 | ConvergeQuantities call@254 → `notifierAlerter.ManagedPositionClosedExternally` call@141 → `Notifier.Notify` B1 if@153 → `publishBestEffort` call@154 → `Publisher.Publish` call@184(고루틴 · defer 없음 — AST `go_statements` 0). 상한 `DefaultPublishTimeout` 10초(alert_lease.go:17 [손]) — Publisher 의 실제 시한은 안 봄 |
| F12 | 같은 키의 QM 이 missing-order 블록을 가림 | blocksFor B1 range@1441 → B2 range@1455, Observe B17 range@738 → B18 if@739 — 재현의 블록 목록이 QM 하나뿐 |

## 3. design 전제(1.5)

| # | 전제 | 판정 | 근거 |
|---|---|---|---|
| (a) | `RunOnce` 가 「이번 주기에 조정이 반영된 종목」을 안다 | **참** | `ConvergeReport.Converged []ConvergedPosition{Market, Symbol, Local, Quantity, PositionID, Applied, ClosedExitState}`(converge.go:84-115 [손]) · `Credited []string`. stale 로 판이 멈춰도 커밋분은 보고서에 남음(defer@172, 시험 `a083b_partial_convergence_test.go`). RunOnce 는 이미 `converged` 를 손에 쥠(call@451). **`ConvergeQuantities` 편집 불필요** |
| (b) | 종목 워터마크 · 전역 최대 체결 id 를 RunOnce 가 부를 수 있다 | **참(조건부)** | `d.opts.Journal *journal.Journal`(구체 타입, reconcileloop.go:120) 에 `FillWatermark`(position_adjustments.go:190) · `MaxFillEventID`(reconcile_order_terminal_release.go:75) 공개. **종목 워터마크는 계좌 조건이 없어 다른 계좌의 같은 종목 체결에 움직임**(SQL [손] + W3-b 재현). 운영 장부가 단일 계좌인지는 코드로 못 정함 — 영향은 (가)를 느슨하게 할 뿐이고 (다)가 거름(claim-check C-D3a) |
| (c) | 로컬 재읽기 함수 · `Compare` 의 일치 판정 | **참(위치 제약)** | 재읽기 = `HeldBySymbol`(공개, compare.go:148 — `LocalStateFromJournal` 이 쓰는 바로 그 합: 비-CLOSED · 종목 수준 · 정확 decimal). 일치 판정 = `canonicalPresentQuantity`(compare.go:497 — 부재는 "0") + `quantitiesAgree`(compare.go:645) — **둘 다 비공개**, 호출자 Compare 하나. 재판정 함수는 `internal/reconcile` 안 새 파일이어야 함 |
| (d) | `Observe` 생산 호출자 | **1 자리** | reconcileloop.go:464(grep · CodeGraph — codegraph-baseline.md) |
| (e) | diff 에서 항목을 뺀 사본이 깨는 불변식 | **design D4 「다른 종목 판정에 영향 없음」 거짓** | `Observe` 는 diff 하나에서 `symbolsInDispute`(call@644) · `promotionKeys`(call@645) · a192 결속 흔적(call@647 · @808) · 블록(`blocksFor` call@704)을 모두 뽑고, `diff.BlocksEntry()`(B2 if@650)로 갈래를 고름. 뺀 결과 비교가 깨끗해지면 **해제 갈래**(B6 range@666 — 활성 QM 블록 전부에 대해 `creditUsableBy` → 해제 / a192 `orderTerminalCandidate`)와 `clearStreaks`(call@655)로 감. 뺀 종목은 `disputed` 에서 빠져 그 종목의 앞 크레딧이 반박되지 않음(B27 range@826 → B29 if@841). digest 는 `Diff` 에 없음(compare.go:292-304 [손]), as-of · 계좌는 사본이 그대로 가짐. **탐침(§4)에서 실측: 앞 주기 크레딧이 있던 종목이 그 주기에 ADJUSTMENT_APPLIED 로 해제됨** |
| (f) | D5 — 수량만 뺐을 때 missing 블록 | **선다, a192 로 깨끗한 2주기 뒤 해제** | 탐침(§4): W2 · W3-a · tx 직전 · W3-b 네 모양 모두 그 주기에 missing-order 블록(장부 cause QUANTITY_MISMATCH, 진입 거절) → 일치 주기 1 「armed」 유지 → 일치 주기 2 `ORDER_TERMINAL_CONFIRMED` 해제. design D5 의 기준(「2주기 안에 풀리면 D5 를 뺀다」) 충족. **단 앞 크레딧 위의 W2 에서는 D5 없이 풀리지 않음**(QM 블록이 키를 쥐고 있어 결속이 안 생김 — 그대로 고착) |

## 4. (e)(f) 탐침 — scratch 사본, 추적 파일 무편집

- 사본: `git worktree add --detach <scratch>/a201-probe/wt 86a131a51`. 최소 탐침 편집 = RunOnce 에 「로컬 읽기 전 `MaxFillEventID`」 + 「Observe 직전 `reconcile.ProbeRejudge`」(새 파일, D3 의 (가)(나)(다) 최소 흉내, `PROBE_D5=1` 이면 다시 읽은 로컬에서 추적 아닌 같은 종목 missing-order 도 뺌). 원문 diff `probe-d5/probe.diff`, 실행 원문 `probe-d5/run-d5-off.txt` · `run-d5-on.txt`(재현 시험을 단언 대신 로그로 돌림).

| 재현 모양 | 편집 전(현재) | 탐침 D5 끔 | 탐침 D5 켬 |
|---|---|---|---|
| W2 | QM 고착(5주기), 재매수 편입 거절 | 그 주기 missing 블록(진입 거절) → 2주기 뒤 `ORDER_TERMINAL_CONFIRMED` 해제 → 재매수 편입 1 | 그 주기 행 0 · 진입 허용 · 재매수 편입 1 |
| W3-a · tx 직전 · W3-b | QM 고착 | 같음(2주기 뒤 해제) | 행 0 |
| 앞 크레딧 위의 W2 | QM 고착(크레딧 반박) | **QM 고착 그대로**(missing 이 비교를 막아 앞 크레딧 반박, 결속 없음) | 그 주기에 **앞 크레딧으로 `ADJUSTMENT_APPLIED` 해제** |
| 판 중단 뒤 종목 B(체결이 Observe 뒤) | A · B 고착 | A 는 2주기 뒤 해제, **B 고착** | A 행 0, **B 고착** |

- 탐침은 design 의 일부(D7 로그 · D8 실패 처분)를 흉내 내지 않았고 변이 · 경합 시험도 아님 — 전제 (e)(f)의 방향만 보는 측정.
