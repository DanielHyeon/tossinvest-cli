# a198 보강 2 증거 — D9(기동이 ACKED 취소의 쓰기를 마침)·D10(세션이 IN_DOUBT 취소를 대상의 종결 기록으로 닫음) (2026-11-24)

기준: 브랜치 HEAD `f8a5bb898`(base `e1eaf720c` + a198 구현·r2). AST 는 `analysis/cited-ast/d2_*.ast.json`(tools/logic-map `extract_go_ast.go`, 같은 작업 트리에서 이번에 뽑음 — 생성 스크립트 `analysis/d9-d10-ast.sh <repo-root>`, 함수·파일 목록은 아래 표). 조사 산출물(`scratchpad/inv-F/`)은 읽지 않고 전부 다시 뽑았다. 분기 번호는 각 ast.json 의 `B<n>`(파일 줄).

| ast | 파일 | sha256(앞 12) | 줄 | 분기 |
|---|---|---|---|---|
| d2_Recovery_confirmAcked | internal/reconcile/acked_boot.go | 4919ea34abc4 | 45-73 | 6 |
| d2_Recovery_Run | internal/reconcile/recovery.go | c8054430ffbe | 247-345 | 13 |
| d2_inDoubtSettle_sweep | internal/app/engine/in_doubt_settle.go | 247eedffa947 | 159-192 | 7 |
| d2_inDoubtSettle_resolve | 〃 | 〃 | 195-222 | 4 |
| d2_Resolver_resolveCancel | internal/execgw/amend_indoubt.go | bd06ffeddaaf | 44-158 | 20 |
| d2_Attempt_DispatchVerified | internal/journal/dispatch.go | f77e0bab6044 | 132-223 | 19 |
| d2_Attempt_ResolveConfirmed · ResolveFailed | internal/journal/resolution.go | f4e27e79f908 | 48-59 · 63-70 | 1 · 0 |
| d2_Attempt_transition · MarkAcked | internal/journal/durability.go | 29ec7a1849fa | 548-674 · 501-506 | 23 · 0 |
| d2_ValidateTransition | internal/journal/lifecycle.go | f71b66c49c64 | 54-68 | 4 |
| d2_Journal_PendingAttempts | internal/journal/recovery.go | bc81bfd5a365 | 30-51 | 4 |
| d2_Gateway_unsettledFor · checkSymbolFree · Cancel | internal/execgw/a094_shared_judgements.go · gateway.go | ccc2c9efa6c1 · c304e32a62e4 | 30-46 · 802-838 · 407- | 4 · 10 · 0 |
| d2_Journal_ConfirmedCancelOf · intentOrdersAwaitingCloseQ | internal/journal/exit_proposal_release.go | 8faee2a9f9f1 | 150-174 · 248-276 | 4 · 4 |
| d2_Journal_LiveOrdersForSymbol · markFillRefused · updateFillSnapshotRefusal · upsertFillSnapshot | internal/journal/fills.go | 92bc44f947ee | 1854-1919 · 710-728 · 833-859 · 730-773 | 7 · 3 · 2 · 2 |
| d2_Journal_ResolveCurrentOrderIDScoped | internal/journal/lineage.go | 739433026795 | 312-344 | 9 |
| d2_brokerstate_Derive · State_IsTerminal | internal/brokerstate/derive.go | 43911775579f | 346-510 · 94-102 | 50 · 3 |
| d2_Saga_settleStep · resolveStep · classifyCancel | internal/flatten/flatten.go | 1f66eed56b47 | 301-416 · 423-446 · 454-474 | 11 · 6 · 6 |
| d2_Saga_sell · planLiquidation | internal/flatten/liquidate.go | ef9ad2603edf | 444-540 · 279-327 | 13 · 6 |
| d2_ExitObserver_clearTheSymbol_head | internal/app/engine/exitloop.go | d090c3f70d6b | 1651-1753 | 17 |
| d2_TestA094AnAckedCancelOrNumberlessPlaceIsOnlyNamed · BootAndPostPlaceShareOneConfirmation | internal/reconcile/a094_acked_boot_test.go | 3b783a366a82 | 181-205 · 265-304 | 3 · 10 |
| d2_TestA193OnlyInDoubtPlaceAttemptsAreTouched · TheSettlerNeverReplays | internal/app/engine/a193_in_doubt_settle_test.go | e96d11fd09fa | 719-745 · 761-777 | 5 · 4 |

