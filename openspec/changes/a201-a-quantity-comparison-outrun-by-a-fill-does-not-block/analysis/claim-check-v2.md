# a201 claim check v2 — 주기 보류 설계의 전제 (p1)~(p4)

HEAD `9722f2220`(v2 문서 개정 전), production Go 는 base 와 동일. AST `cited-ast/`(새로 note · logCycle · stabilise · Runtime.CheckHealth · backfillConfirmedStrategyFillTx), 탐침 `probe-v2/`(scratch 사본 — 끝난 뒤 제거).

## 판정 요약

| 전제 | 판정 | 한 줄 |
|---|---|---|
| (p1) 판 중단 stale 은 언제나 보류 조건과 함께 온다 | **거짓(좁은 반례 1)** | 엔진 안 투영 쓰기 중 체결 행 없이 수량을 바꾸는 길이 하나 있다 — 전략 진입 체결의 뒤늦은 투영 적용 |
| (p2) `Observe` 한 주기 생략이 추적기 불변식을 안 깬다 | **참** | Observe 없는 주기는 지금도 있고, 크레딧 · a192 무장 · 연속 횟수 · 진입 게이트 신선도 모두 다음 주기에 이어진다(탐침 확인) |
| (p3) 보류를 실패로도 깨끗함으로도 안 세는 기존 표현 | **없음** | 결과 소비자는 `note`(건강 계수)와 `logCycle`(로그) 둘이고, 둘 다 `cycle.Err` 의 유무 둘로만 가른다 |
| (p4) 편집할 기존 함수는 `RunOnce` 하나 | **거짓(D7' 그대로라면)** | 「깨끗함 로그에 안 들어감」은 `logCycle` 편집 없이는 못 한다 |

## (p1) — 판 중단 stale 과 보류 조건

- stale 의 원천은 `ApplyPositionAdjustment` 의 재검증 둘(position_adjustments.go:232-271 [손], AST 1단계 B15@253 · B17@267).
  - (b) 워터마크: tx 안 종목 워터마크 ≠ `FillWatermark` call@205 의 값 → 그 사이 그 종목 `fill_events` 행이 들어옴. 그 id 는 call@205 시점의 전역 최대보다 크고, call@205 는 D2' 의 읽기(로컬 읽기 전)보다 뒤이므로 **언제나 (가) 참**. 그 종목은 stale 이라 `Converged` 에 안 들어감 → (나) 참 → **보류**.
  - (a) 수량: 최신 인스턴스 수량 ≠ 로컬 읽기의 값 → 로컬 읽기와 조정 tx 사이에 그 종목 투영이 바뀜. 투영 쓰기 자리 전수(grep `UPDATE positions|INSERT INTO positions`, 비시험): `writeProjection`(position_projection.go:321 · 349 — 체결 apply 훅), `ApplyPositionAdjustment`(:320 · :347 — 수렴 converge.go:209 · 접기 external.go:222), adoption.go:241(adoption_id 만). 접기는 `ExternalPos` 종목만이라 같은 diff 의 `Quantities` 종목과 겹치지 않음(Compare B9 switch@407). 체결 apply 훅은 두 곳에서 돈다:
    - `RecordFill`(fills.go:551): delta > 0 이면 같은 tx 에서 `fill_events` 행을 먼저 넣음(fills.go:464-473) → 워터마크가 움직임 → (가) 참 → 보류. **(p1) 성립.** 훅은 delta 0 관측에서도 돈다(fills.go:478 이후, delta 블록 밖) — `position.Apply` 는 수량을 기준 ± delta 로만 바꾸므로(position.go:263-280 [손]) delta 0 이면 수량 불변. 전이표의 새 인스턴스 행(`NewInstance: true`)은 전부 `MoveAdds`(table.go:64-67 — grep 으로 `NewInstance: true` 이면서 `MoveAdds` 아닌 행 0)이고 delta 0 은 `MoveNone`(position.go:293-294)이라 delta 0 관측은 새 인스턴스를 만들지 않음(codex r2 지적으로 확인 — 앞 판본의 「미확인」 정정).
    - **`backfillConfirmedStrategyFillTx`(strategy_dispatch_runtime.go:1297-1385, 훅 call@1348)**: 전략 진입(BUY) 주문이 ACKED 창에 받은 체결을 확정 뒤에 투영에 뒤늦게 적용. **새 `fill_events` 행을 쓰지 않는다**(주석 :1352-1353 「The append-only fill event retains the original evidence commit time」, AST 호출 목록에 fill_events INSERT 없음). 그 종목이 이미 보유 중(추가 매수)이고 계좌가 그 매수를 이미 보여 `Quantities` 에 있을 때, 이 적용이 로컬 읽기와 조정 tx 사이에 커밋되면 수량 stale(a) → 판 중단인데 워터마크는 D2' 뒤로 안 움직였을 수 있음(체결 행이 D2' 전에 커밋됐으면) → **보류 안 걸림** → 지금과 같은 고착 가능.
