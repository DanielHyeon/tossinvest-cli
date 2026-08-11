## ADDED Requirements

### Requirement: 보호 배선 봉인은 명시된 하나만 해제되고 경계는 넓어진다
엔진 조립은 봉인된 금지 심볼 중 브로커 보호 전송 생성자 하나만 사용할 수 있어야 한다 (SHALL).
보호 controller 생성자, 보호 전용 데이터베이스 경로, 임의 gateway factory 주입은 계속 금지되어야
한다 (MUST NOT).

애플리케이션 코드가 보호 도메인 패키지를 import할 수 있는 파일은 **명시된 조립 파일로 제한되어야
하며** (SHALL), 그 정적 검사는 broker-neutral 도메인 패키지뿐 아니라 **브로커 전송 패키지와
보호 lifecycle 패키지, 그리고 보호 수렴 컴포넌트까지 대상으로 삼아야 한다** (SHALL).
검사 대상이 도메인 패키지 하나로 남아 두 번째 mutation 경로가 다른 패키지를 통해 열려서는
안 된다 (MUST NOT).

해제된 생성자에는 대체 단언이 있어야 한다 (SHALL): 조립된 보호 경로가 journal-backed이고 별도
데이터베이스에 의존하지 않으며 조립 지점이 정확히 하나임을 정적으로 검사해야 한다 (SHALL).
이 검사는 운영 토글, lane, autostart, automation gate 또는 LIVE approval을 생성·변경해서는 안 된다
(MUST NOT).

#### Scenario: 브로커 전송 패키지의 두 번째 조립 지점
- **WHEN** 명시된 조립 파일이 아닌 애플리케이션 파일이 브로커 보호 전송 패키지를 import한다
- **THEN** 정적 검사가 실패하고 빌드는 두 번째 mutation 경로를 승인하지 않는다

#### Scenario: 보호 전용 DB 재도입
- **WHEN** 조립이 보호 전용 데이터베이스 경로를 참조한다
- **THEN** 정적 검사가 실패하고 보호 상태는 기존 journal 외의 기동 의존성을 얻지 못한다

#### Scenario: 봉인 해제 뒤에도 진입은 닫혀 있다
- **WHEN** 브로커 보호 전송이 조립되어 상주 보호주문이 등록된다
- **THEN** supervisor assembly는 여전히 어떤 시장에서도 `Wired`를 주장하지 않고 readiness는 기본 `UNWIRED` 스냅샷이며 신규 진입은 0건이다