편집 전 `clearTheSymbol` 은 `cited-ast/ExitObserver_clearTheSymbol.ast.json`(base, sha `bc81b5fd…`) 과 `git show e1eaf720c:internal/app/engine/exitloop.go`.

## 0. 조사 사실(F) 재확인

| F | 결과 | 근거 |
|---|---|---|
| F15 | **맞음 + 보강.** CANCEL 의 ACKED = `DispatchVerified` B11@181 `MarkAcked(번호)` 커밋 뒤 B15@199 `Settle(CONFIRMED)` 전의 중단·쓰기 실패. CANCEL 은 확인 읽기 없음(`roundTripFor` 가 PLACE 만 — roundtrip.go:71-73, `verify == nil`). 번호가 비면 B10@174 `MarkInDoubt(ack_without_order_id)`. **보강**: 그 번호는 대상 주문 번호가 아니다 — 계약 `OrderOperationResponse.orderId` 「정정/취소로 새로 발급된 주문 식별자. 원주문의 orderId 와 다릅니다」(openapi.latest.json components.schemas). 정상 경로의 CONFIRMED CANCEL 도 같은 번호를 갖는다(Settle 은 broker_order_id 를 바꾸지 않음 — transition L619 는 `o.brokerOrderID != ""` 일 때만 씀) | DispatchVerified AST |
| F16·F17 | 맞음. `transitionTable[StateAcked] = {CONFIRMED, IN_DOUBT}`(lifecycle.go:46). `ResolveConfirmed` 의 from = `{IN_DOUBT, ACKED}`(resolution.go:54), 번호 빈 값 거절(:49). `ResolveFailed` 의 from = `{IN_DOUBT}` 뿐 | resolution AST |
| F18 | 맞음. `confirmAcked` B2@51 `rec.Kind != KindPlace` → `alertAcked(…「a CANCEL/AMEND acknowledgement has no read-back rule; it stays ACKED」)`, false. 호출 자리 `Recovery.Run` B5@279(`rec.State == ACKED && r.confirmAcked`) — true 면 `report.ConfirmedAcked` 에 넣고 continue, false 면 `blockedSymbol` → `StillPending` | confirmAcked · Run AST |
| F19·F20 | 맞음. a094 design D−4.2(archive/2026-09-30-a094…/design.md:353-378): 「해소기 matcher 에 주문 번호 판별자가 없다 … `res.BrokerOrderID = order.OrderID` 로 기록된 번호를 덮는다(indoubt.go:307)」 → 목록 대조 폴백을 통째로 뺌 — CANCEL 고유의 위험 서술은 없다. 그 앞 판(:572-574)은 CANCEL·AMEND ACKED 를 `MarkInDoubt` → `resolveCancel` 로 보내려 했다. 정본 order-execution.md:112(「IN_DOUBT 해소」 Requirement 안)가 「취소·정정 attempt … 상태를 바꾸어서는 안 되고」「바이트 일치 외의 어떤 추론으로도 ACKED attempt를 종결 … 해서는 안 되며」 | 원문 |
| F21·F23 | 맞음. `sweep` B4@175 `rec.State == IN_DOUBT && rec.Kind == KindPlace` 만 targets. 정본 :125 「대상은 IN_DOUBT이면서 발주인 attempt뿐이다(SHALL). 취소·정정 attempt는 기동이 해소한다」, 시나리오 :245-247 「…ACKED 상태이거나 취소·정정인 미종결 attempt를 읽으면 → 상태는 바뀌지 않고 해소 절차에 넣어지지 않는다」 | sweep AST |
| F25·F26 | 맞음. `PendingAttempts` = `state IN (RECORDED, DISPATCH_STARTED, ACKED, IN_DOUBT)`, 종류 무관(recovery.go:31-34). 가드 둘(`checkSymbolFree` B1·B2@804-808 과 `UnsettledOnSymbol`)이 같은 `unsettledFor`. UNRESOLVED 는 `checkSymbolFree` B3@814 `!raisesExposure` 이면 통과 — 위험 감소(손절·취소)를 막지 않음 | unsettledFor · checkSymbolFree AST |
| F28·F29 | 맞음(§2 상세). 원장 API 중 fill 기록으로 attempt 를 닫는 것 없음 — 상태 전이 쓰기 전수: `Attempt.transition`(durability.go:634) · `ConfirmReplacement`(lineage.go:157) · strategy_dispatch_runtime.go:787·928(전략 경로), `ResolveConfirmed/Failed` 호출자는 replay.go:344 · acked_boot.go:65 · indoubt.go:311·343 · amend_indoubt.go(브로커 관측) 뿐 | grep 전수 |
| F4~F6 | 맞음(§5 표) | resolveCancel AST |

