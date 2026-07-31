## MODIFIED Requirements

### Requirement: 콘솔 안전 불변식

운영 콘솔의 접근 경계는 배포 모드와 일치해야 한다(SHALL): 로컬/토큰 모드는 기존 세션 토큰과 재시작 핸드오프 토큰을 사용하고, 명시적으로 구성된 trusted-network 모드는 인증된 VPN membership을 애플리케이션 접근 경계로 사용하며 별도 애플리케이션 로그인을 요구하지 않는다(SHALL NOT). 상태를 바꾸는 모든 라우트는 공통 `session0` 접근 모드 gate와 CSRF gate를 요구한다(SHALL — trusted-network에서 `session0`는 허용 peer의 VPN membership을 소비한다). 콘솔의 상태변경 행위는 검증 실행 제어(시작·승인·중단), **프로세스 기동·정지**(자기 재시작·soak 재시작·**엔진 시작/정지**), 그리고 **쓰기 전용 Open API 자격증명 설정**뿐이다(SHALL). Open API 설정 라우트는 직접 TLS·허용 peer·정확한 Host/origin·공통 접근 모드·CSRF 뒤에서만 key/secret을 받아야 하고(SHALL), 기존 자격증명을 읽어 응답할 수 없으며(SHALL NOT), 공식 조회 probe로 검증된 key/secret을 보호 저장한 뒤 soak을 시작하는 좁은 capability만 가진다(SHALL). 콘솔 자신은 주문으로 계좌를 건드리지 않는다(SHALL — 자격증명 검증의 공식 계좌 목록 조회는 명시된 read-only 예외이며, 엔진 프로세스가 주문 능력을 갖는지는 §0.7로 승인된 게이트 설정과 기동 인터록이 결정한다). 엔진 상태(실행 여부·기동 거부 사유)는 대시보드에 표시한다(SHALL). 주문 발주·정정·취소·게이트 우회·자격증명 조회 라우트는 존재하지 않는다(SHALL NOT — 라우트 표 정적 검사 + 대표 경로 404 검사).

#### Scenario: 비루프백 리스너
- **WHEN** 루프백이 아닌 주소의 리스너로 Serve가 호출되면
- **THEN** 서비스가 거부되고 리스너가 닫힌다

#### Scenario: 핸드오프 토큰 재사용
- **WHEN** 이미 소비된 핸드오프 토큰으로 재접속하면
- **THEN** 인증이 거부된다

#### Scenario: 주문 라우트 부재
- **WHEN** 콘솔의 라우트 표를 검사하면
- **THEN** 주문·정정·취소·게이트 우회·자격증명 조회 라우트가 없고, Open API 자격증명 설정 POST를 포함한 모든 상태 변경 라우트는 명시적으로 열거되어 공통 접근 모드+CSRF 게이트 뒤에 있다

#### Scenario: trusted-network 자격증명 설정
- **WHEN** 허용 VPN peer가 trusted-network HTTPS 콘솔의 Open API 설정 폼을 제출하면
- **THEN** 별도 애플리케이션 로그인 없이 기존 VPN membership 접근 경계·정확한 Host/origin·CSRF를 통과해야 하며, 어느 경계라도 틀리면 저장 seam 전에 거부된다

#### Scenario: Open API 자격증명 비노출
- **WHEN** 콘솔의 Open API 설정 GET·오류 응답·redirect를 검사하면
- **THEN** 저장되었거나 제출된 key와 secret은 어느 응답 표면에도 존재하지 않는다

#### Scenario: 인터록 미충족 상태의 엔진 시작 버튼
- **WHEN** ProtectionReady 미충족 상태에서 [엔진 시작]을 누르면
- **THEN** 엔진 프로세스가 기동 거부한 사유(미충족 항목)가 대시보드에 표시된다 — 콘솔이 인터록을 우회할 수 없다

#### Scenario: 엔진 정지 버튼
- **WHEN** 실행 중 엔진에 대해 [엔진 정지]를 누르면
- **THEN** 엔진은 시그널 종료 규율(루프 완주·journal 정합 close)로 정지하고 상태가 대시보드에 반영된다
