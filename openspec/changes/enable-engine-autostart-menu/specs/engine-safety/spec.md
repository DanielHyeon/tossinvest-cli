## ADDED Requirements

### Requirement: 승인된 엔진 자동 기동

`engine.autostart`는 기본 OFF여야 한다(SHALL). 콘솔 프로세스가 시작될 때 이
설정이 ON인 경우에만 기존 엔진 시작 경로를 정확히 한 번 호출해야 한다(SHALL).
자동 기동은 수동 [엔진 시작]과 동일한 journal flock, automation gate, Guardian,
capability attestation, 거래 정책, ExecutionGateway startup interlock을 사용해야
하며(SHALL), 그 중 어느 조건도 대신 설정하거나 우회해서는 안 된다(SHALL NOT).
설정 읽기 실패는 자동 기동을 생략하는 fail-closed 결과여야 한다(SHALL).

#### Scenario: 기본 설정과 구버전 설정
- **WHEN** 새 기본 config를 만들거나 `engine.autostart`가 없는 기존 config를 읽으면
- **THEN** autostart는 OFF이고 엔진 시작 호출이 발생하지 않는다

#### Scenario: 승인된 부팅 자동 기동
- **WHEN** `engine.autostart`가 ON인 config로 콘솔 프로세스가 시작되면
- **THEN** 기존 엔진 시작 seam이 정확히 한 번 호출되고 그 seam의 startup interlock 결과가 최종 기동 여부를 결정한다

#### Scenario: 자동 기동의 인터록 거부
- **WHEN** autostart는 ON이지만 automation gate 또는 기존 startup interlock 조건이 충족되지 않으면
- **THEN** 엔진은 기존 사유로 거부되고 콘솔은 계속 실행되며 실제 주문 경로는 열리지 않는다

#### Scenario: 설정 읽기 실패
- **WHEN** 콘솔 시작 시 autostart 설정을 읽을 수 없거나 JSON이 잘못되었으면
- **THEN** 엔진 시작 seam은 호출되지 않고 오류가 운영자에게 표시된다

#### Scenario: 부팅 중복 인스턴스
- **WHEN** autostart ON인 콘솔이 시작될 때 동일 journal의 엔진이 이미 실행 중이면
- **THEN** 기존 marker·process 검사와 journal flock이 두 번째 엔진을 허용하지 않는다