**틀린 F 없음.** 보강 하나(F15 — ACKED 취소의 번호는 취소 기록의 새 번호)와 조사에 없던 사실 셋(§2 의 REPLACED, terminal+fail_closed 공존, §6 flatten)을 아래에 적는다.

## 1. (a) `confirmAcked` 와 D9

- 분기(AST): B1@47 intent 읽기 실패 → 알림, false · **B2@51 종류 ≠ PLACE → 알림, false(D9 편집 자리)** · B3@55 번호 없음 → 알림 · B4@59 `ConfirmPlacedOrder` 실패 → 알림 · B5@64 `Resume` 성공 → `ResolveConfirmed(rec.BrokerOrderID, ReasonBootAckedConfirmed, …)` · B6@68 쓰기 실패 → 알림(「ledger write failed」), false · 끝 true.
- D9 에 필요한 것: `Resume(rec.ID)` → `ResolveConfirmed(rec.BrokerOrderID, <사유>, <상세>)`. 전이 ACKED→CONFIRMED 허용(F16·F17). CANCEL 이므로 transition 의 PLACE/AMEND 전용 블록(L575 — 체결 증거 시각 정렬)을 타지 않는다. 브로커 호출 0.
- 번호 없는 CANCEL ACKED: `DispatchVerified` 로는 생기지 않는다(B10 이 IN_DOUBT 로 보냄). 시험 픽스처(`MarkAcked(ctx, "")`)로만 — D9 는 번호가 있을 때만, 없으면 B3 처럼 알림만.
- 그 뒤 흐름: true 면 `Run` 이 `ConfirmedAcked` 에 넣고 `blockedSymbol`·`StillPending` 을 건너뜀(PLACE 성공과 같은 자리). **알림**: PLACE 성공은 알림 0 — D9 도 성공 시 알림 0 이 대칭(질문 Q3). 쓰기 실패는 B6 과 같은 알림 유지.
- 구조 고정 `TestA094BootAndPostPlaceShareOneConfirmation`(:297): `calls("acked_boot.go","confirmAcked")["ConfirmPlacedOrder"]` 가 참이어야 함 — D9 는 그 호출을 지우지 않으므로 **무영향**. 같은 시험의 `confirmCreatedOrder` 단언도 무관.
- **뒤집히는 시험** `TestA094AnAckedCancelOrNumberlessPlaceIsOnlyNamed`(181-205, 표 행은 `{"cancel", KindCancel, "O-target"}` 하나): B2@196 `stored.State != ACKED || reads != 0` → D9 뒤 CONFIRMED 라 실패. `a094AssertNamed` 의 「알림 정확히 1」도 실패(성공 시 알림 0 이면). 수정: 취소 행을 「CONFIRMED · 읽기 0 · 알림 0」 으로, 정정(AMEND) ACKED·번호 없는 PLACE ACKED 행을 「ACKED · 알림 1」 로(시험 이름이 말하는 범위를 실제로 채움).
- 다른 뒤집힘 후보: 기동 복구를 거치며 CANCEL ACKED 를 만드는 시험 전수(MarkAcked + KindCancel + Recovery 동시 등장 파일) = a094_acked_boot_test.go 뿐. a094_exit_clear_test.go:776·a193 픽스처는 세션 중(기동 아님) — 무영향.

## 2. (b) 종결 증거 `confirmedOrderTerminalEvidence` 와 D10 의 조인

