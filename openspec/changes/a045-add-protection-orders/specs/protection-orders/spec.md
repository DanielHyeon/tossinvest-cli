## ADDED Requirements

### Requirement: 지원되는 포지션은 브로커 상주 손절로 보호된다
시스템은 attested market/type/session/quantity capability에 해당하는 체결 노출에 대해 broker-resident stop을 등록하고 ACTIVE를 확인하기 전 추가 exposure-raising 자동화를 허용해서는 안 된다 (MUST NOT).

#### Scenario: 신규 진입 체결
- **WHEN** 자동 진입이 일부 또는 전부 체결된다
- **THEN** 체결 수량을 덮는 보호 saga를 영속하고 공식 조건주문 endpoint로 손절을 등록한다

#### Scenario: 보호 등록 실패
- **WHEN** 보호주문이 거부되거나 결과가 불명확하다
- **THEN** 신규 진입 latch를 닫고 RECONCILE·운영 경고·명시된 flatten 정책으로 전환한다

### Requirement: 한 심볼의 매도 청구권은 중복되지 않는다
시스템은 조건주문 예약, 일반 미체결 매도와 local reservation을 함께 계산해 보유 수량을 넘는 매도 청구권을 만들어서는 안 된다 (MUST NOT).

#### Scenario: 부분체결 증가
- **WHEN** 진입 체결 수량이 1주에서 3주로 증가한다
- **THEN** 보호 수량을 3주로 수렴시키되 교체 window에서 합산 매도 청구권이 3주를 넘지 않는다

#### Scenario: 1주 포지션
- **WHEN** 보유 수량이 정확히 1주다
- **THEN** broker protection quantity는 항상 1주이고 중간 부분익절 생략으로 보호가 제거되지 않는다

### Requirement: 보호선은 더 안전한 방향으로만 교체된다
ACTIVE protection의 trigger는 a041 현재 보호선 상승에 맞춰 올릴 수 있지만 낮출 수 없으며 (MUST NOT), replace의 ID·상태는 crash recovery 가능하게 기록돼야 한다 (SHALL).

#### Scenario: 낮은 교체 요청
- **WHEN** 새 trigger가 현재 ACTIVE trigger보다 낮다
- **THEN** broker mutation 없이 요청을 거부하고 현재 보호를 유지한다

#### Scenario: 프로세스 재시작
- **WHEN** 보호주문 ACTIVE 뒤 엔진이 재시작한다
- **THEN** broker ID를 조회·귀속하고 중복 조건주문을 새로 만들지 않는다
