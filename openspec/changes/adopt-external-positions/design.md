# Design: adopt-external-positions

> 리뷰 라운드 1(P1 8)·라운드 2(P1 7) 반영. High-risk 경로(원장 스키마·exit 자격 게이트)이므로 이 문서가 테이블 정의·의미론의 정본이다. 항목 참조는 반드시 change-id를 붙인다("adopt-external-positions design A1" — add-core-domain design D7과의 혼동 방지).

## A1. 편입의 저장 형태 — 별도 테이블, 결정 테이블 무접촉

`decisions.safety_class`·`preimage_kind`는 SQL CHECK enum이고 additive 규칙이 테이블 재작성을 금지한다. **편입은 mutation이 아니므로 class 축에 놓지 않는다** — engine-safety·risk-management의 3-class 계약은 무변경.

**v7 마이그레이션(additive만)**:

```sql
CREATE TABLE position_adoptions (
    id              TEXT PRIMARY KEY,
    symbol          TEXT NOT NULL,             -- 편입 시점 스냅샷(권위는 positions)
    market          TEXT NOT NULL,             -- 〃
    quantity        TEXT NOT NULL,             -- 〃 decimal string
    cost_basis      TEXT,                      -- 브로커 averagePurchasePrice 원문 보존(있으면)
    cost_basis_src  TEXT NOT NULL,             -- 'BROKER_AVG' | 'ABSENT'
    observed_price  TEXT NOT NULL,             -- t0 EntryPrice — 출처·제약은 A2
    synthetic_stop  TEXT NOT NULL,             -- observed_price × (1 − default_stop_pct)
    observed_at     TEXT NOT NULL,
    preimage_digest TEXT NOT NULL
) STRICT;
ALTER TABLE positions ADD COLUMN adoption_id TEXT REFERENCES position_adoptions(id);
```

- 참조는 `positions.adoption_id` **단방향**이다(라운드 2: 상호 FK 순환 제거 — position_adoptions에 position_id를 두지 않는다). 같은 tx에서 adoption 행 삽입 → adoption_id 기입 순서. symbol·quantity 등은 편입 시점 스냅샷이며 이후 권위는 positions 투영이다.
- `adoption_id`는 **set-once**: 전용 tx API로만 기입, 그 외 UPDATE의 언급은 정적 스캔이 거부(guarded-column 스캔 방식). `entry_decision_id`는 이 change에서 **읽기만**.
- exit 자격 = `entry_decision_id IS NOT NULL OR adoption_id IS NOT NULL` — 단일 술어 함수로 통합. 단, reconcile fold 가드(external.go:225-234 "entry 결정 상속 금지")는 **자격 술어가 아니라 `EntryDecisionID != ""` 명시 비교로 좁혀** 유지한다(편입 포지션에 fold가 착지하는 것은 정상 — 재대사 수량 비교 경로). `ExternalPositionAlert.ExitEligible` 하드코딩 false도 자격 술어로 교체.
- 골든 목록 갱신 필수: schema_test의 wantTables, v6 테이블 목록 2곳.
- "원장 스키마 확장 규칙" 메인 요구사항은 v6 고정 서술이므로 position-ledger delta에 MODIFIED로 v7을 포함시킨다.
- rollback(§0.6): additive라 구버전 바이너리는 ErrSchemaTooNew 거부(기존 계약).

## A2. t0 의미론 — manage-forward (정직한 서술)

원가 기반 t0는 첫 관측 틱에서 원가 대비 ±pct 밴드 밖의 모든 수동 보유를 즉시 매도시킨다(라운드 1 P1). 따라서:

