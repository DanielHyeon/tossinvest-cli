# a200 design — 수렴 조정이 닫은 포지션에 닿은 엔진 청산 체결의 인정

용어: **인정** = 닫힌 인스턴스에 온 엔진 매도 체결을 거절하지 않고, 그 인스턴스를 닫은 수렴 조정이 먼저 반영한 수량 감소의 설명으로 받아들이는 것. 수량은 다시 빼지 않는다.
판 이력: 1단계 재현(2026-11-24) 뒤 Manager 결정으로 성과(옛 D5) · 운영 로그(옛 D7) · 수량 저장(옛 D3)을 뺐다. 근거 `analysis/claim-check.md`.

## 1. 결정

| # | 결정 | 근거 · 기각한 대안 |
|---|---|---|
| D1 | **판정은 체결 tx 안, 거절이 정해진 그 자리 한 곳.** `ProjectPosition` B5(if@101)에서 `enterReconcileInTx`@102 를 부르기 직전에 인정 판정을 한 번 부른다. 인정이면 차단 쓰기를 건너뛰고 표지를 쓴 뒤 nil. 전이표(table.go X58~X60) · `position.Apply` 무편집 | 원장 연결 1개(journal.go:174) + `BEGIN IMMEDIATE` 라 체결 tx 와 조정 tx 는 직렬 — CLOSED 를 읽었다는 것이 「조정이 먼저 커밋됐다」의 증명이고 시각 비교가 필요 없다(재현: 같은 초). 전이표를 고치면 조건 없는 완화 |
| D2 | **인정 조건 넷 — 전부 성립해야 한다(닫힌 집합).** (가) `outcome.Refusal == SELL_ON_CLOSED`. (나) 그 인스턴스의 조정 행 중 `new_quantity` 가 decimal 0 인 행이 **정확히 하나**이고 그것이 그 인스턴스의 마지막 조정 행이며 kind 가 수렴이 쓰는 값(`position.Classify(LocalInstance)` = `UNKNOWN`)이다. (다) `origin.IntentID`(resolveFillOrigin 결과, 다시 유도하지 않음)가 비어 있지 않고 `ApplyTx.PendingState(인스턴스)` 의 `PendingIntentID` 와 같다(exit state 없음 · 비어 있음 · 다름 → 아님). (라) 「그 intent 의 주문들의 누적 체결 합(이 체결 포함)」 ≤ 그 조정의 `prev_quantity − new_quantity` | (나): CLOSED 는 종결이라(전이표 · 수렴은 CLOSED 를 조정 안 함 · EXTERNAL 접기는 새 인스턴스) 0 조정 행 하나 ⟺ 조정이 닫음 — 조정 행과 체결 행의 순서를 id 로 비교할 필요가 없다. 생산에서 MANUAL 을 쓰는 코드는 0 이고 수렴은 UNKNOWN 을 쓴다 — 판정은 「수렴이 쓰는 kind」 를 그대로 요구하고 미래의 다른 kind 를 인정하지 않는다. (다): `pending_intent_id` 철자는 구조 시험 때문에 새 파일에 쓰지 않는다 |
| D2a | **누적 합의 출처(codex r1 P1-1 뒤 개정)** — 주문마다 **브로커 원문 누적**을 decimal 로 더한다: 이 체결의 주문은 `fill.CumulativeQuantity`(이 관측의 원문), 그 밖의 주문은 `fill_snapshots.filled_quantity`(그 주문 번호의 스냅숏 행들 중 가장 큰 값). 대상 주문은 `mutation_attempts` 에서 `intent_id = origin.IntentID ∧ state = CONFIRMED ∧ kind ∈ {PLACE, AMEND} ∧ broker_order_id <> ''` 인 번호 전부. 범위(계좌 · 시장 · 거래일 · 방향) 필터를 걸지 않는다. `fill_events.delta_quantity` 합은 쓰지 않는다 | `delta_quantity` 는 float64 차에서 나오고 `nearlyZero`(fills.go:1924, 상대 1e-9)가 작은 차를 0 으로 만들어 행을 안 남기는데 스냅숏은 전진한다 — 그 합은 실제 누적보다 **작을 수 있다**(codex r1 재현 계산, 대수량 · 소수 수량 경계). 원문 누적은 그 손실이 없다. 범위 필터를 빼고 스냅숏 행 중 최댓값을 쓰면 남는 오차는 **더 많이 세는 쪽**뿐(같은 번호를 다른 범위가 재사용 · 정정 자식의 누적이 부모를 포함)이라 인정이 줄어드는 방향(fail-closed). 정정 자식이 다른 intent 로 기록되면 (다)에서 걸림. 재적용: 바이트 같은 재관측은 훅 전에 반환, delta 0 관측은 CLOSED 이동 없음이라 판정 안 부름 |
| D3 | **표지 행** — 인정 1회마다 `exit_events` 한 줄: `position_id` = 그 인스턴스, `action` = 새 상수(새 파일), `created_at` = tx 시각, 나머지(`proposed_intent_id` · 가격 · 레벨 · 평가 증거 열 전부)는 NULL. 기존 `appendExitEventFromHook` 를 그대로 부른다(무편집) | 소비자 전수(`analysis/edit-surface.md`)가 이 모양을 **무시함**(intent 키 소비자 · 증거 검증 · 백필 · a087 · 성과 매도 귀속) 또는 **원문 표시만**(콘솔 이력 · provenance — 판정 없음). 수량은 싣지 않는다: 수량 모양 열은 평가 증거 묶음뿐이고 단독으로 채우면 `ExitEvents` 가 corrupt. (라)는 상태 없이 원장에서 다시 계산하므로 표지를 읽지 않는다 |
| D4 | **부분 체결은 체결마다 판정** — Working · Done 체결 각각이 넷을 통과해야 하고, 누적이 감소분을 넘는 순간의 체결부터 거절(지금처럼 `ATTRIBUTION_FAILED`, 활성 행이 있으면 F10 대로 안 씀) | Project 시점에 `pending_intent_id` 는 살아 있다(훅 순서 Project@284 → Exit@298, Exit 훅은 종결 체결에서만 비움 — 재현) |
| D5 | **인정돼도 Exit 훅 · 성과는 오늘과 같다.** `ApplyExitFill` 무편집 — 인정된 청산은 `ADJUSTMENT_CLOSED` + (표지) + `PROPOSAL_FILLED`, `COMPLETED` 0, 성과 0 | 성과는 비목표(proposal) |
| D6 | **인정돼도 그 주기의 수량 불일치 차단은 손대지 않는다.** 대사가 세운 일시적 수량 불일치(W4 · W5 의 같은 주기)는 지금처럼 조정 크레딧으로 다음 일치 비교에서 풀린다 — **단 크레딧은 프로세스 기억이라, 그 사이 재기동하면 지금처럼 「조정 대기」 로 남고 운영자 해제가 필요하다**(codex r1 P2-2, `Tracker.Restore` mismatch.go:998). 인정 실패 체결도 그 종목에 활성 행이 있으면 AF 를 안 쓴다(F10, enterReconcileScopeInTx :401) — 오늘 동작 그대로 | 인정 뒤 Observe 시점 원장 = W6 재현의 상태(AF 행 없음 + 조정 · 크레딧) — W6 에서 첫 일치에 `ADJUSTMENT_APPLIED` |
| D7 | **스키마 36 유지.** 새 열 · 새 표 · 새 cause · 새 release cause 없음 | 롤백이 「백업 복원 2단계」가 되는 것을 피함(a192) |

