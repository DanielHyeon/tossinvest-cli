# a200 design — 수렴 조정이 닫은 포지션에 닿은 엔진 청산 체결의 인정

용어: **인정** = 닫힌 인스턴스에 온 엔진 매도 체결을 거절하지 않고, 그 인스턴스를 닫은 수렴 조정이 먼저 반영한 수량 감소의 설명으로 받아들이는 것. 수량은 다시 빼지 않는다.
판 이력: 1단계 재현(2026-11-24) 뒤 Manager 결정으로 성과(옛 D5) · 운영 로그(옛 D7) · 수량 저장(옛 D3)을 뺐다. codex r1 · r2 의 P1 넷이 「체결 누적의 출처」를 지목해 (라)를 **원장 누적을 읽지 않는 정적 수량 상한**으로 바꿨다(Manager 결정 3, 2026-11-24). 근거 `analysis/claim-check.md`.

## 1. 결정

| # | 결정 | 근거 · 기각한 대안 |
|---|---|---|
| D1 | **판정은 체결 tx 안, 거절이 정해진 그 자리 한 곳.** `ProjectPosition` B5(if@101)에서 `enterReconcileInTx`@102 를 부르기 직전에 인정 판정을 한 번 부른다. 인정이면 차단 쓰기를 건너뛰고 표지를 쓴 뒤 nil. 전이표(table.go X58~X60) · `position.Apply` 무편집 | 원장 연결 1개(journal.go:174) + `BEGIN IMMEDIATE` 라 체결 tx 와 조정 tx 는 직렬 — CLOSED 를 읽었다는 것이 「조정이 먼저 커밋됐다」의 증명이고 시각 비교가 필요 없다(재현: 같은 초). 전이표를 고치면 조건 없는 완화 |
| D2 | **인정 조건 넷 — 전부 성립해야 한다(닫힌 집합).** (가) `outcome.Refusal == SELL_ON_CLOSED`. (나) 그 인스턴스의 조정 행 중 `new_quantity` 가 decimal 0 인 행이 **정확히 하나**이고 그것이 그 인스턴스의 마지막 조정 행이며 kind 가 수렴이 쓰는 값(`position.Classify(LocalInstance)` = `UNKNOWN`)이다. (다) `origin.IntentID`(resolveFillOrigin 결과, 다시 유도하지 않음)가 비어 있지 않고 `ApplyTx.PendingState(인스턴스)` 의 `PendingIntentID` 와 같다(exit state 없음 · 비어 있음 · 다름 → 아님). (라) **수량 상한**: 그 intent 의 `mutation_attempts` 중 kind 가 PLACE 또는 AMEND 인 행이 (상태 무관) **정확히 하나**이고 그것이 CONFIRMED PLACE 이며, 그 intent 의 `intents.quantity` ≤ 그 조정의 `prev_quantity − new_quantity` | (나): CLOSED 는 종결이라(전이표 · 수렴은 CLOSED 를 조정 안 함 · EXTERNAL 접기는 새 인스턴스) 0 조정 행 하나 ⟺ 조정이 닫음 — 조정 행과 체결 행의 순서를 id 로 비교할 필요가 없다. 생산에서 MANUAL 을 쓰는 코드는 0 이고 수렴은 UNKNOWN 을 쓴다 — 판정은 「수렴이 쓰는 kind」 를 그대로 요구하고 미래의 다른 kind 를 인정하지 않는다. (다): `pending_intent_id` 철자는 구조 시험 때문에 새 파일에 쓰지 않는다 |
| D2a | **(라)가 상한인 이유(전제 (i)~(iii), claim-check)** — 그 intent 의 주문은 하나(PLACE 하나 · AMEND 0)이고 그 주문 수량은 `intents.quantity` 와 같다. 체결 누적은 주문 수량을 넘지 못하므로 「그 주문의 모든 체결 ≤ 감소분」 이 원장의 체결 누적을 읽지 않고 성립한다. `fill_events` · `fill_snapshots` 는 읽지 않는다 | 누적 출처를 읽던 판은 codex r1 P1-1(float 차를 0 으로 버림) · r2 P1(호환 표 범위 충돌 · 작은 감소 덮어쓰기)로 **적게 세는** 길이 셋 나왔다 — 관측 이력의 최댓값은 원장에 보존되지 않는다. 정적 상한은 그 출처 자체를 없앤다. PLACE 가 둘 이상이면(재시도 · in-doubt 뒤 재제출로 주문이 둘일 수 있음 — attempt_no 는 intent 마다 1부터, durability.go Prepare) 주문 수량 합이 intent 수량을 넘을 수 있으므로 거절. AMEND 는 수량을 바꿀 수 있고 이 change 는 그 계보를 풀지 않으므로 거절 |
| D3 | **표지 행** — 인정 1회마다 `exit_events` 한 줄: `position_id` = 그 인스턴스, `action` = 새 상수 `ADJUSTMENT_CLOSE_FILL_RECOGNISED`(새 파일 — 기존 `ADJUSTMENT_CLOSED` · `PROPOSAL_FILLED` 명명을 따라 콘솔 이력에서 「조정으로 닫힌 뒤 온 체결을 인정」 으로 읽힘), `created_at` = tx 시각, 나머지(`proposed_intent_id` · 가격 · 레벨 · 평가 증거 열 전부)는 NULL. 기존 `appendExitEventFromHook` 를 그대로 부른다(무편집) | 소비자 전수(`analysis/edit-surface.md`)가 이 모양을 **무시함**(intent 키 소비자 · 증거 검증 · 백필 · a087 · 성과 매도 귀속) 또는 **원문 표시만**(콘솔 이력 · provenance — 판정 없음, Manager 승인). 수량은 싣지 않는다. 판정은 표지를 읽지 않는다 |
| D4 | **부분 체결은 체결마다 판정** — Working · Done 체결 각각이 넷을 통과해야 한다. (라)는 정적이라 같은 주문의 체결마다 같은 답을 낸다 | Project 시점에 `pending_intent_id` 는 살아 있다(훅 순서 Project@284 → Exit@298, Exit 훅은 종결 체결에서만 비움 — 재현) |
| D5 | **인정돼도 Exit 훅 · 성과는 오늘과 같다.** `ApplyExitFill` 무편집 — 인정된 청산은 `ADJUSTMENT_CLOSED` + (표지) + `PROPOSAL_FILLED`, `COMPLETED` 0, 성과 0 | 성과는 비목표(proposal) |
| D6 | **인정돼도 그 주기의 수량 불일치 차단은 손대지 않는다.** 대사가 세운 일시적 수량 불일치(W4 · W5 의 같은 주기)는 지금처럼 조정 크레딧으로 다음 일치 비교에서 풀린다 — **단 크레딧은 프로세스 기억이라, 그 사이 재기동하면 지금처럼 「조정 대기」 로 남고 운영자 해제가 필요하다**(codex r1 P2-2, `Tracker.Restore` mismatch.go:998). 인정 실패 체결도 그 종목에 활성 행이 있으면 AF 를 안 쓴다(F10, enterReconcileScopeInTx :401) — 오늘 동작 그대로 | 인정 뒤 Observe 시점 원장 = W6 재현의 상태(AF 행 없음 + 조정 · 크레딧) — W6 에서 첫 일치에 `ADJUSTMENT_APPLIED` |
| D7 | **스키마 36 유지.** 새 열 · 새 표 · 새 cause · 새 release cause 없음 | 롤백이 「백업 복원 2단계」가 되는 것을 피함(a192) |