- **EntryPrice = 편입 판정에 쓴 관측가**, **InitialStop = EntryPrice × (1 − pct)**, HighWater는 OpenRatchetState가 entry로 자동 seed(별도 인자 없음 — 라운드 2 정정). 편입 직후 R=0.
- **편입 행위 자체는 매도 발의를 생성하지 않는다(SHALL NOT)** — 편입 트랜잭션에 exit 판정이 포함되지 않는다. 다만 **편입 관측과 첫 exit 관측은 다른 시각·다른 소스의 값이므로, 그 사이 가격 이동이 첫 틱 발의를 만드는 것은 정상 exit 동작이다**(라운드 2: 절대 무발의 주장은 성립 불가 — 정직하게 명시). 회귀 테스트는 두 케이스를 나눈다: 편입 tx 무발의 + 첫 틱 발의는 래칫 규칙 그대로.
- `observed_price`의 소스는 **exit 관측과 동일한 시세 경로(Prices)**로 지정한다. 이 경로는 float64를 경유한다 — `[기존 제약 — 엔진 가격 경로 전체가 float64(Quote.Last)]` 태그. 원문 decimal 보존 SHALL은 **cost_basis에 한정**하며, 이를 위해 official 어댑터에 holdings 원문 문자열 접근을 additive로 추가한다(태스크).
- **pct 범위: `0.02 ≤ pct < 1`**(설정 거부). 하한 근거: 관측 노이즈·왕복 비용 규모(비용 상한 MAX_RATE=0.05)보다 작은 보호폭은 즉발 청산 장치가 된다 — provenance로 기록. 극소 pct의 formatPrice 반올림으로 stop==entry가 되는 경계는 riskOf가 fail-closed 거부(이중 방어).
- **귀결의 명시**(라운드 2 P2): +0.8R 도달 시 기준선은 **편입가 기준 실질 본전**으로 승격된다 — 편입은 "편입일 가격 + 비용"을 사실상의 보호 바닥으로 만든다. 원가 대비 큰 이익 중인 장기 보유가 편입 후 정상 되돌림으로 청산될 수 있다(원가 이익은 보존된 채로). 사용자 대면 설명에 포함.

## A3. 활성화 — 기본 OFF, flip은 사람 승인

`adoption.enabled` 기본 false(§0.2 zero-value 안전). true 전환은 §0.5 audit + §0.7 사람 승인(게이트 ON과 별도 항목). `exclude_symbols`는 enabled 안의 세밀 제어.

## A4. 알림 — 무관리 보유 알림은 무조건 존치 (라운드 2 P1 회귀 수정)

**exit 관리 자격 없는 보유의 발견 알림은 `adoption.enabled` 값과 무관하게 유지된다(SHALL)** — landed 동작("엔진이 보호하지 않을 포지션 옆에서 거래 중임을 운영자가 알아야 한다")이 §0.2의 기존 동작이다. enabled=true에서 그 알림은 편입 성공 이벤트로 **대체**되고, 제외 목록·편입 실패 시에는 알림이 남는다. 무알림은 **전이 상태**(RECONCILE 대기·인터록 미검증·Stabiliser 미수렴)뿐 — 같은 보유가 전이 상태를 벗어나 무관리로 확정되면 알림된다. (구현 노트: landed alertUnmanaged는 포지션당 in-memory 래치로 첫 관측에 무조건 발화한다 — 전이 상태 구분은 새 구동 루프가 함께 구현한다.)

## A5. 긴급 중지의 정직한 서술

kill switch는 BLOCK-ONLY이고 모든 모드가 RISK_REDUCING을 허용한다 — **편입 포지션의 자동 매도를 멈추는 스위치는 의도적으로 없다**(§0.3: "exit 일시중지"는 그 자체가 위반). 가용 수단: 사전 exclude, enabled=false(신규 편입 중지), flatten, 프로세스 종료. 편입 해제는 PROTECTION_WEAKENING 성격 — 범위 밖.

## A6. 관측 소스·비용 계상·실행 위치 (라운드 2 재작성)