## 2. fail-closed 경계

**새로 통과시키는 것**: D2 넷을 만족하는 엔진 매도 체결 — 차단 없음, 수량 0 유지, 표지 1행.
그것이 틀렸을 때의 귀결: 수량은 조정이 이미 계좌 값(0)에 맞췄고 이후 비교가 계속 계좌와 대조하므로 인정이 수량을 틀리게 만들 수 없다. 틀릴 수 있는 것은 귀속(이 감소가 엔진 것인가)과 차단 유무뿐이고 (다)(라)가 묶는다.

**계속 거절해야 하는 입력**(각각 시험, 결과는 오늘과 같은 `ATTRIBUTION_FAILED` · 문구 무변):
1. 다른 거절 유형 둘: `UNATTRIBUTED_SELL` · `ENTRY_WHILE_CLOSING`(그리고 `OVERSELL` 은 원래 QM).
2. 체결로 닫힌 인스턴스에 온 매도(0 조정 행 없음 — 재현 「둘째 엔진 매도」).
3. `pending_intent_id` 가 비었거나 · 다른 intent 거나 · 이미 해소된 뒤의 체결, exit state 가 없는 인스턴스(무관리).
4. 누적 합이 조정 감소분을 넘는 체결(넘는 그 체결부터).
5. **조정을 사이에 둔 부분 체결** — 아래 「이 change 가 거절하는 정상 입력」.
6. 0 조정의 kind 가 수렴이 쓰는 값이 아닌 것(지어낸 MANUAL 행으로 시험 — 생산 쓰기 0).

