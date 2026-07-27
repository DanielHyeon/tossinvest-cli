# 애그리게이트 경계와 이벤트 흐름

> 참고 산출물(add-core-domain task 6.4). **규범은 `openspec/specs`의 order-execution·position-ledger·exit-policy 요구사항**이고, 이 문서는 그 경계를 코드와 나란히 놓아 읽을 수 있게 한 지도다.
> 구현: `internal/journal`(원장·투영·조정·provenance), `internal/position`(전이표), `internal/execgw`(결정·게이트웨이), `internal/reconcile`(대사).
> 브로커 동작 주장 없음. 이 문서의 모든 관계는 스키마에 선언된 컬럼이다.

## 0. 왜 경계를 문서로 쓰는가

애그리게이트 경계는 "어느 코드가 어느 행을 쓸 수 있는가"의 다른 이름이다. 그것이 흐려지면 같은 수량을 두 곳이 계산하고, 두 계산이 갈라지는 날 엔진은 자기 노출을 모르는 상태로 주문을 낸다. 이 change의 SHALL 세 개가 정확히 그 방지책이다 — 투영은 직접 변이 API를 노출하지 않는다(SHALL NOT), 대사의 로컬 상태는 그 투영을 소비한다(SHALL), 조인 경로는 명시적 참조 컬럼이다(SHALL).

## 1. 네 개의 애그리게이트

| 애그리게이트 | 루트 | 소유 테이블 | 유일한 쓰기 경로 | 불변식 |
|---|---|---|---|---|
| **Order** | `intents.id` | `intents`(불변) · `mutation_attempts` · `attempt_transitions`(append-only) · `lineage_edges` | `journal.Prepare` → `attempt.Mark*/Settle` | intent 행은 기록 후 갱신되지 않는다. 상태 변화는 전부 `attempt_transitions`에 추가된다. 정정 교체는 행 변이가 아니라 lineage **간선**이다(공식 API가 새 주문번호를 준다) |
| **Fill** | `fill_snapshots.order_id` | `fill_snapshots`(주문당 최신 누적) · `fill_events`(append-only 양의 delta) · `execution_corrections` | `journal.RecordFill` | 누적 수량은 감소하지 않는다. fail-closed 스냅샷은 사실로 채택되지 않는다. 재관측 멱등은 누적 watermark가 담보한다 |
| **Position** | `positions.id` (계좌·시장·심볼·instance_seq) | `positions` · `position_adjustments`(append-only) | 체결 apply hook(`ProjectPosition`) · `ApplyPositionAdjustment` | **투영이다**. 체결 이벤트나 기록된 조정에서 오지 않은 수량은 쓰이지 않는다. CLOSED는 종결이며 재진입은 새 instance_seq다. `entry_decision_id` NULL = 외부/수동 편입 |
| **ExitState** | `exit_states.position_id` | `exit_states` · `exit_events`(append-only) | 판정 루프(7.x) · 체결 apply hook | 포지션당 정책 하나. `baseline_price`는 t0에 진입 손절이며 단조 비감소. `taken_ratio_total`·pending 3컬럼은 **apply hook에서만** 움직인다(`ApplyTx` 메서드 외 경로 없음) |

경계 밖에 있는 것: `decisions`·`risk_reservations`·`operating_modes`·`reconcile_states`는 애그리게이트가 아니라 **권한과 상태**다. 결정은 시도의 근거이고, 예약은 그 근거가 한도에서 붙잡은 몫이며, 모드와 RECONCILE은 계좌 수준 게이트다. 포지션 하나의 수명주기에 속하지 않으므로 위 표에 없다.

## 2. 이벤트 흐름

```
                    ┌──────────────┐
                    │  decisions   │  발급자: 체인 ALLOW → 결정+예약 한 트랜잭션
                    └──────┬───────┘
                           │ decision_id (FK)
                    ┌──────▼───────┐
   Order 애그리게이트 │   intents    │──┐
                    │mutation_attempts│ │ intent_id
                    └──────┬───────┘  │
                           │ broker_order_id
                    ┌──────▼───────┐  │
   Fill 애그리게이트  │fill_snapshots│  │  RecordFill (BEGIN IMMEDIATE)
                    │ fill_events  │  │        │
                    └──────┬───────┘  │        │ 같은 트랜잭션 (apply hook)
                           │ AppliedFill       ▼
                    ┌──────▼──────────────────────┐
 Position 애그리게이트│  positions                  │◄── position_adjustments
                    │  (전이표 96행이 결정)         │    (compare-and-append)
                    └──────┬──────────────────────┘
                           │ position_id (FK)
                    ┌──────▼───────┐
 ExitState 애그리게이트│ exit_states  │──► exit_events ──► proposed_intent_id ──┐
                    └──────────────┘                                        │
                                                                            │
                    ┌───────────────────────────────────────────────────────┘
                    │  청산 intent → attempt → fill → Position 감소 → CLOSED
                    ▼
                 trade_outcomes (CLOSED 트랜잭션에서 동결)
```

읽어야 할 세 가지:

