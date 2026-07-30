## ADDED Requirements

### Requirement: 콘솔의 엔진 자동 시작 승인

운영 콘솔은 인증된 운영자가 `engine.autostart`를 ON/OFF로 저장할 수 있어야 한다(SHALL).
이 상태변경 라우트는 세션+CSRF 뒤에 있어야 하고(SHALL), 한 키만
기록하며 변경 전후 값·시각·주체를 audit에 남겨야 한다(SHALL). ON 저장이 성공하면
콘솔은 기존 엔진 시작 seam을 한 번 호출해야 하며(SHALL), 성공·이미 실행·startup
interlock 거부 결과를 화면에 표시해야 한다(SHALL). OFF 저장은 다음 프로세스
기동의 자동 시작만 막고 실행 중 엔진을 정지해서는 안 된다(SHALL NOT).

#### Scenario: 자동 시작 ON 저장
- **WHEN** 인증된 운영자가 유효한 CSRF와 함께 엔진 자동 시작을 ON으로 저장하면
- **THEN** `engine.autostart` 한 키가 true로 기록되고 audit가 추가되며 기존 엔진 시작 seam이 정확히 한 번 호출된다

#### Scenario: 기동 인터록 거부
- **WHEN** 자동 시작 ON 저장 뒤 기존 엔진 시작 seam이 startup interlock 오류를 반환하면
- **THEN** 설정은 ON으로 남고 엔진 자신의 거부 사유가 화면에 표시되며 인터록을 우회하는 두 번째 시작 경로는 실행되지 않는다

#### Scenario: 자동 시작 OFF 저장
- **WHEN** 실행 중 엔진이 있는 상태에서 자동 시작을 OFF로 저장하면
- **THEN** `engine.autostart`는 false가 되고 자동 정지 호출은 발생하지 않으며 화면은 [엔진 정지]를 별도 사용하라고 알린다

#### Scenario: CSRF 없는 자동 시작 변경
- **WHEN** 세션은 유효하지만 CSRF가 없거나 틀린 자동 시작 저장 요청이 도달하면
- **THEN** 요청이 거부되고 config·audit·엔진 프로세스 상태가 모두 변경되지 않는다

#### Scenario: 설정 키 격리
- **WHEN** 콘솔에서 엔진 자동 시작을 저장하면
- **THEN** `engine.autostart` 이외의 automation gate·Guardian 한도·거래 정책 바이트는 그대로다
