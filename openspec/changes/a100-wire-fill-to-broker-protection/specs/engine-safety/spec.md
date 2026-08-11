## ADDED Requirements

### Requirement: supervisor assembly는 배선을 증명하고 실패를 진단 가능하게 거부한다
엔진은 supervisor assembly의 시장별 `Wired` 주장을 세 조건이 모두 참일 때로만 허용해야 한다 (SHALL).
세 조건은 다음과 같다: 그 시장의 공식 conditional-order capability가 검증된 계약과
정확히 일치하고, journal에 커밋된 체결에서 stop/expiry를 유도하는 경로가 조립되어 있으며,
그 시장의 브로커 보호 gateway가 구성되어 있을 것. 세 조건 중 하나라도 아니면 **해당 시장만**
`Wired`가 아니어야 하며 (SHALL), 다른 시장의 판정을 바꾸어서는 안 된다 (MUST NOT).

`Wired` 판단은 assembly 생성자 안에서 리터럴로 결정해서는 안 되고 (MUST NOT), 판정 결과를
주입받아야 한다 (SHALL). assembly 검증 실패는 원인을 구별할 수 있는 typed refusal로 보고해야
하며 (SHALL), 미배선·digest 불일치·중복 시장·잘못된 시장을 하나의 불특정 `Invalid`로 축약해서는
안 된다 (MUST NOT). readiness provider가 아예 구성되지 않은 상태와 구성된 뒤 거부된 상태는
서로 구별 가능해야 한다 (SHALL).

supervisor component digest가 서명된 manifest의 digest와 일치하지 않으면 그 시장은 `UNWIRED`여야
하며 (SHALL), 그 불일치가 조용한 미배선으로 남아서는 안 되고 (MUST NOT) 원인을 지목하는
refusal로 관측 가능해야 한다 (SHALL).

#### Scenario: manifest 재서명 누락
- **WHEN** 재빌드로 component digest가 바뀐 상태에서 이전 manifest가 그대로 pin되어 있다
- **THEN** 두 시장 모두 `UNWIRED`이고 digest 불일치를 지목하는 typed refusal이 보고되며 신규 진입은 0건이다

#### Scenario: 미배선과 미구성의 구별
- **WHEN** manifest pin은 있지만 어떤 시장도 배선을 증명하지 못한다
- **THEN** refusal은 assembly 미배선을 지목하고 `Invalid` 같은 불특정 코드로 보고되지 않는다

#### Scenario: 한 시장만 배선됐다
- **WHEN** KR만 세 조건을 모두 충족하고 US는 gateway가 구성되어 있지 않다
- **THEN** KR assembly만 `Wired`이고 US는 그 사유를 지목하는 refusal이며 두 시장의 기존 보호·청산·대사는 계속된다

#### Scenario: supervisor 입력이 없다
- **WHEN** 배선 판정 입력이 제공되지 않는다
- **THEN** 두 시장 모두 `Wired`가 아니고 관측 가능한 동작은 배선 이전과 동일하다

### Requirement: 보호 배선 봉인은 명시된 하나만 해제된다
엔진 조립은 봉인된 금지 심볼 중 브로커 보호 전송 생성자 하나만 사용할 수 있어야 한다 (SHALL).
보호 controller 생성자, 보호 전용 데이터베이스 경로, 임의 gateway factory 주입은 계속 금지되어야
한다 (MUST NOT). 애플리케이션 코드에서 보호 도메인 패키지를 import할 수 있는 파일은 **단일 조립
파일 하나**로 유지되어야 하며 (SHALL), 두 번째 mutation 경로를 만들어서는 안 된다 (MUST NOT).

해제된 생성자에는 대체 단언이 있어야 한다 (SHALL): 조립된 보호 경로가 journal-backed이고 별도
데이터베이스에 의존하지 않으며 조립 지점이 정확히 하나임을 정적으로 검사해야 한다 (SHALL).
이 검사는 운영 토글, lane, autostart, automation gate 또는 LIVE approval을 생성·변경해서는 안 된다
(MUST NOT).

#### Scenario: 두 번째 조립 지점
- **WHEN** 단일 조립 파일이 아닌 애플리케이션 파일이 보호 도메인 패키지를 import한다
- **THEN** 정적 검사가 실패하고 빌드는 두 번째 mutation 경로를 승인하지 않는다

#### Scenario: 보호 전용 DB 재도입
- **WHEN** 조립이 보호 전용 데이터베이스 경로를 참조한다
- **THEN** 정적 검사가 실패하고 보호 상태는 기존 journal 외의 기동 의존성을 얻지 못한다

#### Scenario: 봉인 해제 뒤에도 진입은 닫혀 있다
- **WHEN** 브로커 보호 전송이 조립됐지만 서명된 attestation과 manifest pin이 설치되지 않았다
- **THEN** readiness는 기본 `UNWIRED` 스냅샷이고 관측 가능한 진입 동작은 봉인 이전과 동일하다
