## ADDED Requirements

### Requirement: M-B1 prescribed rebuild 실패는 원인별로 보고된다

M-B1 collector의 prescribed rebuild 검증은 실패를 원인별로 구분해 보고해야 한다 (SHALL): rebuild 자체의 실패(120-second collector deadline 초과 포함)는 원래 오류를 감싼 "prescribed rebuild failed" 로, 빌드 명령 불일치는 "unexpected command" 로, 실행 파일 digest 불일치는 "running executable does not match the prescribed rebuild" 로 보고한다. 세 경우 모두 collector는 지금과 같이 HOLD 해야 하며 (SHALL), 이 구분은 120-second deadline 이나 HOLD 조건을 바꾸지 않는다 (SHALL NOT).

#### Scenario: 부하로 rebuild 가 시한을 넘김
- **WHEN** prescribed rebuild 가 collector deadline 을 넘겨 실패한다
- **THEN** collector 는 HOLD 하고, 오류는 deadline 초과를 감싼 "prescribed rebuild failed" 로 보고되며 digest 불일치 문구를 쓰지 않는다

#### Scenario: 실행 파일이 rebuild 와 다름
- **WHEN** rebuild 는 성공했지만 digest 가 실행 중인 파일과 다르다
- **THEN** collector 는 HOLD 하고 "running executable does not match the prescribed rebuild" 로 보고한다
