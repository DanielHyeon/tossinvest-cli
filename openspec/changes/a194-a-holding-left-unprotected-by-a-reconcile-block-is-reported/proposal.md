# a194 — 대사 차단 때문에 손절이 걸리지 않은 보유는 알려진다

> **상태: 1단계(계약·증거), High-risk(대사·편입·알림).** Story `STORY-TOS-a194`(FEAT-TOS-005). 작성: Manager(2026-11-23).
> 분기 주장의 AST 는 tasks 1.3 에서 이 change 안 `analysis/cited-ast/`(HEAD `064f59fd9`)로 다시 뽑아 대조했다 — 각 줄의 근거 표시가 그 사본을 가리킨다. 대조 기록은 `review.md` 「1.3 주장 대조」.

## Why

2026-11-23 운영(Manager 실측 — 운영 장부·엔진 로그, 시각·상태만):

- 08:09 KST 사람이 앱에서 산 종목의 보유가 장부에 외부 포지션으로 접혔지만 **편입되지 않았다**(`adoption_id` 없음, `exit_states`·`exit_events` 0 행). 10:38 엔진 밖에서 닫힐 때까지 손절이 없었다.
- 그 종목에는 11-20 12:56 에 걸린 `ATTRIBUTION_FAILED` 차단이 풀리지 않은 채 있었다. 차단이 걸린 순간에도, 3 일 동안에도, 보유가 생긴 순간에도 `alert_outbox` 에 이 일에 대한 행이 없다. 로그는 매분 `reconcile.clean … blocked:1`(INFO)였고, 08:09:02 의 `exit.position_unmanaged` 는 일반 등급 로그 한 줄이었다.
- 같은 날 08:24 에 산 다른 종목은 정상 편입되어 손절까지 나갔다.

코드 근거:

| 사실 | 근거 |
|---|---|
| 차단 종목의 보유는 편입 판정에서 조용히 빠지고 무관리 보고에도 들어가지 않는다 | `internal/app/engine/adoption.go` `ReconcileDriver.judgeHoldings` if@152 → continue@153; 빠진 보유는 `unmanaged` 에 없어 range@188 의 `alertUnmanaged` 에 닿지 않는다(AST: analysis/cited-ast/judgeHoldings.ast.json B10 if@152 · B16 range@188 — continue@153 은 소스, logic-map 은 continue 를 return 으로 세지 않음) |
| 그 침묵은 정본의 요구다 — 전제는 「전이 상태는 스스로 풀린다」 | exit-policy 「외부 취득 포지션의 자동 편입」 알림 문단·Scenario 「RECONCILE 중 편입 시도 … (전이 상태 — 무알림)」, adopt-external-positions design A4 |
| 그 전제가 거짓인 차단이 있다 — 운영자만 풀 수 있는 원인 | `blockFromReconcileState` 는 종목 범위 `QUANTITY_MISMATCH` 만 자동 해제(`ReleaseOnAdjustedReconcile`)이고 나머지 원인은 전부 `ReleaseOperatorOnly` 다. `Permanent` 는 사유가 `ReasonReconcilePermanent` 인 원인(IDENTIFIER_CONFLICT·ATTRIBUTION_FAILED·알 수 없는 원인, 그리고 계좌 범위 QUANTITY_MISMATCH)만 참이고 SNAPSHOT_UNAVAILABLE·SNAPSHOT_STALE 은 operator_only 이면서 Permanent 거짓이다 — 그 둘은 생산 쓰기 자리가 0 이다(AST: analysis/cited-ast/blockFromReconcileState.ast.json B1 if@1411 · B2 if@1414 · B3 if@1417, analysis/cited-ast/ReconcileReasonFor.ast.json B2 case@159 · B3 case@163; 원인별 표는 analysis/code-context/evidence-reconciliation.md) |
| 차단이 걸리는 순간 알림을 내는 자리가 없다 | `reconcile_states` 를 쓰는 자리 어디에도 알림 호출이 없다 — 쓰기는 `internal/journal`(obs 를 import 하지 않음 — obs 가 journal 을 import)과 `internal/reconcile`(알림 호출은 `acked_boot.go:98` ACKED attempt 보고 하나)에만 있다(grep, 자리 목록은 analysis/code-context/evidence-reconciliation.md) |
| critical 로 정의된 `reconcile.mismatch_permanent` 를 내는 production 코드가 0 곳이다 | `internal/obs/event.go:95`(`EventReconcilePermanent`)·`:370`(critical 표)에 정의만 있다 — 그 밖의 참조는 obs 시험 두 파일뿐(grep, 2026-11-23 HEAD `064f59fd9`) |

즉 **운영자만 풀 수 있는 차단은 걸려도 아무도 모르고, 그 종목에 보유가 생겨 손절이 빠져도 아무도 모른다.**

## What Changes