**거절하면 안 되는 정상 입력**(각각 시험):
1. 조정과 체결이 같은 초.
2. 조정 뒤 부분 체결 여러 번(Working 뒤 Done), 누적이 감소분과 정확히 같아지는 마지막 체결.
3. 같은 대사 비교에 그 주문이 `MissingOrders` 로도 잡힌 경우.
4. 수량 불일치 차단이 이미 선 창(W6) — 같은 판정(차단은 원래 안 씀, 표지만 남음).

**이 change 가 거절하는 정상 입력(대가, 이름으로)**: 그 주문의 체결 일부가 조정보다 먼저 투영에 반영된 경합(부분 체결이 조정을 사이에 둠)은 결국 `ATTRIBUTION_FAILED` 가 된다 — 조정 뒤 체결 중 누적 합(조정 전 체결 포함)이 감소분 이하인 앞쪽 체결은 인정되고, **누적이 감소분을 넘는 체결부터** 거절된다(codex r1 P2-1 정정). 그 주문이 끝까지 체결되면 누적 = 원주문 수량 > 감소분이라 마지막 체결은 반드시 거절된다. 누적 합이 조정 전 체결을 포함하기 때문이고, 정확히 가르려면 「어느 체결이 조정 뒤였는가」 의 원장 기록이 필요한데 그 열이 없다(스키마 36). 재현 `TestA200CurrentBehaviourAStopFillStraddlingTheConvergenceIsAnAttributionFailure`. γ 를 다시 열 첫 후보.

**범위 밖 사실(판정에 닿지 않음)**: 체결 투영과 Exit 훅은 (계좌 · 시장 · 종목)의 최신 인스턴스를 읽는다(`currentInstance` · `exitContextForFill`). 조정 뒤 재매수로 새 인스턴스가 생긴 다음에 온 옛 주문의 체결은 `SELL_ON_CLOSED` 에 오지 않으므로 인정 판정에 닿지 않는다 — 전량이면 OVERSELL → 수량 불일치, 부분이면 차단 없이 새 인스턴스를 줄인다(재현, `analysis/repro/out-of-scope.md`, 별도 Story 후보).

## 3. 시험 (RED 의 뼈대)

