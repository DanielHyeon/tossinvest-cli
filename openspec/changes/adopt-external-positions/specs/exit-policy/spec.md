# exit-policy Specification (delta)

## MODIFIED Requirements

### Requirement: t0 기준선 — 진입 손절가가 첫 독자를 갖는다
포지션이 열리는 순간 보호 기준선은 진입 결정의 손절가(RiskIntent.stop)로 초기화되어야 한다(SHALL — `exit_states.baseline_price`는 NOT NULL이며 빈 기본값이 없다). 기준선 하회 발의 규칙은 t0부터 유효하다(SHALL) — "No Stop = No Trade"의 손절가는 사전 검사 입력에 그치지 않고 체결 후 관측 판정의 기준선이 된다. entry 결정이 없는 포지션(외부·수동 취득)은 **편입 결정(ADOPTION)이 영속되기 전까지** exit 정책의 대상이 아니며(SHALL — R·기준선의 근거가 없다), 편입되면 편입 결정의 합성 손절가가 t0 기준선이 된다(SHALL — "외부 취득 포지션의 자동 편입" 요구사항). 편입이 제외·실패로 일어나지 않은 포지션의 발견은 운영자 알림을 발송한다(SHALL).

#### Scenario: 개시 직후 손절 하회
- **WHEN** 진입 체결 직후(+0.4R 도달 전) 관측 가격이 진입 손절가를 하회하면
- **THEN** 전량 청산이 RISK_REDUCING 의도로 발의된다

#### Scenario: 편입 전 외부 포지션
- **WHEN** 편입 결정이 아직 영속되지 않은(또는 제외 목록의) 외부 포지션이 존재하면
- **THEN** exit 판정이 수행되지 않고 운영자 알림이 발송된다

#### Scenario: 편입 후 외부 포지션
- **WHEN** ADOPTION 결정이 영속된 외부 포지션의 관측 가격이 합성 손절가를 하회하면
- **THEN** 엔진 진입 포지션과 동일하게 전량 청산이 RISK_REDUCING 의도로 발의된다

## ADDED Requirements

### Requirement: 외부 취득 포지션의 자동 편입
엔진은 결정이 정당화하지 않는 보유(외부·수동 취득)를 발견하면 자동으로 exit 관리에 편입해야 한다(SHALL — 사용자 결정 2026-07-26). 편입은 결정이다(SHALL): ADOPTION class의 결정을 전용 preimage(심볼·시장·수량·비용기준과 그 출처·합성 손절가·관측 시각)로 journal에 영속한 뒤에만 exit_state가 열리고, 포지션의 진입 결정 참조가 이 결정을 가리킨다 — "결정 없는 포지션은 exit 대상이 아니다"라는 불변식은 유지된다(SHALL, decide→persist→execute).

합성 t0(SHALL): EntryPrice는 브로커 평균단가, 알 수 없으면 편입 시점 관측가 — 어느 출처를 썼는지 preimage에 기록한다. InitialStop은 `EntryPrice × (1 − adoption.default_stop_pct)`이며 기본값은 보수적으로 정하고 출처·산정 근거를 코드 provenance로 기록한다(SHALL — 임의 수치 금지). 편입 이후 래칫·ladder·부분익절·관측·pending 수명주기는 엔진 진입 포지션과 동일 코드 경로다(SHALL — 별도 축소 경로를 만들지 않는다).

편입 조건(SHALL): Guardian 인터록 활성 + RECONCILE 상태 아님 + 신선한 보유 확인이 모두 충족될 때만. `adoption.exclude_symbols`(기본 빈 목록)에 있는 심볼은 편입하지 않고(SHALL NOT) 기존 알림 경로를 유지한다. 편입 성공은 편입가·합성 손절을 담은 이벤트로 기록·통지되고(SHALL), 이 이벤트가 종전의 "외부 포지션 발견" 알림을 대체한다. 편입은 매수·사이징을 변경하지 않는다(SHALL NOT — 보호 추가만, §0.9 보수 방향).

#### Scenario: 수동 매수 종목의 자동 편입
- **WHEN** 사용자가 수동 매수한 심볼이 신선한 보유 확인에 나타나고 인터록이 활성이며 RECONCILE이 아니면
- **THEN** ADOPTION 결정이 영속되고 exit_state가 합성 t0로 열리며 편입 이벤트가 통지된다

#### Scenario: RECONCILE 중 편입 시도
- **WHEN** 심볼이 RECONCILE 상태일 때 무결정 보유가 관측되면
- **THEN** 편입은 일어나지 않고 RECONCILE 해소 후 다음 관측에서 재평가된다

#### Scenario: 제외 목록 심볼
- **WHEN** `adoption.exclude_symbols`에 있는 심볼의 무결정 보유가 관측되면
- **THEN** 편입 없이 운영자 알림만 발송된다

#### Scenario: 편입 결정 영속 전 크래시
- **WHEN** ADOPTION 결정 영속 후 exit_state open 전에 크래시·재시작되면
- **THEN** 재시작 복구가 영속된 결정으로부터 exit_state open을 완결하고 중복 편입은 발생하지 않는다