reconcile 구동 루프를 신설한다(프로덕션 호출자 0인 Ingest/Converge/Tracker에 구동자). **1회 수집 = 전체 스냅샷**(메인 스펙 SHALL — 고정 순서·부분 실패 폐기): 미체결 페이지네이션(≤ MaxPages 50) + holdings 1 + 통화별 잔고 N. **Stabiliser 판정에는 수집 2회**(최소 2초 간격). 주기 60초(재대사 최소 간격 30초 요구사항 충족 — 이 루프가 곧 주기적 재대사 절차다). §0.4 계상: 정상 상태(미체결 0~1페이지·통화 1)에서 수집당 3콜×2회 수집 + 편입 후보 발생 시 심볼당 시세 1콜, 상한은 MaxPages가 계상 한계 — 수치를 스펙에 고정. 루프는 `Tracker.Observe`까지 구동한다(확정 하한 캡은 Tracker block 위에서만 동작 — exitwiring.go:151-155). 실행 술어는 `AutomationStatus.Verified`. Stabiliser 미수렴(사용자가 계속 수동 매매) 시 편입은 무기한 연기되며 이는 fail-closed 의도 동작이다(명시).

## A7. 편입 포지션의 성과 동결 (라운드 2 P1 재설계)

landed 동결 경로는 `entry_decision_id == ""`에서 조기 반환한다(trade_outcomes.go:175-179) — 확장 없이는 편입 포지션이 성과 행을 절대 만들지 않는다. 결정:

- **엔진이 청산을 실행한 편입 포지션은 성과 행을 만든다(SHALL)**: computeTradeOutcome에 편입 분기 신설 — 매수 leg은 fill이 아니라 `position_adoptions`에서 합성하되 **기준가는 `observed_price`로 단일화한다**(라운드 3: 분자·분모의 t0 epoch 일치 — cost_basis를 분자에 쓰면 "원가→매도 생애 손익 ÷ 편입일 위험"이라는 귀속 오류가 된다; cost_basis는 기록·표시용으로만 남고 성과 산식에서 제외, cost_basis_src 분기 삭제). 매도 leg은 **명시 조인**(`exit_events.proposed_intent_id → mutation_attempts → broker_order_id → fill_events` — 시간창 휴리스틱 금지, provenance CTE 체인 재사용)으로 모은다(SHALL). realized_r 분모는 exit_state의 합성 initial_risk. 스키마 무변경 — 합성 구분은 `positions.adoption_id` 조인.
- **외부 매도(조정)로 수량 0이 된 경우는 성과 행을 만들지 않는다(SHALL NOT)** — 돈이 엔진 밖에서 움직였고 매도 leg이 없다. 대신 exit_state completed 처리 + ADJUSTMENT_CLOSED exit_event 기록 + 알림(라운드 2 P1-3의 provenance 컬럼 충돌은 이 결정으로 소멸 — trade_outcomes에 provenance가 필요 없다). 이 규칙은 엔진·편입 포지션 공통이다(고아 exit_state 방지).

## A8. 편입 이후의 외부 개입

- 외부 수량 증가: exit_states t0·initial_risk·initial_quantity 동결 유지, 감지 시 알림(재산정 없음 — 후속 change 후보).
- 외부 부분 매도: taken_ratio_total 비율 회계의 완전 정합은 후속 change(기존 공통 갭). 수량 0 도달만 A7 규칙으로 닫는다.
- 재편입: CLOSED 후 재매수 → 새 인스턴스 → 재편입은 의도된 동작.

## A9. 순서 — 대시보드와의 조율 (라운드 2 순환 의존 해소)

대시보드는 **entry_decision_id 자격만으로 먼저 landed**된다(adoption_id를 알지 못한다). 이 change가 v7 착지 후 **task 2.7로 대시보드 자격 표시를 편입 포함으로 확장**한다(`internal/console` 소폭 수정 — 대시보드 landed 후이므로 충돌 없음; "콘솔 무접촉" 원칙의 유일한 예외로 명시). `internal/journal` 작업은 여전히 대시보드 조각(RO open·계좌 질의) landed 후 시작한다.