- 술어 원문(exit_proposal_release.go:281-296): `EXISTS (SELECT 1 FROM all_fill_snapshots f WHERE f.order_id = c.order_id AND f.terminal = 1 AND c.ownership_at < f.committed_at AND 종목·시장 일치 AND (f.side 빈 값 또는 일치) AND ((계좌·거래일 일치) OR (계좌·거래일·방향이 빈 legacy 행이고 legacy_snapshot_owner 의 소유 intent 가 c.intent_id)))`. `c` 는 order_id · ownership_at · intent_id · account_ref · market · trading_day · symbol · side. **`fail_closed` 를 보지 않는다.** 앞에 `allFillSnapshotsCTE`(fills.go:888-959 — scoped 행 + scoped 가 없는 legacy 행, legacy 소유자 = 후보 intent 가 정확히 1 인 것).
- 소비자 둘: `LiveOrdersForSymbol`(fills.go:1854 — `confirmed_orders` = CONFIRMED PLACE/AMEND 의 `broker_order_id`, 소유 intent 가 같은 계좌·시장·**거래일**·종목·방향에서 유일한 것만, `NOT 술어`, 그 뒤 `ResolveCurrentOrderIDScoped` 로 번호를 현재 후계로 바꿈) · `intentOrdersAwaitingCloseQ`(유일성 필터 없음).
- **취소 attempt 에서 대상의 소유 intent 찾기**: `mutation_attempts.target_order_id` = 길 치우기가 넘긴 `order.OrderID` = `LiveOrdersForSymbol` 이 후계로 바꾼 **현재 번호**. 후계 간선은 로컬 AMEND 의 CONFIRMED 로만 생기고(lineage.go:157-200 — attempt 의 broker_order_id = 자식 번호) 그 AMEND attempt 자신이 `confirmed_orders` 의 직접 소유 행이다 → **lineage 를 다시 따라갈 필요 없이** `broker_order_id = target_order_id` 인 CONFIRMED PLACE/AMEND 가 소유 후보. 범위: 그 소유 intent 의 계좌·시장·종목·방향 = 취소 intent 의 계좌·시장·종목·방향(취소 intent 의 방향 = 주문 방향, D1), 거래일은 **소유 intent 의 것**(취소 intent 의 거래일 = 발행일 — 전일 주문이면 다르다; `ConfirmedCancelOf` 의 전일 결속 잔여와 같은 함정을 피함). 소유 intent 가 거래일을 건너 둘 이상이면 닫지 않음.
- 스냅숏 `state` 열거(brokerstate/derive.go:58-89, 운영 쓰기 유일 자리 filldetect/ledger.go:60 `string(snap.Derived.State)`, terminal = `Derived.Terminal` = `State.IsTerminal()`): `OPEN_UNFILLED` · `OPEN_PARTIALLY_FILLED` · `CANCEL_PENDING` · `REPLACE_PENDING` · `UNKNOWN_BROKER_STATE`(terminal 0) / **`FILLED` · `CANCELLED` · `CANCELLED_PARTIALLY_FILLED` · `REPLACED` · `REJECTED` · `REJECTED_PARTIALLY_FILLED`(terminal 1)**. `unknown()` 은 terminal 0 · fail_closed 1(derive.go:588-594).
- **반례 탐색 — 「잔량이 살아 있는데 terminal」: 있다 = `REPLACED`.** 정의(derive.go:69-72) 「the exposure did not go away, it moved」 — 후계가 잔량을 든다. terminal 1 이 되는 것은 후계가 **원장에 있을 때만**(lineage = 로컬 AMEND, derive.go:203-208 「the lineage comes from the journal」; 없으면 :475 `replaced_without_successor` → UNKNOWN, terminal 0). `CANCELLED_PARTIALLY_FILLED`·`REJECTED_PARTIALLY_FILLED` 는 잔량이 취소·거부됨(살아 있지 않음). → D10 표에서 `REPLACED` 는 **닫지 않음**이어야 한다(`resolveCancel` 도 B19 park). 도달성: 취소 대상이 이미 lineage 로 후계 번호가 된 뒤라 대상 자신이 REPLACED 가 되려면 취소가 미종결인 동안 엔진 AMEND 가 CONFIRMED 돼야 하는데, 같은 종목 미종결 가드가 그 AMEND 를 막는다 — **도달 거의 불가, 그래도 표 밖으로 둔다.**
- **반례 — terminal 1 이면서 fail_closed 1: 있다.** `updateFillSnapshotRefusal`(fills.go:833-859)은 기존 행에 `fail_closed = 1` 만 세우고 `terminal` 을 안 건드림 — terminal 행 뒤에 거절된 관측(감소·순서 역전·호출자 fail-closed)이 오면 공존. 첫 관측의 거절은 terminal 0(`markFillRefused` B3@724 `refused.Terminal = false`). 기존 술어는 이 행을 「종결」로 친다(길 치우기는 이미 그 주문을 미체결 목록에서 뺌) — D10 은 Manager 결정대로 `fail_closed = 1` 이면 닫지 않는다(그 점에서 술어보다 좁다).
- terminal → 비종결 되돌림: upsert 는 `terminal = excluded.terminal` 을 그대로 씀(fills.go:741) — 막는 규칙은 없으나 추적 목록이 terminal 행을 빼므로(fills.go:1604·1627 `f.terminal = 0`) 운영 폴링은 다시 쓰지 않음. 판독과 쓰기를 한 트랜잭션에 두면 그 사이 되돌림은 원리적으로 없다.
- 번호 재사용: 술어가 계좌·거래일·종목·시장(scoped 행)과 `ownership_at < committed_at` 을 요구 → 다른 날·다른 계좌의 같은 번호 스냅숏은 맞지 않음. legacy 빈 범위 행은 `legacy_snapshot_owner` 의 유일 소유자 = 그 소유 intent 일 때만.
- 여러 행 일치(scoped + legacy, 같은 order_id): D10 은 일치하는 terminal 행 **전부**가 같은 판정으로 모일 때만 닫는다.