## 2. fail-closed 경계

**새로 통과시키는 것**: D2 넷을 만족하는 엔진 매도 체결 — 차단 없음, 수량 0 유지, 표지 1행.
그것이 틀렸을 때의 귀결: 수량은 조정이 이미 계좌 값(0)에 맞췄고 이후 비교가 계속 계좌와 대조하므로 인정이 수량을 틀리게 만들 수 없다. 틀릴 수 있는 것은 귀속(이 감소가 엔진 것인가)과 차단 유무뿐이고 (다)(라)가 묶는다.

**인과는 증명하지 않고 받아들인다(codex r1 P1-2, Manager 결정 2026-11-24 (A))**: D2 는 「그 주문이 이 인스턴스의 청산 발의다」 를 증명할 뿐 「이 체결이 그 조정이 본 감소를 일으켰다」 를 증명하지 않는다(원장에 조정 스냅숏과 체결의 인과를 가를 열이 없음). 반례는 「엔진 매도가 살아 있는 채 외부 처분으로 계좌가 0 → 조정 → 재매수가 원장에 접히기 전에 옛 엔진 매도 체결」 이다.
1. 인정이 요구하는 것은 「조정이 그 인스턴스를 **0 으로** 만들었다」 와 「체결된 주문이 그 인스턴스의 살아 있는 청산 발의다」 의 동시 성립이다. 계좌 수량이 0 으로 읽히려면 그 매도 주문이 묶은 수량도 계좌에서 사라져야 하고, 미체결 매도가 묶은 수량은 그 주문이 체결되거나 취소되기 전에는 다른 처분으로 나가지 않는다 — 반례는 그 주문의 체결 없이는 성립하기 어렵다. 뒷받침: **미체결 지정가 매도는 매도가능수량을 예약한다 — US M29(115 → 114), KR M33(6 → 5)**, 보유 초과 매도는 `422 insufficient-sellable-quantity`(M30 · M33) — `openspec/changes/verify-execution-capability/measurements.md`. 엔진 청산은 지정가 매도다(`sellIntent` exitloop.go:1839 `OrderType: "limit"`). 조건주문은 예약하지 않지만(M13 · M34) 엔진 청산은 조건주문이 아니다.
   **미측정으로 남는 것**(브로커의 통상 동작에 대한 가정, Story limits): ① 측정은 Open API 주문이고, 사람의 앱 매도가 같은 매도가능수량 규칙을 따른다는 것과 「부분 체결 뒤 잔량」 의 예약은 측정하지 않았다. ② 보유 조회와 주문 서비스 사이의 일관성 — 주문이 살아 있는데 보유가 0 으로 읽히는 지연 응답(codex r2 P2). 대사의 안정화(2 회 동일 스냅숏)가 그 창을 좁히지만 없애지는 않는다.
