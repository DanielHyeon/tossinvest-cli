# operator-console Specification (delta)

> MODIFIED는 아직 archive되지 않은 선행 change 중 **가장 최근 본문**
> (`enable-vpn-console-access`)을 이어받아 개정한 누적 전문이다 — 이 저장소의 관례대로
> 상태변경 목록에 a075의 한 항목을 더하고 그 항목의 계약을 명시한다. 앞선
> `console-owns-the-operating-toggles`·`console-sets-guardian-limits` 등이 이미
> 그 목록에 넣은 항목은 그대로 유지된다.

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
policy 설정, 검증된 system update 작업, 그리고 **알림 설정**(대상은
`engine.notifications`의 `enabled`·`base_url`·`topic` 셋뿐이다)뿐이다(SHALL).
console에는 direct 주문 발주·정정·취소 또는 credential 표시 route가 없어야 하며
(SHALL NOT), engine/verify가 계좌를 변경할 수 있는 조건은 기존 사람 승인, audit,
startup interlock과 공식 API 경로가 계속 결정한다 (SHALL).

알림 설정을 켜는 것은 사람의 클릭이어야 한다 (SHALL — §0.7의 "사람이 직접 승인한다"는
승인의 주체를 정하는 문장이며, 로컬 콘솔에서 사람이 누르는 것이 곧 그것이다).
그 클릭은 **타이핑 확인이나 2단계 승인 마찰을 요구해서는 안 된다** (SHALL NOT —
사용자 지시 2026-07-27). 화면은 누르기 전에 무엇이 켜지는지, 언제 반영되는지를
읽는 문장으로 말해야 한다 (SHALL).

알림 채널 식별자는 **기계가 만들어야 하며 화면이 사람에게서 받아서는 안 된다**
(SHALL / SHALL NOT — 공개 알림 서비스에서 채널 이름은 유일한 접근 제어이고, 사람이
고른 이름은 추측 가능하다). 생성된 식별자는 암호학적 난수여야 한다 (SHALL).
전송 경로의 인증 토큰을 받는 입력란이 화면에 존재해서는 안 된다 (SHALL NOT —
토큰은 환경에서만 읽는다는 기존 계약을 화면에서도 구조로 유지한다).

채널 식별자는 audit 로그·구조화 로그·리다이렉트 URL의 결과 문구에 기록되어서는
안 되며 (SHALL NOT — 그 자체가 접근 제어이고, 리다이렉트 결과 문구는 브라우저
히스토리와 프록시 로그에 남는다), 화면 본문에는 표시되어야 한다 (SHALL — 표시하지
않으면 운영자가 구독할 수 없다). 알림 설정 변경은 audit 로그에 시각·주체와 함께
기록되어야 하며, 비밀에 해당하는 항목은 값 대신 **설정 여부**로 기록한다 (SHALL).

알림 설정을 끄는 저장은 `enabled` 한 키만 다시 써야 한다 (SHALL — 채널 식별자를
지우면 다시 켤 때 기존 구독이 죽는다; 꺼진 설정에 남은 식별자가 알림을 다시 켜지
않는다는 것은 엔진 조립 경로가 이미 보장한다).

콘솔은 설정한 채널로 **테스트 메시지 한 통**을 보낼 수 있다 (SHALL — 설정은 다음
엔진 기동부터 반영되므로, 그 전에는 전송 가능 여부를 확인할 방법이 없다). 그 발송은
critical 알림 outbox에 기록되어서는 안 되고, 실패해도 entry gate를 latch하거나
operating mode를 변경해서는 안 된다 (SHALL NOT — 테스트는 진단이지 사건이 아니다).
발송 실패는 알림 설정을 되돌려서는 안 되며 (SHALL NOT — 일시적 네트워크 실패로
알림이 꺼진 상태에 머무는 것이 더 나쁜 결과다), 사유가 화면에 표시되어야 한다 (SHALL).

route table 정적 검사는 package 모든 파일의 `HandleFunc`/`Handle`을 검사하고,
모든 state-changing route가 mode와 무관하게 CSRF와 해당 mode의 origin gate chain
뒤에 있음을 보장해야 한다 (SHALL). 상태변경 행위 목록의 확장은 **스펙 문장과 정적
검사 목록이 같은 커밋에서** 함께 움직여야 한다 (SHALL). fixed health 외 별도의
public login/logout lifecycle route는 필요하지 않으며 direct account mutation
capability를 추가해서는 안 된다 (SHALL NOT).

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

#### Scenario: 알림을 버튼 한 번으로 켠다
- **WHEN** 알림이 꺼진 상태에서 운영자가 알림 켜기를 누른다
- **THEN** 기계가 만든 암호학적 난수 채널로 `engine.notifications`의 세 키가 기록되고, audit에 시각·주체와 설정 여부가 남고, 화면이 구독 주소를 본문에 표시하며, 타이핑 확인은 요구되지 않는다

#### Scenario: 채널 식별자는 결과 문구에 실리지 않는다
- **WHEN** 알림 켜기·테스트·끄기의 결과가 리다이렉트로 표시된다
- **THEN** 리다이렉트 URL의 결과 문구에 채널 식별자가 포함되지 않는다

#### Scenario: 테스트 발송 실패
- **WHEN** 설정된 채널로의 테스트 발송이 실패한다
- **THEN** 알림 설정은 켜진 채로 남고, outbox 행이 생기지 않으며, entry gate와 operating mode가 바뀌지 않고, 실패 사유가 화면에 표시된다

#### Scenario: 알림 끄기는 채널을 지우지 않는다
- **WHEN** 알림이 켜진 상태에서 운영자가 알림 끄기를 누른다
- **THEN** 저장된 파일의 `enabled`만 false가 되고 `topic` 바이트는 그대로 남는다

#### Scenario: 토큰 입력란 부재
- **WHEN** 알림 화면의 폼을 검사한다
- **THEN** 전송 경로의 인증 토큰을 받는 입력란이 존재하지 않는다