1. **편입되지 못한 보유의 보고**: 대사 차단 때문에 편입 판정에서 빠진 보유가 (가) 운영자만 풀 수 있는 차단 아래에 있으면 **즉시**, (나) 스스로 풀리는 차단 아래에서 **10 분 이상 연속으로** 빠졌으면 그때, 「<관측 시각> 대사에서 이 종목은 차단 아래에 있어 편입하지 못했고 그 시각 손절이 걸려 있지 않았다」를 원인·차단 시각과 함께 알린다. 등급은 기존 규칙 그대로 — 알림이 켜진 엔진에서 critical(`exit.position_adoption_failed`), 꺼진 엔진에서 일반 등급(`exit.position_unmanaged`, 한 번). 무관리 계수 `cycle.Unmanaged` 는 올리지 않는다.
2. **차단 자체의 보고**: 운영자만 풀 수 있는 차단이 활성인 동안, 보유가 없어도 **하루(KST) 한 번** 알린다 — 알림이 켜진 엔진은 critical(기존 타입 `reconcile.mismatch_permanent`), 꺼진 엔진은 엔진 로그 한 줄. 걸린 날 첫 성공한 대사 주기에 나간다. 「하루 한 번」은 기록 성공 뒤에만 서는 드라이버 메모리 래치이고, 재시작하면 그날 한 번 더 나갈 수 있다.
3. 편입 여부·차단을 거는 조건·차단 해제 규칙·청산 경로는 바꾸지 않는다. 알림만 더한다. critical 알림이 전달되지 못했을 때의 진입 잠금은 critical 알림의 기존 규칙 그대로다(아래 Impact).

## 이 change 가 하지 않는 것

- 차단 종목의 편입 허용. 「RECONCILE 인 종목은 편입하지 않는다」는 그대로다.
- 차단의 해제(a195)·차단이 생기지 않게 하는 수리(a195 design §5 의 후속).
- 이미 편입된 보유에 확정 하한 캡이 적용 중이라는 알림.
- `reconcile.clean` 이 `blocked>0` 에서도 INFO 로 찍히는 이름 문제.
- 새 알림 타입. `internal/obs` 의 타입 열거는 편집하지 않는다.
- 대사 주기가 보유 판정 전에 조기 반환하는 동안(안정화 미성공·로컬 상태 조회 실패·차단 목록 갱신 실패·`Observe` 실패)의 보고. 그 주기에는 차단 목록·보유 판정 자체가 없으므로 보유 보고도 차단 보고도 나가지 않는다 — 그 구간은 기존 대사 실패 보고와 루프 저하 경보의 영역이다(Story limits).
- 스스로 풀리는 차단이 걸리는 순간의 알림(정본 reconciliation 「수량 불일치 감지 … 알림이 발송된다」) — 이름 붙인 공백으로 남긴다.

## 정본의 기존 공백과의 관계

정본 engine-safety 「등급화된 알림」은 「영구 불일치」를 critical 로 적지만(`openspec/specs/engine-safety/spec.md:199`) 지금 생산 발행자가 0 이다(위 F7). 이 change 의 차단 보고가 운영자 전용 차단(계좌 전체 영구 차단 포함)에 대해 그것을 처음 구현한다. 정본 reconciliation 「수량 불일치 감지」 Scenario 의 「알림이 발송된다」(`openspec/specs/reconciliation/spec.md:83`)는 이 change 가 메우지 않는다 — 스스로 풀리는 차단의 진입 알림은 범위 밖이다.

## Capabilities

### Modified Capabilities

- `exit-policy`: 편입 판정의 「전이 상태 무알림」이 어디까지인지 정한다(운영자 전용 차단은 전이가 아니고, 스스로 풀리는 차단도 10 분을 넘기면 전이가 아니다). 「외부 취득 포지션의 자동 편입」을 MODIFIED 로 좁힌다.
- `engine-safety`: 「무관리 보유 보고의 등급은 사실이 정한다」의 critical 조건 집합에 「전이 상태가 아닌 대사 차단으로 편입되지 못한 편입 후보」를 더한다(MODIFIED).
- `reconciliation`: 운영자만 풀 수 있는 차단은 활성인 동안 하루 한 번 알려진다.

## Impact

- 기존 함수 본문 편집(gate ⑤ FLM 대상): `ReconcileDriver.judgeHoldings` 하나(새 판정·보고 함수는 새 파일, 메모리 상태는 드라이버 구조체의 새 필드 — 구조체 선언은 함수 본문이 아니고 맵은 새 함수 안에서 만든다). `NewReconcileDriver`·`Context.ReconcileDriver`·`RunOnce` 무편집 — 새 배선 없음.
- a166 census·a163 판정 핀 밖(1.4 확인 — analysis/code-context/codegraph-baseline.md 「구조 고정·census」). 기존 시험 `TestAdoptionIsSilentUnderReconcile`(수량 불일치, 10 분 전)·`TestUSIncludedSymbolWaitsUnderAccountWidePermanentReconcile`(`cycle.Unmanaged` 0 단언)·`TestA095NotificationsOffNeverRecordsACritical` 는 무편집 통과해야 한다.
- 브로커 호출 추가 0. 원장 스키마 변경 없음. 알림 수: 운영자 전용 차단당 하루 1 건(재시작 시 그날 1 건 더 가능), 보호 없는 보유당 재알림 창(기본 1 시간)마다 1 건 — 전달이 정산된 행에 한한다. 미전달(PENDING) 행은 창과 무관하게 매 관측이 다시 전송을 시도한다(journal/outbox.go:379, r2 R2-2). 알림 꺼짐이면 프로세스당 1 건. 계좌 전체 영구 차단 아래에서는 편입 후보 보유마다 1 건씩 나간다.
- 알림이 켜진 엔진에서 critical 이 전달되지 못하면 기존 규칙대로 계좌 전체 진입이 잠기고, 지속되면 운영 모드가 승격된다(사람의 mode-release 로만 풀림 — `internal/app/engine/alertdelivery.go` 머리 주석 「지속 실패의 주인 (a124)」) — 새 동작이 아니라 critical 의 기존 성질이다.