실제 원장(임시 DB) · 운영 훅 리터럴로, 조정 tx 를 먼저 커밋한 뒤 체결을 적용한다. 1단계 재현 시험이 첫 줄이다.
1. **본 사고 모양**(전량, 같은 초): 차단 행 0 · 투영 CLOSED/0 · 표지 1 · `ADJUSTMENT_CLOSED` 와 `PROPOSAL_FILLED` · `COMPLETED` 0 · 성과 0. 재현 시험 `…AConvergenceCloseThenTheStopFillIsAnAttributionFailure` 를 이 기대로 뒤집는다(RED).
2. **계속 거절** — §2 의 1~6 각각: 활성 `ATTRIBUTION_FAILED` 1 · 표지 0 · 기존 evidence 문구 무변. 5 는 재현 시험 `…StraddlingTheConvergence…` 가 그대로 PASS 여야 하고, 조정 뒤 체결을 둘로 나눈 판(앞쪽은 누적 ≤ 감소분이라 인정 · 넘는 체결에서 AF)을 더한다(codex r1 P2-1). 원문 누적 출처 시험: 같은 주문의 관측이 `nearlyZero` 로 delta 0 이 된 뒤의 체결에서 합이 원문 누적을 따르는지(codex r1 P1-1).
3. **정상 입력** — §2 의 1~4 각각: 차단 행 0 · 표지(인정된 체결 수만큼).
4. **실제 `Tracker`**(1단계 `…_cycles_test.go` 의 하네스): W4 · W5 는 그 주기 귀속 실패 행 0, W5 의 Observe 오류 없음, 다음 일치 비교에서 수량 불일치가 `ADJUSTMENT_APPLIED` 로 풀리고 그 종목 `EntryAllowed` = nil(편입 경로가 그 종목을 차단 보유로 거르지 않음). W6 은 무변(오늘과 같은 결과 + 표지). 재기동 판: W4 인정 뒤 Tracker 를 새로 만들어 `Restore` 하면 수량 불일치가 「조정 대기」 로 남는 것(오늘 동작 — 공시 시험, codex r1 P2-2).
5. **구조**: 인정 판정 함수의 생산 호출 자리 1(`ProjectPosition`), 새 파일에 `pending_intent_id` 철자 0(기존 구조 시험), 판정이 `resolveFillOrigin` · ownership 쿼리를 재구현하지 않음(호출 · SQL census).
6. **변이**(사본): (가) · (나) · (다) · (라) 각각 삭제, (나)의 「정확히 하나 · 마지막」 삭제, kind 검사 삭제, 누적 출처를 이번 체결 delta 로 바꾸기, 「합 ≥ 이번 delta」 검사 삭제, 인정 뒤 투영 쓰기(수량 다시 빼기), 판정을 거절 유형 무관하게 부르기.

## 4. 잔여 · 공시

- 성과(F15) 없음 — 인정된 손절도 성과에 잡히지 않는다. 후속 Story 후보.
- 조정을 사이에 둔 부분 체결은 여전히 `ATTRIBUTION_FAILED`(§2).
- 재매수 뒤 옛 주문의 체결은 판정에 닿지 않는다(§2, 별도 Story 후보).
- W2 · W3(체결 먼저)의 수량 불일치 고착은 재현됨, 이 change 가 다루지 않는다(별도 Story 후보). 부분 조정 뒤 초과 체결은 크레딧으로 풀림(`analysis/repro/out-of-scope.md`).
- a137 일일 손실 측정 불가는 인정 뒤에도 그대로.
- 인정된 손절에도 「외부에서 청산됨」 알림이 먼저 간다(수렴이 닫는 순간에 보내므로). 정정 알림 · 운영 로그 없음.
- flatten · 운영자 매도와 겹친 경합은 여전히 `ATTRIBUTION_FAILED`(결속 없음) — a194 가 알리고 a195 로 푼다.
- γ 보류(proposal).

## 5. 순서 · 결합

- 편집할 기존 함수는 **하나**: `ProjectPosition`(B5 1 자리). 판정 · 표지 쓰기는 새 파일. 그 밖으로 늘면 멈추고 보고.
- 고정 장치: `TestGuardedExitColumnsAreWrittenOnlyByTheApplyHook`(새 파일은 `PendingState` 로 읽음), `adoptionIDReaders` · `eligibilitySpellers`(새 파일은 그 열을 안 읽음), 전이표 census · `ValidReconcileCause` · `TradeOutcomeRefusalReasons`(새 값 없음 — 무적중), a166 census(배선 무편집 — 무적중), a130 `bindApplyHooks` 인자 수(무편집).
- a167(착수 전)은 `ApplyExitFill` 을 FLM 대상으로 적었으나 이 change 는 `ApplyExitFill` 무편집 — 겹침 없음. a199 와 편집 파일 겹침은 Manager 대조.
- base 재고정은 a198 착지 · merge 뒤(Manager 승인 2026-11-24, `review.md`).
