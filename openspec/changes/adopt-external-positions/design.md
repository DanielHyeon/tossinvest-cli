# Design: adopt-external-positions

> 리뷰 라운드 1(2026-07-27, P1 8건)의 결정 사항. High-risk 경로(원장 스키마·exit 자격 게이트)이므로 이 문서가 테이블 정의·의미론의 정본이다.

## D1. 편입의 저장 형태 — 별도 테이블, 결정 테이블 무접촉

`decisions.safety_class`·`preimage_kind`는 SQL CHECK enum이고 additive 규칙이 테이블 재작성을 금지한다(execution_contract.go:144-147, schema.go 규칙). **편입은 mutation이 아니므로 class 축에 놓지 않는다** — engine-safety·risk-management의 3-class 계약은 무변경(delta 불필요).

**v7 마이그레이션(additive만)**:

```sql
CREATE TABLE position_adoptions (
    id              TEXT PRIMARY KEY,          -- adoption id
    position_id     TEXT NOT NULL REFERENCES positions(id),
    symbol          TEXT NOT NULL,
    market          TEXT NOT NULL,
    quantity        TEXT NOT NULL,             -- decimal string, 편입 시점 관측 수량
    cost_basis      TEXT,                      -- 브로커 averagePurchasePrice 원문(있으면), 분석용 — R 분모 아님
    cost_basis_src  TEXT NOT NULL,             -- 'BROKER_AVG' | 'ABSENT'
    observed_price  TEXT NOT NULL,             -- 편입 시점 관측가 = t0 EntryPrice
    synthetic_stop  TEXT NOT NULL,             -- observed_price × (1 − default_stop_pct)
    observed_at     TEXT NOT NULL,
    preimage_digest TEXT NOT NULL              -- 위 필드 정본의 digest (재검증용)
) STRICT;
ALTER TABLE positions ADD COLUMN adoption_id TEXT REFERENCES position_adoptions(id);
```

- `positions.adoption_id`는 **set-once**: 전용 tx API로만 기입하고, 그 외 UPDATE 문이 이 컬럼을 언급하면 실패하는 정적 스캔 가드(guarded-column 스캔과 같은 방식, 소유 파일 화이트리스트)를 둔다. `entry_decision_id`는 이 change에서 **읽기만** 한다(첫 변이자 도입 금지 — 리뷰 P1-5).
- exit 자격 게이트 확장: 대상 = `entry_decision_id IS NOT NULL OR adoption_id IS NOT NULL`. 게이트를 읽는 전 지점(position/provenance.go, exitloop.go 열거, exit_state open)을 하나의 술어 함수로 모아 drift를 막는다.
- rollback(§0.6): v7은 additive라 구버전 바이너리는 `ErrSchemaTooNew`로 거부된다(기존 계약 그대로).

## D2. t0 의미론 — manage-forward (편입은 즉시 매도를 유발하지 않는다)

원가 기반 t0(EntryPrice=평단)는 첫 관측 틱에서 원가 대비 ±pct 밴드 밖의 모든 수동 보유를 즉시 매도시킨다(전량 손절 또는 40% 부분익절 — 리뷰 P1-1, ratchet.go:364-478 항등식). 이는 §0.9 보수 방향이 아니다. 따라서:

- **EntryPrice = 편입 시점 관측가**(신선 조건 D5 충족), **InitialStop = EntryPrice × (1 − pct)**, **HighWater seed = EntryPrice**. 편입 직후 R=0 — 래칫·부분익절은 편입 이후의 상승분에 대해서만 작동한다.
- 원가(`cost_basis`)는 기록·분석용이다. R 분모로 쓰지 않는다.
- 귀결: 오래 보유한 종목의 기존 이익·손실은 편입 시점에 실현되지 않고, 이후 움직임부터 보호·극대화가 시작된다. 사용자가 당일 매수한 종목은 관측가≈매수가이므로 사실상 원가 기준과 같다.
- `0 < adoption.default_stop_pct < 1` 범위 검증 실패는 설정 거부다(pct=0 → risk=0 나눗셈 경로, pct≥1 → stop≤0 거부 — ratchet.go:581-598).