2. 틀렸을 때의 귀결: 수량은 조정이 이미 계좌 값에 맞췄고 이후 비교가 계속 대조하므로 틀어지지 않는다. 잃는 것은 「그 종목의 차단 하나」 이고, 그 경우에 일어난 일(엔진의 낡은 매도가 사람의 새 보유에서 체결)은 표지 행과 체결 기록에 남는다.
3. 기각: (B) 「주문 접수가 조정보다 앞섬」 은 반례를 막지 못한다(그 주문은 조정보다 먼저 있었다). (C) 인정 포기는 재현된 결함을 그대로 둔다.

**계속 거절해야 하는 입력**(각각 시험, 결과는 오늘과 같은 `ATTRIBUTION_FAILED` · 문구 무변):
1. 다른 거절 유형 둘: `UNATTRIBUTED_SELL` · `ENTRY_WHILE_CLOSING`(그리고 `OVERSELL` 은 원래 QM).
2. 체결로 닫힌 인스턴스에 온 매도(0 조정 행 없음 — 재현 「둘째 엔진 매도」).
3. `pending_intent_id` 가 비었거나 · 다른 intent 거나 · exit state 가 없는 인스턴스(무관리). **정정 여부와 무관하게 그 주문(또는 그 부모)의 종결 관측이 pending 을 비운 뒤에 온 체결**(codex r2 P2 — `ApplyExitFill` B11 if@1009 는 종결 관측이면 successor 를 보지 않고 비움).
4. 그 intent 의 주문 수량이 조정 감소분보다 큰 체결 — 아래 「거절하는 정상 입력 (가)」.
5. 그 intent 에 PLACE · AMEND attempt 가 둘 이상(정정 · 재시도) — 아래 (나).
6. 0 조정의 kind 가 수렴이 쓰는 값이 아닌 것(지어낸 MANUAL 행으로 시험 — 생산 쓰기 0).