- 귀결: 반례는 **지금 동작 그대로**(보류가 안 걸릴 뿐 더 나빠지지 않음)이고, 전략 진입 경로 · 추가 매수 · 확정 지연 · 그 창이 모두 겹쳐야 한다. 그 경로가 운영에서 켜져 있는지는 미확인. design §4 잔여 (iii) 으로 공시함.
- 엔진 밖 쓰기(다른 tossctl 프로세스): 위 자리를 부르는 생산 호출은 엔진 안뿐(grep — `ApplyPositionAdjustment` 생산 호출 2 자리 모두 대사 루프). 운영자 해제 명령은 엔진 정지를 요구함(a195).

## (p2) — `Observe` 한 주기 생략

- 지금도 `Observe` 가 안 도는 주기: RunOnce B1 if@429(불안정 — 계좌가 움직이는 동안 흔함) · B2 if@435(로컬 읽기 오류) · B5 if@458(Refresh 오류). **`Refresh` 성공 뒤 `Observe` 만 빠지는 주기는 지금 없음** — 보류가 새로 만드는 조합.
- `Refresh`(mismatch.go:1019-1084)는 장부의 활성 행으로 메모리 블록을 다시 만들고 게이트를 맞출 뿐 `Observe` 에 기대는 상태가 없다. 연속 횟수는 `continuityBrokenProposal` 일 때만 지움(B11 if@1070 → `clearStreaks` call@1075) — 지금도 매 주기 같은 조건으로 돈다.
- 크레딧: `creditUsableBy`(mismatch.go:509-519)는 「크레딧의 비교보다 엄격히 나중 as-of」면 씀. 보류한 주기의 수렴이 준 크레딧(그 주기 as-of)과 앞 주기 크레딧 모두 다음 주기 as-of 가 더 나중이라 쓰임. 반박(`refuteUsableAdjustmentCredits` call@716, B29 if@841)도 `Observe` 안에만 있어 보류한 주기에는 일어나지 않음. 탐침: 앞 크레딧 위 W2 → 보류 → 다음 깨끗한 주기에 `ADJUSTMENT_APPLIED`.
- a192 결속 · 무장: `beginOrderTerminalObservation`(call@647) · `orderTerminalCandidate`(call@682) · `updateOrderTerminalBindings`(call@808) 모두 `Observe` 안 — 보류하면 메모리 그대로 한 주기 늦게 이어짐. `Restore` 는 부르지 않으므로 결속이 버려지지 않음.
- `t.lastAt` · `t.observed`(mismatch.go:638-639)는 `Due` · `NextDueAt`(:1276-1293)만 읽고 그 둘의 생산 소비자는 0(grep).
- 진입 게이트의 신선도 요구(execgw/retry.go:196-203 `DefaultStaleness`)는 브로커 조회(열린 주문 · 매수가능 · 보유 · 가격)이고 수집은 보류 전에 끝남 — `Observe` 생략과 무관.
- 판정: **참**.

## (p3) — `RunOnce` 결과의 소비자 전수

