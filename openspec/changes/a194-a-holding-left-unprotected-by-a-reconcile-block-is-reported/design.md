# a194 Design

## 1. 결정

| # | 결정 | 근거 |
|---|---|---|
| D1 | 「운영자만 풀 수 있는 차단」의 판정은 기존 재료만 쓴다: `Tracker.Permanent()`(계좌 전체 **수량** 영구 승격 — 다른 원인의 계좌 행은 `EntryAllowed` 의 계좌 범위가 잡는다, AST: analysis/cited-ast/Permanent.ast.json · Refresh.ast.json) 또는 `Tracker.EntryAllowed(시장, 종목)` 의 사유가 `execgw.ReasonReconcilePermanent`. 새 Tracker 메서드를 만들지 않는다 | analysis/code-context/evidence-reconciliation.md 「F5 상세」. 판정을 둘로 두지 않는다 — 편입을 막는 판정(`blocked`)과 같은 출처 |
| D2 | 스스로 풀리는 차단의 유예 = **10 분**(상수 하나). 포지션별 「처음 빠진 시각」을 드라이버 메모리에 두고, 그 주기에 빠지지 않으면(편입·해제·청산) 지운다. 재시작하면 시계가 다시 0 이다 | 대사 주기 약 62 초 × 10. 수량 불일치는 조정 뒤 한두 주기에 풀리는 것이 정상이고(운영 장부의 수량 불일치 133 행 전부 해제됨), 10 분을 넘기면 「스스로 풀린다」는 전제가 이미 깨진 것이다. 재시작 뒤 시계 초기화는 공시한다 — 운영자 전용 차단은 유예가 없으므로 재시작에 가려지지 않는다 |
| D3 | 보유 보고는 기존 critical 타입 `exit.position_adoption_failed` 를 쓰고, 사실 칸에 새 값 둘(`blocked_operator_only`, `blocked_stale`)을 둔다. key 는 기존 `reconcileUnmanagedKey` 형식(타입·`reconcile`·사실·포지션) | 보호를 기대한 보유가 보호되지 않는다는 같은 사실이다. `internal/obs` 무편집. critical 은 메모리 래치를 거치지 않고 매 관측 기록을 시도하며 중복은 outbox key 와 재알림 창이 맡는다(a095 원칙) |
| D4 | 알림이 꺼진 엔진·편입이 꺼지고 그 종목이 지정되지도 않은 엔진의 등급은 기존 표를 따른다: 알림 꺼짐 → 일반 등급. **편입 꺼짐 ∧ 미지정 종목은 이 보고의 대상이 아니다** — 그 보유는 차단이 없어도 편입되지 않으므로 「차단 때문에 보호가 빠졌다」가 거짓이다. 그 경우는 기존 무관리 보고(일반 등급)가 차단 해제 뒤에 나간다 | engine-safety 「무관리 보유 보고의 등급은 사실이 정한다」의 표와 같은 방향. 판정 순서: 제외 목록·꺼짐∧미지정이면 대상 아님 → 그 밖의 차단 보유만 |
| D5 | 차단 보고는 기존 critical 타입 `reconcile.mismatch_permanent`, key = 타입·범위(계좌 전체 또는 종목)·원인·KST 날짜, 재알림 창 0(그날 한 번). 주기마다 활성 블록을 보고 기록을 시도한다 | 보유가 없는 차단은 급하지 않지만(11-20 의 차단은 3 일 뒤에야 해를 냈다) 잊히면 안 된다. 시간당 재알림은 하루 24 건이라 읽히지 않는다 |
| D6 | 본문은 원인·차단이 걸린 시각·할 일을 말한다: 「이 종목은 대사 차단(<원인>, <시각>부터) 중이라 엔진이 편입하지 않았다. 손절·익절이 걸려 있지 않다. 운영자 해제(`tossctl engine reconcile-resolve`) 전까지 앱에서 직접 관리해야 한다」. 스스로 풀리는 차단이면 「10 분 넘게 풀리지 않았다」를 말한다 | 2026-11-23 사고에서 운영자가 받은 것은 「보고만 하고 그대로 둔다」였다 |
| D7 | 제목·본문·필드에 계좌 정보를 싣지 않는다. 종목 표시는 기존 `label` 을 쓴다 | engine-safety 「시스템 밖으로 나가는 알림에는 계좌 정보가 실리지 않는다」 |

## 2. 편집 자리

- A(채택): `judgeHoldings` 의 차단 분기(if@152)에서 `continue` 하기 전에 그 보유를 모으고, 순회 뒤 새 파일(`internal/app/engine/reconcile_block_report.go`)의 새 함수들이 판정·기록한다. 차단 보고(D5)도 같은 자리에서 부른다 — `judgeHoldings` 는 `Observe` 가 성공한 주기의 마지막 단계라 블록 목록이 그 주기의 것이다.
- B(기각): exit 루프 `workingSet` 에서 보고 — `a090_workingset_exits_test.go:22-35`(`a090FrozenWorkingSetExits`)가 그 루프의 이탈 11 개를 고정한다.