**거절하면 안 되는 정상 입력**(각각 시험):
1. 조정과 체결이 같은 초.
2. 조정 뒤 부분 체결 여러 번(Working 뒤 Done), 주문 수량 = 감소분.
3. 같은 대사 비교에 그 주문이 `MissingOrders` 로도 잡힌 경우.
4. 수량 불일치 차단이 이미 선 창(W6) — 같은 판정(차단은 원래 안 씀, 표지만 남음).
5. 주문 수량 < 감소분(부분 익절 주문 — 사다리 · 래칫 부분).

**이 change 가 거절하는 정상 입력(대가, 이름으로)**:
- (가) **그 주문의 체결 일부가 조정보다 먼저 투영에 반영된 경합** — 조정이 그만큼 작은 보유(31 → 21)를 닫았으므로 주문 수량(31) > 감소분(21) 이고, **조정 뒤 첫 체결부터** `ATTRIBUTION_FAILED` 다. 정확히 가르려면 「어느 체결이 조정 뒤였는가」 의 원장 기록이 필요한데 그 열이 없다(스키마 36). 재현 `TestA200CurrentBehaviourAStopFillStraddlingTheConvergenceIsAnAttributionFailure`. γ 를 다시 열 첫 후보.
- (나) **정정된 청산 주문의 체결**(그 intent 에 AMEND attempt 가 생긴 뒤에 온 체결 전부 — 정정 전에 인정된 체결은 그대로). 지금 exit 관측은 청산을 정정하지 않는다. a167(보호 잔량 재호가, 착수 전)이 착지하면 이 범위가 넓어진다.

**범위 밖 사실(판정에 닿지 않음)**: 체결 투영과 Exit 훅은 (계좌 · 시장 · 종목)의 최신 인스턴스를 읽는다(`currentInstance` · `exitContextForFill`). 조정 뒤 재매수로 새 인스턴스가 생긴 다음에 온 옛 주문의 체결은 `SELL_ON_CLOSED` 에 오지 않으므로 인정 판정에 닿지 않는다 — 전량이면 OVERSELL → 수량 불일치, 부분이면 차단 없이 새 인스턴스를 줄인다(재현, `analysis/repro/out-of-scope.md`, 별도 Story 후보).

## 3. 시험 (RED 의 뼈대)

실제 원장(임시 DB) · 운영 훅 리터럴로, 조정 tx 를 먼저 커밋한 뒤 체결을 적용한다. 1단계 재현 시험이 첫 줄이다.
1. **본 사고 모양**(전량, 같은 초): 차단 행 0 · 투영 CLOSED/0 · 표지 1 · `ADJUSTMENT_CLOSED` 와 `PROPOSAL_FILLED` · `COMPLETED` 0 · 성과 0. 재현 시험 `…AConvergenceCloseThenTheStopFillIsAnAttributionFailure` 를 이 기대로 뒤집는다(RED).
2. **계속 거절** — §2 의 1~6 각각: 활성 `ATTRIBUTION_FAILED` 1 · 표지 0 · 기존 evidence 문구 무변. 4 는 재현 시험 `…StraddlingTheConvergence…` 가 그대로 PASS. 3 의 종결 관측 뒤 체결 · 5 의 정정 계보(AMEND 가 접수 확정된 **뒤**에 온 체결은 부모 · 자식 모두 거절 — 정정 전에 이미 인정된 부모 체결은 그대로 남음) · 5 의 PLACE 둘 각각.
3. **정상 입력** — §2 의 1~5 각각: 차단 행 0 · 표지(인정된 체결 수만큼).
4. **실제 `Tracker`**(1단계 `…_cycles_test.go` 의 하네스): W4 · W5 는 그 주기 귀속 실패 행 0, W5 의 Observe 오류 없음, 다음 일치 비교에서 수량 불일치가 `ADJUSTMENT_APPLIED` 로 풀리고 그 종목 `EntryAllowed` = nil(편입 경로가 그 종목을 차단 보유로 거르지 않음). W6 은 무변(오늘과 같은 결과 + 표지). 재기동 판: W4 인정 뒤 Tracker 를 새로 만들어 `Restore` 하면 수량 불일치가 「조정 대기」 로 남는 것(오늘 동작 — 공시 시험, codex r1 P2-2).
5. **구조**: 인정 판정 함수의 생산 호출 자리 1(`ProjectPosition`), 새 파일에 `pending_intent_id` 철자 0(기존 구조 시험), 판정이 `resolveFillOrigin` 을 부르지 않음, `fill_events` · `fill_snapshots` 를 읽지 않음.
6. **변이**(사본): (가) · (나) · (다) · (라) 각각 삭제, (나)의 「정확히 하나 · 마지막」 삭제, kind 검사 삭제, 상한 비교를 `<` 로(같은 수량인 본 사고가 거절돼야 잡힘), 상한 비교 삭제, AMEND 검사 삭제, PLACE 유일 검사 삭제, 감소분 대신 `new_quantity` 만(또는 `prev_quantity` 대신 다른 열), 인정 뒤 투영 쓰기(수량 다시 빼기), 판정을 거절 유형 무관하게 부르기.