## 3. (c) IN_DOUBT CANCEL 을 닫는 기존 API — 새 API 가 필요한가

- `Attempt.ResolveConfirmed`(IN_DOUBT·ACKED → CONFIRMED, 번호 필수) · `Attempt.ResolveFailed`(IN_DOUBT → FAILED_CONFIRMED). 둘 다 `Attempt.transition` 이 **자기 트랜잭션**을 연다(durability.go:553 `BeginTx`) — 외부 트랜잭션을 받지 않음. 사유 코드: `resolveCancel` 이 쓰는 `journal.ReasonResolvedFound`("in_doubt_resolved_found") · `journal.ReasonResolvedAbsent`("in_doubt_resolved_absent").
- 「판독 + 종결을 한 트랜잭션」(Manager 지시)은 기존 둘을 그대로 묶어서는 안 된다. 선택지:
  - (가) **새 파일의 새 함수가 자기 트랜잭션 안에서** 증거를 읽고 `UPDATE mutation_attempts … WHERE id = ? AND state = 'IN_DOUBT' AND kind = 'CANCEL'` + `attempt_transitions` 행 + (FAILED_CONFIRMED 면) 기존 `releaseReservationsForAttempt(ctx, tx, …)` 를 직접 씀. 기존 함수 본문 편집 0. 쓰기 모양은 `transition` 을 옮겨 적는 것이라 「옮겨 적은 코드는 양쪽을 다 못 박는다」 — 같은 attempt 를 `ResolveConfirmed`/`ResolveFailed` 로 닫은 행과 새 함수로 닫은 행이 상태·settled_at 설정·broker_order_id·전이 행·예약 해제에서 같음을 단언하는 시험을 둔다. 선례: `ConfirmReplacement`(lineage.go:140-180)가 같은 방식으로 전이 쓰기를 자기 트랜잭션에 갖고 있음.
  - (나) 새 판독 함수 + 기존 `Resume` → `ResolveConfirmed/Failed`(IN_DOUBT CAS) — 기존 함수 편집 0, 판독과 쓰기 사이 창이 생김(종결 증거는 되돌아가지 않으므로 §2 실해는 없음). Manager 지시(한 트랜잭션)와 다름.
  - (다) `transition` 을 tx 를 받는 내부 함수로 떼어 재사용 — 기존 High-risk 함수 편집 1(FLM 번들 필요).
- **채택안(설계 반영): (가).** 질문 Q1 에 (나)를 대안으로 남김.
- 술어 공유: D10 은 종결 행의 `state`·`fail_closed` 를 읽어야 하므로 `EXISTS(...)` 상수를 그대로 쓸 수 없다. 같은 정의를 지키는 방법: `confirmedOrderTerminalEvidence` 를 「일치 조건 문자열 상수 + `EXISTS (SELECT 1 FROM all_fill_snapshots f WHERE ` + 그것 + `)`」 로 **상수 선언만** 나누고(함수 본문 무편집 — gate ⑤ 대상 아님), D10 은 그 일치 조건 상수를 써서 `SELECT f.state, f.fail_closed FROM all_fill_snapshots f WHERE <같은 조건>` 을 읽는다. 결합 문자열이 편집 전 상수와 바이트 동일함을 시험으로 단언(편집 전 문자열은 base 에서 떠서 시험 안 리터럴로).

## 4. (d) `inDoubtSettle.sweep` 와 뒤집히는 a193 시험