`Observe` 가 실패한 주기에는 `judgeHoldings` 가 돌지 않는다(AST: analysis/cited-ast/RunOnce.ast.json B7 if@463 → return@467). 같은 호출 전에 반환하는 갈래가 셋 더 있다 — 안정화 실패·미수렴 B1 if@422 → return@423, 로컬 상태 읽기 실패 B2 if@428 → return@430, `Tracker.Refresh` 실패 B5 if@451 → return@455. 그 주기는 보고도 없다 — 대사 실패 자체의 기존 보고(`reconcile.mismatch` — `logCycle` 의 ERROR 로그 줄(reconcileloop.go:546)이고 알림이 아니다. 안정화 미수렴 주기는 오류가 아니라 이 줄도 없다)가 있다. 유예 시계는 그런 주기에 지우지 않는다(빠지지 않았다는 관측이 아니다).

## 3. 거절·침묵하는 정상 입력

- 수량 불일치로 막 차단된 보유: 10 분 전까지 침묵(정본의 전이 상태). 사람이 방금 산 종목의 정상 경로다.
- 편입 꺼짐 ∧ 미지정 종목의 차단 보유: 이 보고 없음(D4).
- 이미 exit 자격을 가진(`ExitEligible` — 진입 결정 또는 편입 기록이 있는, internal/journal/position_projection.go:515-518) 보유: 이 보고 없음. 자격은 exit state 가 실제로 열렸음을 보증하지 않는다 — 편입 커밋 뒤 보호 미개설 보유(adoption.go `adoptOne`, 정본 engine-safety 「편입 기록이 커밋됐는데 보호가 아직 열리지 않은 보유」 경계)는 이 change 범위 밖이다.

## 4. 시험 전략 (RED 먼저)

새 시험 파일 `internal/app/engine/a194_reconcile_block_report_test.go`(기존 `reconcileloop_test.go` 하네스 재사용, 가짜 시계):

1. 운영자 전용 차단(`ATTRIBUTION_FAILED` 종목 행) + 편입 켜짐 + 그 종목 보유 → 첫 주기에 critical 1(타입·key·사실 `blocked_operator_only`), 편입 0, 다음 주기들에 outbox 행은 그대로 1.
2. 수량 불일치 차단 + 보유 → 9 분 59 초까지 0, 10 분에 critical 1(사실 `blocked_stale`). 중간에 차단이 풀려 편입되면 0 이고 시계가 지워진다(다시 차단되면 새 10 분).
3. `Observe` 실패 주기가 끼어도 시계는 이어진다.
4. 편입 꺼짐 ∧ 미지정 + 운영자 전용 차단 + 보유 → 이 보고 0. include 지정 종목이면 1.
5. 알림 꺼진 엔진 → critical 0, 일반 등급 기록.
6. 제외 목록 종목 → 이 보고 0.
7. 운영자 전용 차단 활성, 보유 없음 → `reconcile.mismatch_permanent` 그날 1, 같은 날 주기 반복에도 1, 가짜 시계로 다음 KST 날짜에 1 더. 차단이 풀리면 더는 없음. 계좌 전체 영구 차단도 1(종목마다가 아니라 한 통).
8. 스스로 풀리는 차단만 활성 → 차단 보고 0.
9. 제목·본문·필드에 합성 계좌·수량·가격 문자열 0.
10. 기존 `TestAdoptionIsSilentUnderReconcile`·`TestActiveForeignCauseIsCountedAndBlocksBeforePriceRead` 무편집 통과(후자가 알림 수를 단언하면 Manager 에게 보고).

변이(전부 CAUGHT, 변이마다 닿음 확인): 운영자 전용 판정을 항상 거짓으로 · 유예 비교 방향 뒤집기 · 시계를 지우지 않기 · D4 조건 삭제 · 차단 보고 key 에서 날짜 삭제 · critical 앞에 메모리 래치 넣기(기록 실패 뒤 재시도가 죽어야 함).

선택 시험: `./internal/app/engine -run 'A194|Adoption|ReconcileLoop|Unmanaged'`, `-race` 는 `-run A194` 만.

## 5. 정본 delta 가 기존 문장과 어떻게 맞는가

정본 exit-policy 는 「전이 상태만 무알림」이라고 적는다. 이 change 는 그 문장을 뒤집지 않고 **전이 상태의 경계**를 정한다 — 운영자만 풀 수 있는 차단은 전이가 아니고, 스스로 풀리는 차단은 10 분까지만 전이다. 기존 Scenario 「RECONCILE 중 편입 시도 … 무알림」은 유예 안의 스스로 풀리는 차단에 대해 여전히 참이다.

## 6. 잔여·후속

- 재시작이 유예 시계를 0 으로 되돌린다(D2). 반복 재시작 중에는 스스로 풀리는 차단의 보유 보고가 계속 미뤄질 수 있다 — 엔진 생존 감시(a134)의 영역.
- 차단이 걸리는 순간의 원인별 상세(어느 주문·어느 전이)는 본문에 싣지 않는다. 장부의 `reconcile_states.evidence` 가 정본이다.