## 4. 잔여 · 공시

- 성과(F15) 없음 — 인정된 손절도 성과에 잡히지 않는다. 후속 Story 후보.
- 거절하는 정상 입력 (가) 조정을 사이에 둔 부분 체결 · (나) 정정된 청산 주문(§2). a167 착지 뒤 (나)가 넓어진다.
- 인과는 증명하지 않는다(§2 「인과는 …」) — 뒷받침은 M29 · M33 · M30, 앱 매도 · 부분 체결 잔량 · 보유 조회 지연은 미측정 가정. 틀리면 그 종목의 차단 하나를 잃고 표지 · 체결 기록에 남는다.
- 재기동이 끼면 그 주기의 수량 불일치는 지금처럼 조정 대기(D6).
- 재매수 뒤 옛 주문의 체결은 판정에 닿지 않는다(§2, 별도 Story 후보).
- W2 · W3(체결 먼저)의 수량 불일치 고착은 재현됨, 이 change 가 다루지 않는다(별도 Story 후보). 부분 조정 뒤 초과 체결은 크레딧으로 풀림(`analysis/repro/out-of-scope.md`).
- a137 일일 손실 측정 불가는 인정 뒤에도 그대로.
- 인정된 손절에도 「외부에서 청산됨」 알림이 먼저 간다(수렴이 닫는 순간에 보내므로). 정정 알림 · 운영 로그 없음.
- flatten · 운영자 매도와 겹친 경합은 여전히 `ATTRIBUTION_FAILED`(결속 없음) — a194 가 알리고 a195 로 푼다.
- γ 보류(proposal).

## 5. 순서 · 결합

- 편집할 기존 함수는 **하나**: `ProjectPosition`(B5 1 자리). 판정 · 표지 쓰기는 새 파일. 그 밖으로 늘면 멈추고 보고.
- 고정 장치: `TestGuardedExitColumnsAreWrittenOnlyByTheApplyHook`(새 파일은 `PendingState` 로 읽음), `adoptionIDReaders` · `eligibilitySpellers`(새 파일은 그 열을 안 읽음), 전이표 census · `ValidReconcileCause` · `TradeOutcomeRefusalReasons`(새 값 없음 — 무적중), a166 census(배선 무편집 — 무적중), a130 `bindApplyHooks` 인자 수(무편집).
- a167(착수 전)은 `ApplyExitFill` 을 FLM 대상으로 적었으나 이 change 는 `ApplyExitFill` 무편집 — 편집 겹침 없음. 다만 a167 이 청산을 정정하면 그 체결은 이 인정의 (나)로 거절된다. a199 와 편집 파일 겹침은 Manager 대조.
- base 재고정은 a198 착지 · merge 뒤(Manager 승인 2026-11-24, `review.md`).
