## MODIFIED Requirements

### Requirement: 콘솔 안전 불변식

운영 콘솔은 두 mode 중 하나로만 서비스해야 한다 (SHALL): ① 기본 native local mode는
127.0.0.1 listener와 기존 terminal possession session을 사용하고, ② 명시적
trusted-network mode는
`vpn-console-access` spec의 host loopback/VPN publish·CIDR·TLS·canonical Host/Origin
계약을 모두 사용하며 application login을 요구하지 않는다. remote mode가
완전히 구성되지 않은 non-loopback listener는 거부하고 닫아야 한다 (SHALL).

재시작 handoff는 프로세스 교체 무결성을 위한 내부 단발성 자격으로만 유지하며,
사용자 application login으로 사용해서는 안 된다 (SHALL NOT).

상태를 바꾸는 모든 route는 CSRF gate를 요구하고 trusted-network에서는
same-origin 검사도 요구한다 (SHALL). 현재 허용된 상태 변경은 검증 제어, 자기/soak/
engine process 제어, 편입·Guardian 한도·trading policy·automation gate·공통 exit
policy 설정, 검증된 system update 작업뿐이다(SHALL). 이 change는 그 목록이나 각
seam의 기존 쓰기 범위를 넓히지 않는다 (SHALL NOT). console에는 direct 주문 발주·
정정·취소 또는 credential 표시 route가 없어야 하며(SHALL NOT), engine/verify가
계좌를 변경할 수 있는 조건은 기존 사람 승인, audit, startup interlock과 공식 API
경로가 계속 결정한다 (SHALL).

route table 정적 검사는 package 모든 파일의 `HandleFunc`/`Handle`을 검사하고,
모든 state-changing route가 mode와 무관하게 CSRF와 해당 mode의 origin gate chain
뒤에 있음을 보장해야 한다 (SHALL). fixed health 외 별도의 public login/logout
lifecycle route는 필요하지 않으며 direct account mutation capability를 추가해서는
안 된다 (SHALL NOT).

#### Scenario: 기본 local listener
- **WHEN** remote option 없이 console을 시작한다
- **THEN** 127.0.0.1만 bind하고 기존 terminal session URL로 인증한다

#### Scenario: 완전한 remote listener
- **WHEN** 유효한 remote access 구성을 가진 console이 non-loopback listener로 Serve된다
- **THEN** TLS/network/origin gate 아래에서 application login 없이 local console과 같은 handler 기능을 제공한다

#### Scenario: 불완전한 비루프백 listener
- **WHEN** remote access 구성이 없거나 불완전한 console이 non-loopback listener를 받는다
- **THEN** service가 거부되고 listener가 닫힌다

#### Scenario: 사용자 credential 부재
- **WHEN** loopback 또는 trusted VPN browser가 console을 연다
- **THEN** token 파일 조회, login form 또는 login cookie 교환 없이 화면을 제공한다

#### Scenario: remote 상태 변경 gate
- **WHEN** CSRF 또는 same-origin 증거가 틀린 설정/engine/verify POST가 도착한다
- **THEN** 요청은 handler seam 전에 거부되고 config, process, account는 바뀌지 않는다

#### Scenario: 주문 route 부재
- **WHEN** console route table과 capability closure를 검사한다
- **THEN** remote login 기능이 direct order/broker credential capability를 추가하지 않았고 기존 direct 주문 route 금지가 유지된다

#### Scenario: engine interlock 유지
- **WHEN** remote session에서 ProtectionReady 미충족 engine start를 누른다
- **THEN** engine process의 기존 interlock refusal이 그대로 표시되고 console이 우회하지 않는다