- 분기(AST): B1@163 원장 읽기 실패 → B2@164 ctx 면 반환, 아니면 `noteUnreadable` · B3@173 순회 · **B4@175 `IN_DOUBT && PLACE` → targets(D10 편집 자리)** · B5@181 1단 알림(처음 봄·10분) · B6@186 2단 직렬 해소 · B7@187 해소 오류(ctx) 반환.
- D10 편집: B4 에 갈래 하나 — `IN_DOUBT && CANCEL` 이면 별도 목록. 그 목록은 **1단 알림에 넣지 않고**(a193 알림은 PLACE 만 — 처음 봄·10분·park 통지 무변), **2단 앞**에서 원장 함수로 닫는다(2단의 PLACE 해소는 분 단위라 그 뒤에 두면 닫힘이 밀림). 원장 함수 오류: ctx 취소면 반환, 아니면 `lastFailure` 와 같은 방식으로 분류만 한 번 로그(원문 없음), 다음 스윕에서 다시.
- 원장 면: `inDoubtLedger` 인터페이스(in_doubt_settle.go:42-46)에 새 메서드 하나. 운영 구현 `*journal.Journal`, 시험 가짜 `a193Ledger` 는 `*journal.Journal` 을 내장하므로 무편집으로 승격. 인터페이스 주석 「전부 기존 질의(새 질의 없음)」 은 고쳐야 함.
- 구조 고정 `TestA193TheSettlerNeverReplays`(761-777): in_doubt_settle.go 의 선택자 이름이 `ReplayInDoubt`·`Gateway`·`Place`·**`Cancel`**·`Amend`·`Submit`·`RecoverPending` 과 **정확히 같으면** 실패 — 새 원장 메서드 이름은 그 일곱과 달라야 함(예: `CloseInDoubtCancelOnTargetClose` 는 통과 — 비교는 전체 이름 일치). `journal.KindCancel` 은 이름이 `KindCancel` 이라 통과.
- **뒤집히는 a193 시험: 0.** `TestA193OnlyInDoubtPlaceAttemptsAreTouched`(719-745)의 IN_DOUBT CANCEL 행은 대상 `ord-synthetic-target` 의 소유 PLACE·스냅숏이 없어 D10 이 닫지 않음 → B4@738 상태 무변 통과, Resolve 호출 0·알림 0 그대로. 다른 a193 시험은 CANCEL 을 만들지 않음(grep: `KindCancel` 은 :725 하나). 시험 이름은 「PLACE 만 해소 대상」으로 여전히 참(D10 은 해소기를 부르지 않음).
- `TestA193NoInDoubtMeansNoBrokerCall` 등 브로커 0 단언: D10 은 원장만 — 무영향.

## 5. (e) `resolveCancel` 의 상태 → 판정(분기 id)

| 대상 파생 상태 | 분기 | 결과 | 사유 |
|---|---|---|---|
| `CANCELLED` · `CANCELLED_PARTIALLY_FILLED` | B7@81 | `ResolveConfirmed(대상 번호)` → CONFIRMED | `in_doubt_resolved_found` |
| `FILLED` | B9@92 | `ResolveFailed` → FAILED_CONFIRMED | `in_doubt_resolved_absent` |
| `REJECTED` · `REJECTED_PARTIALLY_FILLED` | B11@103 | `ResolveFailed` → FAILED_CONFIRMED | `in_doubt_resolved_absent` |
| `CANCEL_PENDING` · `REPLACE_PENDING` | B13@116 | 기다림(예산 끝나면 park) | — |
| `OPEN_UNFILLED` · `OPEN_PARTIALLY_FILLED` | B15@127 | 안정 관찰 N회·최소 시간 뒤 FAILED_CONFIRMED, 아니면 기다림 | `in_doubt_resolved_absent` |
| `REPLACED` | B19@145 | park(UNRESOLVED) | — |
| 그 밖(`UNKNOWN_BROKER_STATE`) | B20@153 | park | — |
| 읽기 실패 | B4@71 | 기다림·예산 끝 park | — |

D10 표 = 위의 terminal 행 중 **B7·B9·B11 셋**만(CANCELLED* → CONFIRMED(대상 번호), FILLED·REJECTED* → FAILED_CONFIRMED). REPLACED·비종결·UNKNOWN·fail_closed·소유 모호 → 아무것도 하지 않음(IN_DOUBT 유지 — 기동이 이어받음). 「OPEN 안정 관찰 → FAILED」 는 브로커 관찰이 근거라 D10 은 하지 않음.

