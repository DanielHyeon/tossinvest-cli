# a194 — 대사 차단 때문에 손절이 걸리지 않은 보유는 알려진다

> **상태: 1단계(계약·증거), High-risk(대사·편입·알림).** Story `STORY-TOS-a194`(FEAT-TOS-005). 작성: Manager(2026-11-23).
> 분기 주장의 AST 는 조사 산출물(세션 스크래치 `inv-D/*.ast.json`, HEAD `a9dd53400`)에서 인용했고, 이 change 안 사본과 현재 HEAD 재확인은 tasks 1.3 이다 — 그 전까지 「조사 인용」이다.

## Why

2026-11-23 운영(Manager 실측 — 운영 장부·엔진 로그, 시각·상태만):

- 08:09 KST 사람이 앱에서 산 종목의 보유가 장부에 외부 포지션으로 접혔지만 **편입되지 않았다**(`adoption_id` 없음, `exit_states`·`exit_events` 0 행). 10:38 엔진 밖에서 닫힐 때까지 손절이 없었다.
- 그 종목에는 11-20 12:56 에 걸린 `ATTRIBUTION_FAILED` 차단이 풀리지 않은 채 있었다. 차단이 걸린 순간에도, 3 일 동안에도, 보유가 생긴 순간에도 `alert_outbox` 에 이 일에 대한 행이 없다. 로그는 매분 `reconcile.clean … blocked:1`(INFO)였고, 08:09:02 의 `exit.position_unmanaged` 는 일반 등급 로그 한 줄이었다.
- 같은 날 08:24 에 산 다른 종목은 정상 편입되어 손절까지 나갔다.

코드 근거:

| 사실 | 근거 |
|---|---|
| 차단 종목의 보유는 편입 판정에서 조용히 빠지고 무관리 보고에도 들어가지 않는다 | `internal/app/engine/adoption.go` `ReconcileDriver.judgeHoldings` if@152 → continue@153; 빠진 보유는 `unmanaged` 에 없어 range@188 의 `alertUnmanaged` 에 닿지 않는다(조사 인용 F1) |
| 그 침묵은 정본의 요구다 — 전제는 「전이 상태는 스스로 풀린다」 | exit-policy 「외부 취득 포지션의 자동 편입」 알림 문단·Scenario 「RECONCILE 중 편입 시도 … (전이 상태 — 무알림)」, adopt-external-positions design A4 |
| 그 전제가 거짓인 차단이 있다 — 운영자만 풀 수 있는 원인 | `blockFromReconcileState` 는 `QUANTITY_MISMATCH` 만 자동 해제 대상이고 나머지는 `ReleaseOperatorOnly`·`Permanent`(`internal/reconcile/mismatch.go:1398-1412`, 조사 인용 F5) |
| 차단이 걸리는 순간 알림을 내는 자리가 없다 | `reconcile_states` 를 쓰는 8 자리 어디에도 알림 호출이 없다(조사 인용 F6, grep) |
| critical 로 정의된 `reconcile.mismatch_permanent` 를 내는 production 코드가 0 곳이다 | `internal/obs/event.go` 에 정의만 있다(조사 인용 F7, grep) |

즉 **운영자만 풀 수 있는 차단은 걸려도 아무도 모르고, 그 종목에 보유가 생겨 손절이 빠져도 아무도 모른다.**

## What Changes

1. **편입되지 못한 보유의 보고**: 대사 차단 때문에 편입 판정에서 빠진 보유가 (가) 운영자만 풀 수 있는 차단 아래에 있으면 **즉시**, (나) 스스로 풀리는 차단 아래에서 **10 분 이상 연속으로** 빠졌으면 그때, 「이 종목은 대사 차단 중이라 손절이 걸려 있지 않다」를 알린다. 등급은 기존 규칙 그대로 — 알림이 켜진 엔진에서 critical, 꺼진 엔진에서 일반 등급.
2. **차단 자체의 보고**: 운영자만 풀 수 있는 차단이 활성인 동안, 보유가 없어도 **하루 한 번** critical 로 알린다(기존 타입 `reconcile.mismatch_permanent`). 걸린 날 첫 주기에 나간다.
3. 편입 여부·차단을 거는 조건·차단 해제 규칙·청산 경로는 바꾸지 않는다. 알림만 더한다.

## 이 change 가 하지 않는 것

- 차단 종목의 편입 허용. 「RECONCILE 인 종목은 편입하지 않는다」는 그대로다.
- 차단의 해제(a195)·차단이 생기지 않게 하는 수리(a195 design §5 의 후속).
- 이미 편입된 보유에 확정 하한 캡이 적용 중이라는 알림.
- `reconcile.clean` 이 `blocked>0` 에서도 INFO 로 찍히는 이름 문제.
- 새 알림 타입. `internal/obs` 의 타입 열거는 편집하지 않는다.

## Capabilities

### Modified Capabilities

- `exit-policy`: 편입 판정의 「전이 상태 무알림」이 어디까지인지 정한다(운영자 전용 차단은 전이가 아니고, 스스로 풀리는 차단도 10 분을 넘기면 전이가 아니다).
- `reconciliation`: 운영자만 풀 수 있는 차단은 활성인 동안 하루 한 번 알려진다.

## Impact

- 기존 함수 본문 편집(gate ⑤ FLM 대상): `ReconcileDriver.judgeHoldings`. 가능하면 이 하나로 끝낸다(새 판정·보고 함수는 새 파일). 차단 보고를 다른 기존 함수(`RunOnce` 등)에 걸어야 하면 그 함수의 번들도 만든다.
- a166 census·a163 판정 핀 밖(조사 인용). 기존 시험 `TestAdoptionIsSilentUnderReconcile`(수량 불일치, 10 분 전)는 그대로 통과해야 한다.
- 브로커 호출 추가 0. 원장 스키마 변경 없음. 알림 수: 운영자 전용 차단당 하루 1 건, 보호 없는 보유당 재알림 창(기본 1 시간)마다 1 건.
- 알림이 켜진 엔진에서 critical 이 전달되지 못하면 기존 규칙대로 진입이 잠긴다 — 새 동작이 아니라 critical 의 기존 성질이다.
