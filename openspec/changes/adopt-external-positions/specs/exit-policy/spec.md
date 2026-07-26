# exit-policy Specification (delta)

## MODIFIED Requirements

### Requirement: t0 기준선 — 진입 손절가가 첫 독자를 갖는다
포지션이 열리는 순간 보호 기준선은 그 포지션을 정당화하는 기록의 손절가로 초기화되어야 한다(SHALL — `exit_states.baseline_price`는 NOT NULL이며 빈 기본값이 없다): 엔진 진입 포지션은 진입 결정의 손절가(RiskIntent.stop), 편입 포지션은 편입 기록의 합성 손절가("외부 취득 포지션의 자동 편입" 요구사항)다. exit_state 개설 경로는 두 출처를 모두 수용해야 한다(SHALL — 진입 preimage 단일 타입 가정 금지). 기준선 하회 발의 규칙은 t0부터 유효하다(SHALL) — "No Stop = No Trade"의 손절가는 사전 검사 입력에 그치지 않고 체결 후 관측 판정의 기준선이 된다. entry 결정도 편입 기록도 없는 포지션은 exit 정책의 대상이 아니다(SHALL — R·기준선의 근거가 없다).

#### Scenario: 개시 직후 손절 하회
- **WHEN** 진입 체결 직후(+0.4R 도달 전) 관측 가격이 진입 손절가를 하회하면
- **THEN** 전량 청산이 RISK_REDUCING 의도로 발의된다

#### Scenario: 편입 후 외부 포지션
- **WHEN** 편입 기록이 영속된 외부 포지션의 관측 가격이 합성 손절가를 하회하면
- **THEN** 엔진 진입 포지션과 동일하게 전량 청산이 RISK_REDUCING 의도로 발의된다

#### Scenario: 어떤 기록도 없는 포지션
- **WHEN** entry 결정도 편입 기록도 없는 포지션이 존재하면
- **THEN** exit 판정이 수행되지 않는다 (알림 규칙은 편입 요구사항이 정의)

## ADDED Requirements

### Requirement: 외부 취득 포지션의 자동 편입
`adoption.enabled`가 참인 엔진은 기록이 정당화하지 않는 보유(외부·수동 취득)를 발견하면 exit 관리에 편입해야 한다(SHALL — 사용자 결정 2026-07-26). `adoption.enabled`의 기본값은 false이며(SHALL — §0.2 zero-value 안전), true 전환은 audit 기록과 사람의 직접 승인을 요한다(SHALL — §0.5·§0.7). 편입은 영속이 선행한다(SHALL — decide→persist→execute): `position_adoptions` 기록(심볼·시장·수량·원가와 그 출처 표기·편입 시점 관측가·합성 손절가·관측 시각·digest)이 journal에 영속되고 `positions.adoption_id`가 이를 가리킨 뒤에만 exit_state가 열린다.

**manage-forward t0**(SHALL): EntryPrice와 워터마크 seed는 **편입 시점 관측가**이고, InitialStop은 `관측가 × (1 − adoption.default_stop_pct)`다. 원가는 기록·분석용이며 R 분모로 쓰지 않는다(SHALL NOT). 따라서 편입 직후 R=0이며 **편입 자체는 어떤 매도 발의도 유발하지 않는다**(SHALL NOT — 첫 관측 틱 포함). `adoption.default_stop_pct`는 `0 < pct < 1` 범위를 벗어나면 설정 거부다(SHALL). 기본값은 보수적으로 정하고 출처·산정 근거를 provenance로 기록한다(SHALL — 임의 수치 금지).

편입 조건(SHALL): `AutomationStatus.Verified`인 엔진에서, 해당 심볼이 RECONCILE 상태가 아니고, 신선한 보유 확인 — Stabiliser 통과(최소 2초 간격 연속 2회 동일 스냅샷) + 관측 staleness 10초 이내 — 을 만족할 때만. `adoption.exclude_symbols`의 심볼은 편입하지 않는다(SHALL NOT). 편입 이후 래칫·부분익절·관측·pending·staleness 규칙은 엔진 진입 포지션과 동일 규칙이 적용된다(SHALL — 별도 축소 경로 금지; ladder는 정책 배정 경로가 생기면 동일 적용). CLOSED 후 재매수로 생긴 새 인스턴스의 재편입은 의도된 동작이다(SHALL 명시).

알림(SHALL — 이 규칙이 종전 "외부 포지션 발견" 알림을 대체한다): 제외 목록 심볼의 무결정 보유 발견과 편입 실패 시에만 발송한다. 정상 지연 상태(enabled=false·RECONCILE 대기·인터록 미검증)는 알림하지 않는다(SHALL NOT — enabled=false는 대시보드 표시로 드러난다). 편입 성공은 편입가·합성 손절을 담은 이벤트로 기록·통지된다(SHALL). 편입은 매수·사이징을 변경하지 않는다(SHALL NOT).

#### Scenario: 수동 매수 종목의 자동 편입 — 즉시 매도 없음
- **WHEN** enabled=true 엔진에서 수동 매수 심볼이 신선 조건을 충족해 편입되면
- **THEN** exit_state가 관측가 기준 t0(R=0)로 열리고, 그 관측 틱에서 어떤 매도 발의도 생성되지 않으며, 편입 이벤트가 통지된다

#### Scenario: 원가와 관측가가 크게 다른 장기 보유
- **WHEN** 원가 대비 +50%인 보유가 편입되면
- **THEN** EntryPrice는 관측가, 워터마크도 관측가로 seed되어 부분익절·기준선 상승은 편입 이후 상승분부터 시작한다

#### Scenario: RECONCILE 중 편입 시도
- **WHEN** 심볼이 RECONCILE 상태일 때 무결정 보유가 관측되면
- **THEN** 편입도 알림도 일어나지 않고 해소 후 다음 관측에서 재평가된다

#### Scenario: 제외 목록 심볼
- **WHEN** `adoption.exclude_symbols`에 있는 심볼의 무결정 보유가 관측되면
- **THEN** 편입 없이 운영자 알림만 발송된다

#### Scenario: 편입 기록 영속 후 크래시
- **WHEN** position_adoptions 영속 후 exit_state open 전에 크래시·재시작되면
- **THEN** 재시작 복구가 영속된 기록으로부터 exit_state open을 완결하고 중복 편입은 발생하지 않는다

#### Scenario: 범위 밖 손절폭 설정
- **WHEN** `adoption.default_stop_pct`가 0 이하 또는 1 이상으로 설정되면
- **THEN** 설정이 거부되고 편입은 전면 비활성으로 남는다
