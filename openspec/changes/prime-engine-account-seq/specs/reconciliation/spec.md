## MODIFIED Requirements

### Requirement: 재시작 복구
프로세스 재시작 시 엔진은 journal의 미확정 intent 해소 → 계좌·미체결·체결 조회 → 로컬 상태 재구성 순서의 복구를 완료한 후에만 신규 주문을 허용해야 한다(SHALL). 엔진은 공식 계좌 목록의 첫 레코드에 비어 있지 않은 계좌번호와 양수 account sequence가 함께 있을 때만 그 레코드를 기본 계좌로 해석해야 하며(SHALL), 둘 중 하나라도 유효하지 않으면 뒤 레코드로 건너뛰지 않고 기동을 거부해야 한다(SHALL). 클라이언트에 양수 sequence가 명시되어 있으면 엔진은 그 값이 첫 레코드 sequence와 정확히 같을 때만 기동해야 한다(SHALL). 기동이 이 레코드를 성공적으로 해석했다면 같은 공식 클라이언트의 재시작 복구는 동일 레코드의 sequence를 재사용하고 첫 account-scoped 조회 전에 계좌 목록을 다시 요청해서는 안 된다(SHALL NOT).

#### Scenario: 복구 완료 전 주문 시도
- **WHEN** 복구 절차가 완료되기 전에 신규 주문 요청이 발생하면
- **THEN** 요청은 거부되고 복구 미완료 사유가 반환된다

#### Scenario: 기동 계좌 해석의 sequence 재사용
- **WHEN** 엔진 기동 계좌 해석이 한 번 성공한 뒤 재시작 복구가 첫 account-scoped 스냅샷 조회를 수행하면
- **THEN** 같은 sequence가 `X-Tossinvest-Account`에 사용되고 `/api/v1/accounts`는 다시 호출되지 않는다

#### Scenario: 엔진의 명시된 sequence 일치
- **WHEN** 엔진의 공식 클라이언트가 명시적 account sequence 7로 구성되고 첫 계좌 레코드 sequence도 7이면
- **THEN** 기동과 이후 account-scoped 조회는 같은 계좌번호와 sequence를 유지한다

#### Scenario: 엔진의 명시된 sequence 불일치
- **WHEN** 엔진의 공식 클라이언트가 명시적 account sequence 99로 구성되고 첫 계좌 레코드 sequence가 7이면
- **THEN** 엔진은 journal을 열기 전에 계좌 해석 실패로 기동을 거부한다

#### Scenario: 첫 계좌 레코드가 불완전함
- **WHEN** 공식 계좌 목록의 첫 레코드에 계좌번호가 없거나 sequence가 양수가 아니고 뒤 레코드는 유효하면
- **THEN** 엔진은 뒤 레코드를 암묵적으로 선택하지 않고 계좌 해석 실패로 기동을 거부한다