| 소비자 | 읽는 것 | 보류한 주기를 무엇으로 읽나(표현을 안 바꾸면) |
|---|---|---|
| `RunOnce` 의 defer → `note`(reconcileloop.go:297-312, defer@426) | `cycle.Err == nil` → `Consecutive = 0`(B1 if@301 → return@305), 아니면 +1 | W2 모양(오류 없음) = **성공**(실패 연속을 0 으로 되돌림). W3 모양(수렴 stale 이 `cycle.Err`, B4 if@452) = **실패** — 오늘 W3 주기와 같음 |
| `Runtime.CheckHealth`(runtime.go:415-435) | `ConsecutiveFailures()` — 임계(5) 이상이면 치명 알림 · 운영 모드 강화, 0 이면 래치 해제(B4 if@422 → `clearLatch`@426) | 위 계수 그대로 |
| `Run` → `logCycle`(reconcileloop.go:548-569, call@413) | `cycle.Err` 있으면 `reconcile.mismatch` 오류 줄(B2 if@552), 없으면 `reconcile.clean`(call@558) | W2 모양 = `reconcile.clean`, W3 모양 = `reconcile.mismatch` 오류 |
| `judgeHoldings(&cycle)`(adoption.go:108 · 216) · a194 `blockedSince` 류 | 편입 판정 안의 카운터와 유예 시계 | 보류면 호출 안 됨(`Observe` 오류 경로와 같은 범위) — 유예 시계는 그 주기 진행 안 함 |
| `Health().Cycles` | 쓰기만(reconcileloop.go:300) | 생산 소비자 0 |
| 콘솔 · heartbeat | `ReconcileCycle` 을 읽는 생산 코드 0(grep — adoption.go 의 `*ReconcileCycle` 인자뿐) | — |

- 「오류 아닌 조기 반환」의 기존 선례는 B1 불안정 경로(`Err` 없음 · `Stable=false`)뿐이고, 그것은 **성공으로 세고 `reconcile.clean`(stable=false)으로 로그**한다. `Observe` 오류 경로(B7)는 **실패로 센다**. 실패도 성공도 아닌 표현은 없다.
- 「둘 다 안 셈」을 하려면 결과 타입에 칸(예: 보류 여부)을 더하고 (i) 건강 계수를 건너뛰고 (ii) 로그를 `stale_local_read` 줄로 바꿔야 함. (i)은 `RunOnce` 의 defer 클로저 안에서 할 수 있음(`note` 무편집, 대가: 그 주기는 `Cycles` 에도 안 들어감 — 소비자 0). (ii)는 `Run` 이 `logCycle` 을 부르고 `logCycle` 이 `Err` 유무 둘로만 가르므로 **`logCycle`(또는 `Run`) 편집이 필요**.
- 칸을 더할 때의 소비자 영향: 결과 타입을 읽는 생산 코드는 `logCycle` · `note`(defer 경유) · `judgeHoldings` 셋. 시험은 `cycle` 필드를 많이 단언함(reconcileloop_test 등) — 새 칸은 영값이 「보류 아님」이라 기존 단언 무변.

## (p4) — 편집할 기존 함수

- D7' 그대로: `RunOnce`(보류 판정 · defer 계수 건너뛰기) + `logCycle`(보류 줄 · clean 줄 안 씀) = **2 함수** → 거짓.
- `RunOnce` 하나로 하는 대안(D7' 완화): 보류 줄은 `RunOnce` 안에서 직접 쓰고, 계수와 clean/mismatch 줄은 지금 표현 그대로 둠 — W2 모양은 성공 · clean 줄, W3 모양은 실패 · mismatch 오류 줄(둘 다 **오늘 그 주기와 같은 계수 · 줄**). 연속 상한 1 이라 보류가 실패 연속을 길게 만들거나 지우는 효과는 한 주기에 그침.

## 「판 중단 뒤 B」 재현(v2 탐침)

- 편집 전: A · B 모두 QM 고착(1단계 재현).
- v2 탐침(보류 · 연속 상한 1 · 판독 실패 시 보류 안 함 — `probe-v2/probe.diff`):
  - B 의 체결이 보류한 주기 **뒤**, 다음 주기 **전**에 옴 → A · B 행 0, 다음 주기 깨끗.
  - B 의 체결이 다음 주기 **뒤**에 옴 → 다음 주기 수렴이 B 를 조정(closed 1 · 크레딧) · QM B → 체결 뒤 깨끗한 주기에 `ADJUSTMENT_APPLIED` 해제(기존 경로).
  - **남는 모양**: 보류한 주기 바로 다음 주기에 또 그 창에 체결이 떨어짐(연속 둘째) → 지금과 같은 고착(탐침 「연속 둘째」에서 S2 QM 고착).
- 같은 탐침: W2 · W3-a · tx 직전 · W3-b(합성) 모두 보류 → 행 0. 앞 크레딧 위 W2 → 보류 → 다음 주기 `ADJUSTMENT_APPLIED`. R1-1 순서 → 그 주기 해제 0 → 다음 주기(계좌 10 · 로컬 15) QM 유지 · 수렴 15→10(그다음 주기도 차단 유지 — 픽스처 원인 미조사).