## 6. (f) 수동 flatten 과 엔진의 미종결 취소

- flatten 프로세스는 기동 복구(`Recovery.Run`)를 돌리지 않는다(cmd/tossctl/flatten.go — 자체 `engine.NewOrderPath` 배선, reconcile 호출 0). 자기 Resolver 는 **자기 단계의 attempt**만 해소(`resolveStep` — `step.AttemptID`).
- flatten 의 게이트웨이도 `submit` → `checkSymbolFree`(gateway.go:529) → `unsettledFor` 로 같은 원장의 `PendingAttempts` 를 본다 → **엔진의 IN_DOUBT·ACKED 취소가 있는 종목에서 flatten 의 취소는 `SymbolInFlight` 거절 → `classifyCancel`(상태 없음이라 switch B1 의 어느 case 도 아님 → B6@467 `err != nil`) → `FlattenStepHeld`(청산에서 보류), 그 종목 매도도 같은 거절 → `sell` 의 default → `FlattenStepFailed`.** 엔진 PLACE 의 미종결도 종전부터 같았지만, 엔진이 **취소**의 미종결을 만들기 시작한 것은 a198 이다(편집 전에는 취소가 원장에서 거절돼 attempt 가 없었음) — a198 이 여는 flatten 영향(비상 청산의 즉시성, 불변식 4). design §2-6 문장 「같은 종목 미종결 가드가 손절·flatten 을 막고」 가 이미 언급했으나 잔여 목록에 이름이 없었다.
- D9 의 영향: 없음(flatten 은 기동 복구를 안 돌림). 엔진을 다시 띄우면 그 기동이 ACKED 취소를 닫아 flatten 이 풀린다.
- D10 의 영향: 엔진이 **돌고 있을 때만**, 대상 주문의 종결 기록이 들어오면 다음 스윕에 그 취소를 닫아 **게이트웨이 차단**이 풀린다. **정정(보강 2 freeze 리뷰 r1 #6)**: 이미 `SymbolInFlight` 로 HELD 가 된 flatten 단계는 청산 계획이 attempt 가 아니라 `flatten_steps` 의 HELD·PENDING·IN_DOUBT 를 읽으므로(journal/flatten.go:324) 같은 saga 에서는 보류가 남는다 — flatten 을 다시 실행해야 단계가 재평가된다. 또 「다음 스윕」 은 앞 스윕의 직렬 PLACE 해소(건당 예산 5 분) 뒤다(리뷰 #3). 엔진이 멈춰 있으면 영향 없음(그 종목은 엔진 기동 전까지 flatten 에서 보류 — 잔여로 이름 붙임).
- D10 이 flatten 이 만든 IN_DOUBT 취소를 닫는 경우: 원장은 공유이므로 대상이 엔진 소유 주문이고 종결 기록이 있으면 닫힌다. 외부 주문이 대상이면 소유 후보 0 → 닫지 않음 — **단 정정(리뷰 r1 #2)**: 외부 주문이 예전 엔진 주문의 번호를 재사용하면 예전 소유 intent 가 유일 후보가 되고 그 예전 종결 기록으로 닫힐 수 있다(design Q7 — 「종결 행의 기록 시각이 취소 attempt 의 기록 시각보다 뒤」 결속 추천, Manager 결정 대기). flatten 의 단계 행(`flatten_steps`)은 그대로 InDoubt 로 남고, 다음 flatten 실행의 `resolveStep` → `Resolve` 는 종결된 attempt 를 그대로 돌려줌(indoubt.go:253 멱등) → CONFIRMED 면 Done, FAILED_CONFIRMED 면 Pending(재취소 — 대상은 이미 종결이라 브로커가 409 already-* → flatten 정책은 모호 → resolver). 같은 시각 flatten 의 `resolveCancel` 과 D10 이 같은 attempt 를 닫으려 하면 상태 조건 쓰기(`WHERE state = 'IN_DOUBT'`)로 하나만 이김 — 진 쪽 flatten 은 오류 → 그 단계 InDoubt(보류, 안전측).

## 7. 「이전에도 보류였다」 두 문장 — base 코드 확인

base `clearTheSymbol`(git show e1eaf720c, base ast `bc81b5fd`): 미체결 목록의 각 주문에 대해 매도는 `ConfirmedCancelOf` 대기 판독 뒤 `!withPending` 이면 continue(보류 없음), 그 밖(매수 전부·withPending 의 매도)은 `o.opts.Issuer.IssueReduction`(매도·PLACE 결정) → `Submit.Cancel` 이 원장에서 거절(이 change 가 고친 결함) → `err != nil` → `res.cleared = false`. 즉 **취소를 시도하는 주문이 원장 미체결 목록(`LiveOrdersForSymbol`)에 남아 있는 동안 base 는 언제나 보류**, 목록에서 빠지면(종결 기록 — 술어 §2) 치움.

1. 「`already-modified` 로 IN_DOUBT: 이 change 이전에도 그 주문이 원장에서 살아 있는 한 손절은 보류됐다」 — **참.** 앱 정정은 후계가 원장에 없어 그 주문의 파생이 `replaced_without_successor` → UNKNOWN(terminal 0)이고 목록에 남는다 → base 도 보류. 단 차이 하나: 그 주문이 뒤에 종결로 바뀌면 base 는 그 주기에 풀렸고, D10 뒤에는 다음 스윕이 그 IN_DOUBT 취소를 닫은 뒤 풀린다(지연 ≤ 스윕 한 번). 종결이 REPLACED(후계 있음)이면 base 는 풀렸고 D10 은 닫지 않는다 — §2 대로 도달 거의 불가, 잔여로 이름 붙임.
2. 「대상이 살아 있는데 취소가 5xx 로 IN_DOUBT: 이전에도 그 주문이 살아 있는 동안은 보류」 — **참**(같은 근거). `!withPending` 매도는 base·이후 모두 취소를 내지 않아 IN_DOUBT 가 생기지 않는다.

## 8. 편집 표면(예상) — 고칠 기존 함수·뒤집히는 시험·census/핀

| 종류 | 대상 | 비고 |
|---|---|---|
| 기존 production 함수 편집 | `Recovery.confirmAcked`(acked_boot.go — B2 앞에 CANCEL 갈래) · `inDoubtSettle.sweep`(in_doubt_settle.go — B4 갈래 + 2단 앞 닫기 루프) | 둘 다 FLM 번들 필요(gate ⑤). High-risk(원장·해소) |
| 기존 선언 편집(함수 아님) | `inDoubtLedger` 인터페이스(메서드 1 추가·주석) · `confirmedOrderTerminalEvidence` 상수(일치 조건 상수로 분리 — 결합 문자열 바이트 동일) · `reconcile` 사유 상수 1 추가 · 파일 머리 주석(acked_boot.go·in_doubt_settle.go 의 「CANCEL 은 대상 밖」 문장) | gate ⑤ 대상 아님 |
| 새 파일·새 함수 | journal: D10 판독+종결 한 트랜잭션 함수 · execgw 또는 journal: 판정 표 순수 함수 · engine: 스윕의 취소 닫기 헬퍼(선택) | 새 코드는 새 파일에(기억: 기존 파일 속 새 함수도 Logic Map 에 잡힘) |
| 뒤집히는 기존 시험 | `TestA094AnAckedCancelOrNumberlessPlaceIsOnlyNamed`(시험 함수 본문 편집 → FLM 번들 필요) | 1 개 |
| 무영향 확인 | `TestA094BootAndPostPlaceShareOneConfirmation` · a193 전부(§4) · a094_exit_clear · a165 latch(`isCancelRefusalLatch` 는 a165 꼬리표가 있는 FAILED_CONFIRMED 만, 그리고 대상이 살아 있을 때만 — D10 의 FAILED_CONFIRMED 는 꼬리표 없음·대상 종결) | |
| census·핀 | a166 `testdata/a166-wiring/shapes.json`·`sites.json` 에 acked_boot · in_doubt_settle · confirmAcked · sweep · exit_proposal_release · amend_indoubt **0 건**(`sweep` 문자열 1 건은 cmd 의 다른 함수 본문 텍스트) · a163 핀 파일(official·output·config·engine judgement)에 대상 파일 0 · a087 번들 파일 무관 | **census 고정 함수 편집 없음 — 멈출 사유 없음** |
| 예약 해제 | FAILED_CONFIRMED 는 같은 트랜잭션에서 그 attempt 의 예약을 푼다(transition L654) — 취소 결정은 위험 감소라 예약이 없음(예상, 시험으로 0 단언) | |
