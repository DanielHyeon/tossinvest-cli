## ADDED Requirements

### Requirement: 서버 권위 공통 정책 레지스트리
시스템은 `COMMON_LADDER_BALANCED`, `COMMON_LADDER_RUNNER`, `COMMON_LADDER_HYBRID_50`의 immutable decimal 정책을 등록하고, 등록되지 않은 policy ID를 설정하거나 평가해서는 안 된다 (SHALL NOT).

#### Scenario: HYBRID_50 정책 수치
- **WHEN** 운영자가 HYBRID_50 정책 상세를 조회한다
- **THEN** 목표는 1.8/3.0/4.8/6.5%, 보호선은 0/1.2/2.5/3.8%, 잔량 기준 부분익절은 0/0.25/1/3/0으로 표시된다

#### Scenario: 알 수 없는 정책 저장
- **WHEN** 클라이언트가 등록되지 않은 공통 policy ID를 제출한다
- **THEN** 전체 저장이 거부되고 기존 config와 audit의 effective value는 바뀌지 않는다

### Requirement: 공통 정책은 명시적으로 승인된다
시스템은 HYBRID_50을 권장값으로 표시하되 migration, 설치 또는 기동만으로 공통 정책을 채우거나 변경해서는 안 된다 (MUST NOT).

#### Scenario: 업그레이드 직후 미승인
- **WHEN** 기존 config에 `engine.exit_policy.common_policy`가 없다
- **THEN** 기존 RATCHET 동작이 유지되고 최적화 화면은 공통 정책 승인이 필요하다고 표시한다

#### Scenario: 운영자가 HYBRID_50 승인
- **WHEN** 인증된 운영자가 최적화 화면에서 HYBRID_50을 선택하고 CSRF가 포함된 저장을 제출한다
- **THEN** ID만 config에 저장되고 변경 전후 값이 audit에 기록되며 다음 엔진 기동부터 신규 관리 포지션에 적용된다

### Requirement: HYBRID_50은 절반을 남겨 runner로 보호한다
HYBRID_50은 T2에서 잔량의 25%, T3에서 잔량의 1/3을 부분익절하고 T4에서 고정 전량익절 없이 남은 수량을 high-water runner로 관리해야 한다 (SHALL).

#### Scenario: 원수량 100주의 부분익절
- **WHEN** 100주 포지션의 T2와 T3 부분익절이 rounding 손실 없이 체결된다
- **THEN** 약 50주가 남고 T4는 전량매도 proposal을 만들지 않는다

#### Scenario: T4 이후 신고가
- **WHEN** HYBRID_50이 T4에 도달한 뒤 high-water가 상승한다
- **THEN** 보호선 후보는 `high_water × 0.935`로 상승하고 저장된 보호선보다 내려가지 않는다

#### Scenario: runner 보호선 이탈
- **WHEN** 현재가가 가장 높은 유효 ladder/trailing 보호선 아래로 내려간다
- **THEN** 기존 cancel-first·Guardian reduce-only·idempotent submission 경로로 잔량 전량 청산을 제안한다

### Requirement: 활성 포지션은 정책을 스냅샷한다
시스템은 exit 관리 시작 시 선택된 policy kind와 policy ID를 포지션의 exit state에 영속하고, 공통값 변경만으로 활성 포지션을 다른 정책으로 재해석해서는 안 된다 (SHALL NOT).

#### Scenario: 공통값 변경 전 열린 포지션
- **WHEN** 활성 포지션이 BALANCED snapshot을 가진 상태에서 공통값을 HYBRID_50으로 변경한다
- **THEN** 그 포지션은 BALANCED를 계속 사용하고 새로 열리는 exit state만 HYBRID_50을 사용한다

#### Scenario: legacy ladder row
- **WHEN** migration 전 LADDER row의 policy ID가 NULL이다
- **THEN** 시스템은 historical row를 rewrite하지 않고 그 row만 기존 `default_v1` BALANCED 의미로 읽는다

### Requirement: 외부 구매분도 편입 시 정책을 고정한다
시스템은 외부 구매 포지션을 편입할 때 관측 가격, synthetic stop, 승인된 common policy ID를 adoption record에 함께 저장하고 그 record로 exit state를 열어야 한다 (SHALL).

#### Scenario: HYBRID_50 외부 편입
- **WHEN** HYBRID_50이 승인된 상태에서 broker-confirmed 외부 보유분을 편입한다
- **THEN** adoption 관측 가격을 entry/high-water t0로 하고 HYBRID_50 policy ID를 exit state에 기록한다

#### Scenario: 편입 기록 뒤 crash
- **WHEN** adoption commit 뒤 exit state 생성 전에 process가 종료되고 그 사이 공통 설정이 바뀐다
- **THEN** 복구는 새 config가 아니라 adoption record에 저장된 policy ID로 exit state를 연다

#### Scenario: RUNNER 외부 편입
- **WHEN** RUNNER가 승인된 상태에서 broker-confirmed 외부 보유분을 편입하고 목표 rung에 도달한다
- **THEN** 동일한 RUNNER 보호선은 승격되지만 자동 부분익절 proposal은 생성되지 않고 breach 시에만 잔량 전량보호를 제안한다

### Requirement: 정책 설정은 주문 권한을 만들지 않는다
공통 정책의 조회·저장은 broker 호출, 주문 생성, automation gate 변경, trading toggle 변경 또는 기존 포지션 rebind를 수행해서는 안 된다 (MUST NOT).

#### Scenario: 정책만 저장
- **WHEN** 운영자가 공통 정책을 저장한다
- **THEN** broker call count와 주문 수는 0이고 gate/trading/adoption 설정의 기존 바이트는 보존된다
