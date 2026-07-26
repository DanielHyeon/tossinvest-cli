# Change: adopt-external-positions

## Why

사용자 결정(2026-07-26): **사용자가 수동 매수한 종목에도 익절·손절·기준선 수익 극대화(래칫)를 자동 적용**할 것. 현행 계약(2d)은 반대다 — exit-policy·position-ledger·reconciliation 메인 스펙이 "entry 결정 없는 포지션(외부·수동 편입)은 exit 정책의 대상이 아니다"라고 명시하고 알림만 발송한다. 근거는 "R·기준선의 근거(진입 손절가)가 없다"였다. 이 change는 그 근거를 **영속된 편입 기록**으로 만들어 공급한다.

## What Changes

- **편입 = 영속된 기록**: 무결정 보유를 발견하면 `position_adoptions` 행(심볼·수량·원가와 출처·편입 시점 관측가·합성 손절·관측 시각·digest)을 journal에 영속하고, `positions.adoption_id`(set-once, additive v7)가 이를 가리킨다. exit 자격은 `entry_decision_id 또는 adoption_id`로 확장된다 — `entry_decision_id`는 읽기만 하며(불변 유지), **decisions 테이블·safety class 3종은 무접촉**(편입은 mutation이 아니다 — design A1).
- **manage-forward t0**(design A2): EntryPrice = **편입 트랜잭션 직전의 신선한 시세 관측**(staleness ≤ 15s, 가격 경로는 `[기존 제약 float64]`), InitialStop = 관측가×(1−`adoption.default_stop_pct`, 범위 0.02≤pct<1). 편입 직후 R=0 — **편입 행위 자체는 매도 발의를 생성하지 않으며**, 래칫·부분익절은 편입 이후의 움직임부터 작동한다(편입 관측과 첫 exit 관측 사이의 가격 이동 발의는 정상 exit 동작). 원가는 기록·분석용. **귀결의 정직한 명시**: +0.8R부터 기준선이 편입가 기준 실질 본전으로 승격되므로, 편입은 "편입일 가격+비용"을 사실상의 보호 바닥으로 만든다 — 원가 대비 큰 이익 중인 장기 보유는 편입 후 정상 되돌림에 청산될 수 있다(원가 이익은 보존된 채로).
- **기본 OFF**: `adoption.enabled` 기본 false(§0.2 — false의 동작은 무관리 보유 알림을 포함한 기존 동작과 동일), true 전환은 §0.5 audit + §0.7 사람 승인. `exclude_symbols`(기본 빈 목록)는 enabled 안의 세밀 제어.
- **관측 소스 신설**(design A6): 엔진 reconcile 구동 루프 — 전체 스냅샷(미체결 pagination+holdings+잔고)×Stabiliser 2회/주기 60초, Tracker.Observe 포함, §0.4 계상 — 현재 프로덕션 호출자가 없는 Ingest/Converge/Tracker 경로에 구동자를 만든다.
- **성과 동결 확장**(design A7): 엔진이 청산한 편입 포지션은 trade_outcomes 행을 만든다(매수 leg은 편입 기록에서 합성). 외부 매도로 닫힌 포지션은 성과 행 없이 completed+이벤트+알림.
- **무관리 보유 알림은 enabled와 무관하게 존치**(design A4 — §0.2). enabled=true의 편입 성공 이벤트가 그 알림을 대체하고, 전이 상태만 무알림.

## Non-Goals

- 매수(진입) 자동화 변경 — 편입은 이미 존재하는 보유에 보호를 붙일 뿐이다
- 손절 즉시성·사이징 규칙 변경(§0.3·§0.9 무접촉)
- **편입 해제**(보호 제거 = PROTECTION_WEAKENING 성격 — 범위 밖, design A5)
- 브로커측 보호주문(2c 소관 — 편입 포지션도 2c 계약을 동일하게 받는다)
- 외부 부분 매도의 비율 회계 완전 정합(기존 공통 갭 — 고아 exit_state 방지만 이 change에서, design A8)

## 긴급 중지에 대한 정직한 서술 (design A5)

kill switch는 신규 진입 차단 전용(BLOCK-ONLY)이며 **편입 포지션의 자동 매도를 멈추는 스위치는 의도적으로 존재하지 않는다** — "exit 일시중지"는 §0.3(손절 즉시성 약화 금지) 위반이기 때문이다. 가용 수단은 사전 exclude, `adoption.enabled=false`(신규 편입 중지), flatten, 프로세스 종료다. flatten은 편입 포지션을 동일하게 덮는다(감축이므로).

## Capabilities

### New Capabilities

(없음)

### Modified Capabilities

- `exit-policy`: "t0 기준선" — 외부 포지션 비대상 조항을 편입 계약으로 대체 + 편입 요구사항 추가
- `position-ledger`: "Position 투영과 단일 권위" — adoption_id 경로·set-once·조정 종결 규칙
- `reconciliation`: "Reconciliation 계약" — 외부 보유의 편입 경로와 구동 루프
- `trade-analytics`: 합성 R 구분 집계

## Impact

- Affected code: `internal/journal`(v7: position_adoptions·positions.adoption_id — additive만), 엔진 reconcile 구동 루프 + 편입 파이프라인, exit 자격 술어 단일화, config(enabled·default_stop_pct·exclude_symbols), `internal/console`(task 2.7 자격 표시 확장 — 대시보드 landed 후)
- **실효 시점**: 게이트 ON + `adoption.enabled` 승인 이후. 코드는 landed 후에도 기본 OFF
- **순서**(design A9): add-operator-dashboard의 journal 조각(RO open·계좌 질의) landed 후 구현 착수 — `internal/journal` 동시 작업 금지
- §0 검토: 편입은 보호 추가만(보수 방향), 즉시 매도 없음(A2), 기본 OFF(§0.2), flip 사람 승인(§0.7), 매도는 전부 RISK_REDUCING 기존 레일 경유