## D3. 활성화 — 기본 OFF, flip은 사람 승인

`adoption.enabled` 기본 false(§0.2 — zero-value가 안전값, config/engine.go 선례). true 전환은 §0.5 audit + §0.7 사람 승인 대상(게이트 ON과 별도 항목으로 명시). `exclude_symbols`는 enabled=true 안에서의 세밀 제어다.

## D4. 긴급 중지의 정직한 서술

kill switch는 BLOCK-ONLY이고 모든 모드가 RISK_REDUCING을 허용한다(operating_mode.go:217-231) — **편입 포지션의 자동 매도를 멈추는 스위치는 의도적으로 없다**(§0.3: 손절 즉시성은 약화 금지 — "exit 일시중지"는 그 자체가 §0.3 위반이다). 편입 해제(보호 제거)는 PROTECTION_WEAKENING 성격이므로 이 change 범위 밖(잔존: flatten, 프로세스 종료, 사전 exclude). proposal의 "kill switch가 덮는다" 문구는 이 서술로 교체.

## D5. 관측 소스·신선도·실행 위치

reconcile fold(IngestExternalPositions/ConvergeQuantities)에는 프로덕션 구동 루프가 없다(리뷰 P1-4 — 비테스트 호출자 0). 이 change가 **엔진 reconcile 구동 루프**(브로커 스냅샷 수집→Stabiliser→fold→편입 후보 판정)를 신설한다. §0.4 계상: 스냅샷은 holdings 1콜/주기(주기는 exit 관측보다 느리게, 기본 60초), Stabiliser(최소 2초 간격 연속 2회 동일 — snapshot.go:255-338) 통과 + 관측 시각 staleness ≤ 10초(riskcalc.AccountSnapshotStaleness)일 때만 "신선한 보유 확인"으로 인정. 실행 조건 술어는 `AutomationStatus.Verified`(기동 인터록 통과 엔진 — 런타임 별도 래치 없음)다.

## D6. 편입 이후의 외부 개입

- **외부 수량 증가**(사용자 추가 매수): exit_states의 t0·initial_risk·initial_quantity는 동결 유지(메인 스펙 계약), 증가분 감지 시 알림 발송 — 재편입·재산정은 하지 않는다(후속 change 후보로 기록).
- **외부 부분 매도**: 조정 경로는 `taken_ratio_total`을 이동시키지 않는다(landed — apply_hook.go:368 유일 writer). 편입·엔진 포지션 공통의 기존 갭이므로 이 change에서 다음만 닫는다: 조정으로 수량이 0이 되면 exit_state를 completed 처리하고 trade_outcome을 ADJUSTMENT_CLOSED provenance로 동결(고아 행 방지). 비율 회계의 완전한 정합은 후속 change로 기록.
- **재편입 루프**: CLOSED 후 재매수 → 새 instance → 재편입은 **의도된 동작**이다(명시).

## D7. 분석 구분 — 합성 R

편입 포지션의 realized_r은 합성 분모에서 나온다. trade_outcomes 스키마는 무변경 — 구분은 `positions.adoption_id IS NOT NULL` 조인으로 한다(명시 조인, 휴리스틱 아님 — position-ledger 계약 준수). trade-analytics delta: 집계는 measured/synthetic을 분리 표기해야 한다(SHALL).

## D8. 알림 규칙 통일

알림은 **제외 목록 심볼의 무결정 보유 발견**과 **편입 실패**에만. 정상 지연 상태(RECONCILE 대기·인터록 미검증·enabled=false)는 무알림 — enabled=false는 대시보드 표시로만 드러난다.

## D9. 순서 제약

`internal/journal` 파일 표면이 add-operator-dashboard(RO open·계좌 단위 질의)와 겹친다. **대시보드의 journal 조각이 먼저 landed된 뒤** 이 change의 구현을 시작한다(동시 작업 금지 — WORKFLOW 병렬 규칙).