1. **체결 → 투영 → exit은 한 커밋이다.** `RecordFill`의 트랜잭션이 주입된 apply 함수를 tx-scope에서 호출한다(`apply_hook.go`). journal 밖의 후속 커밋은 원자성 요구를 만족하지 않는다 — 그 사이의 크래시는 투영이 보지 못한 체결을 남기고, 투영은 대사·exit 정책·진입 게이트가 전부 읽는 것이다.
2. **조정은 별도의 트랜잭션이지만 같은 권위다.** 계좌가 권위이고(SHALL), 그 값은 append-only 조정 이벤트로만 행에 닿는다. 덮어쓰기는 금지다 — 투영이 체결과 불일치하는 **이유**가 그 행이기 때문이다.
3. **exit은 포지션을 직접 줄이지 않는다.** 판정은 intent를 발의하고, 그 intent는 Guardian의 위험 감소 경로를 지나 Order 애그리게이트로 들어가며, 포지션은 그 주문의 체결로만 줄어든다. 화살표가 한 바퀴 도는 것은 실수가 아니라 경계다.

## 3. Provenance 조인 경로

"이 포지션은 왜 존재하는가"는 `journal.PositionProvenance(ctx, positionID)` 단일 질의다(`internal/journal/provenance.go`). 전부 선언된 참조 컬럼이며 **시간창 휴리스틱은 없다**.

| 단계 | 간선 |
|---|---|
| `DECISION` | `positions.entry_decision_id → decisions.id` |
| `INTENT` | `mutation_attempts.decision_id → decisions.id`, `mutation_attempts.intent_id → intents.id` |
| `ATTEMPT` | 위 attempt 행 |
| `FILL` | `mutation_attempts.broker_order_id → fill_events.order_id` |
| `POSITION` | `positions.opened_at` |
| `ADJUSTMENT` | `position_adjustments.position_id → positions.id` |
| `EXIT_EVENT` | `exit_events.position_id → positions.id` |
| `EXIT_INTENT` | `exit_events.proposed_intent_id → intents.id` |
| `EXIT_ATTEMPT` · `EXIT_FILL` | 그 intent의 attempt와 그 주문번호의 체결 |
| `CLOSE` | `positions.closed_at` |

**닿지 않는 것**(의도적): exit 판정을 거치지 않은 감소 — 수동 flatten, 운영자의 매도. 그것들은 자기 결정과 자기 intent를 가지며, 포지션에 붙일 선언된 컬럼이 없다. 시간으로 붙이는 것이 스펙이 금지한 바로 그 휴리스틱이다. 대안(`fill_events`에 인스턴스 참조 추가)은 새 스키마 버전이고 이 change의 마이그레이션 규칙(단일 원자 v6)에 어긋난다 — `openspec/changes/add-core-domain/issues.md`에 선택지를 기록했다.

**외부 포지션**은 `entry_decision_id`가 NULL이므로 체인에 DECISION·INTENT·ATTEMPT·FILL이 **없다**. 가장 가까운 결정을 붙이지 않는 것이 정답이다 — NULL이 이미 기록한 사실("이 포지션을 정당화하는 결정이 없다")을 그대로 말하는 것이다.

## 4. 보호 saga 경계 — 2c 예약

브로커측 보호주문(조건주문·상주 stop)은 **이 change에 코드가 0줄**이다. 그것이 도착하면 다섯 번째 애그리게이트가 생긴다:

- 루트: 보호주문 자체(브로커에 상주하는 주문 식별자)
- saga: 포지션 수량이 움직일 때마다 보호주문 수량을 맞추는 보상 트랜잭션 — 로컬 판정과 달리 **실패할 수 있는 원격 상태**라서 saga다
- 경계: `exit_states.baseline_price`(로컬 판정)와 브로커 상주 stop(원격 사실)은 **다른 값**이며, 둘을 한 컬럼에 넣으면 로컬 판정이 실패한 원격 배치를 성공으로 읽는다
- 인터록 조항 6(`ProtectionReady`)이 그때까지 게이트 ON을 기계적으로 막는다 — 현재는 미충족 상수다

지금 이 문서에 적어두는 이유: 2c가 `exit_states`에 컬럼을 더하는 대신 자기 테이블을 갖도록 경계를 미리 그어두면, 그 change가 additive 마이그레이션 하나로 끝난다.

## 5. 무엇이 무엇을 쓸 수 없는가 (테스트가 강제하는 것)

| 규칙 | 강제 |
|---|---|
| `taken_ratio_total`·pending 3컬럼은 apply hook에서만 | `TestGuardedExitColumnsAreWrittenOnlyByTheApplyHook`(런타임 거부 + 파일 배치 검사) |
| 투영은 체결·조정에서만 움직인다 | `positions`에 대한 쓰기가 `ProjectPosition`과 `ApplyPositionAdjustment` 둘뿐 |
| 대사의 로컬 포지션은 투영이다 | `TestLocalStateReadsTheProjectionNotTheFills`(조정 후 체결 원장 10, 로컬 상태 4) |
| provenance는 명시 참조만 | `TestAnotherDecisionsFillsAreNotThisPositionsProvenance`, `TestAnExternalPositionHasNoEntryProvenance` |
| 금지 전이는 보정하지 않고 RECONCILE | `internal/position` 96행 전이표 전수 테스트 + `TestAnAdjustmentConvergesAFrozenProjection` |
