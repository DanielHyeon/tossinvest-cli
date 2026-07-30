## ADDED Requirements

### Requirement: 원격 listener는 완전한 VPN 보안 구성이 있어야 열린다

원격 콘솔은 완전한 원격 보안 구성이 있을 때만 non-loopback listener를 열어야 한다 (SHALL).
명시적 bind IP, 하나 이상의 허용 VPN CIDR, canonical HTTPS public URL,
public host를 검증하는 TLS certificate/key와 명시적 접근 모드가 모두 유효할 때만
non-loopback listener를 열어야 한다 (SHALL). trusted-network 접근은 운영자가
loopback 또는 인증된 VPN membership을 application 접근 경계로 승인한 경우에만
명시적으로 선택해야 한다 (SHALL). 부분 구성, HTTP URL, 전역 CIDR 또는 암묵적으로
선택된 접근 모드는 socket을 열기 전에 fail-closed로 거부해야 한다 (SHALL).

#### Scenario: 접근 모드 없이 bind만 지정
- **WHEN** 운영자가 non-loopback bind만 지정하고 접근 모드/TLS/public URL/CIDR 중 하나를 생략한다
- **THEN** 콘솔은 기동을 거부하고 어떤 non-loopback socket도 열지 않는다

#### Scenario: 기존 기본 실행
- **WHEN** 원격 관련 옵션을 하나도 지정하지 않고 콘솔을 기동한다
- **THEN** 기존처럼 127.0.0.1 HTTP listener와 터미널 session URL만 사용한다

#### Scenario: VPN 밖 peer
- **WHEN** 원격 listener에 허용 CIDR 밖 실제 peer IP가 접속한다
- **THEN** `X-Forwarded-For` 값과 무관하게 어떤 콘솔 handler 전 단계에서 거부된다

### Requirement: trusted-network 접근은 application credential을 요구하지 않는다

trusted-network 모드는 access token, login form, login session 또는 application identity provider를 요구해서는 안 된다 (SHALL NOT).
허용 peer가 canonical HTTPS
Host로 요청하면 network·Host 보안 검사를 통과한 뒤 dashboard를 직접 제공해야 한다
(SHALL). application login을 제거하더라도 TLS, allowed CIDR, exact Host,
Origin/Referer, CSRF와 기존 action audit을 우회해서는 안 된다 (SHALL NOT).

#### Scenario: 정상 모바일 접근
- **WHEN** 허용 VPN peer가 canonical HTTPS Host로 dashboard를 GET한다
- **THEN** token 입력, login redirect 또는 session cookie 없이 dashboard가 반환된다

#### Scenario: loopback host-publish 접근
- **WHEN** Compose가 host의 127.0.0.1에만 publish된 trusted-network console을 제공한다
- **THEN** 같은 호스트 브라우저는 token 파일을 읽지 않고 dashboard를 직접 연다

#### Scenario: 상태변경 안전 유지
- **WHEN** application login 없이 설정 또는 engine POST를 보낸다
- **THEN** same-origin과 CSRF가 모두 맞아야 기존 handler에 도달하고 기존 action audit 및 interlock이 그대로 적용된다

### Requirement: trusted-network 선택은 명시적이고 credential-free다

trusted-network 선택은 command/config에 명시되어야 하며(SHALL), token value나
token file을 함께 요구하거나 image, Compose secret, environment, banner에
credential을 생성해서는 안 된다 (SHALL NOT). 인증형 compatibility mode가 남아 있는
경우 trusted-network와 동시에 선택할 수 없어야 하고(SHALL NOT), 어떤 모드인지
기동 banner와 비밀값 없는 audit/운영 기록으로 확인할 수 있어야 한다 (SHALL).

#### Scenario: 접근 모드 충돌
- **WHEN** trusted-network와 token-auth mode가 동시에 설정된다
- **THEN** console은 socket을 열기 전에 모호한 접근 구성을 거부한다

#### Scenario: token secret 제거
- **WHEN** trusted-network Compose 구성을 렌더링한다
- **THEN** remote token file, token secret mount와 login credential 환경값이 존재하지 않는다

### Requirement: 원격 요청은 canonical same-origin이다

원격 mode의 health endpoint 외 모든 요청은 exact canonical Host를 가져야 한다 (SHALL).
(SHALL). 상태 변경 POST는 기존 CSRF와 함께 exact HTTPS Origin을 가져야 하며,
Origin이 없는 경우에만 same-origin Referer를 허용해야 한다 (SHALL). 다른 Host,
Origin/Referer 없음, proxy forwarding header만 일치하는 요청은 거부해야 한다
(SHALL).

#### Scenario: DNS rebinding Host
- **WHEN** 허용 VPN peer가 유효 session을 보내지만 Host가 public URL과 다르다
- **THEN** handler나 config/process seam에 도달하기 전에 403으로 거부된다

#### Scenario: cross-origin 상태 변경
- **WHEN** 유효 session과 CSRF가 있지만 Origin이 canonical origin과 다르다
- **THEN** 상태 변경은 403이며 아무 seam도 호출되지 않는다

### Requirement: 원격 network boundary와 응답 보안은 유지된다

실제 peer CIDR과 canonical Host 검사는 forwarding header보다 먼저 수행해야 한다 (SHALL).
forwarding header로 접근 경계를 바꿀 수 없어야 한다 (SHALL NOT).
응답은 CSP, clickjacking, MIME, referrer, no-store 헤더를 가져야 하고 원격 mode는
HSTS를 가져야 한다 (SHALL). 기존 상태변경 action audit은 application login 유무와
관계없이 유지되어야 한다 (SHALL).

#### Scenario: forwarding header 우회
- **WHEN** 허용 CIDR 밖 peer가 허용 주소의 `X-Forwarded-For`를 보낸다
- **THEN** request는 handler 전에 거부된다

#### Scenario: 보안 헤더
- **WHEN** trusted-network 화면 응답을 검사한다
- **THEN** HSTS, CSP frame/form 제한, DENY frame, nosniff, no-referrer, no-store가 존재한다

### Requirement: health endpoint는 고정된 최소 응답이다

`GET|HEAD /healthz`는 session 없이 TLS server liveness를 확인할 수 있어야 한다 (SHALL).
(SHALL), 고정 status/body 외 account/config/session 정보를 읽거나 표시해서는 안
된다(SHALL NOT). 다른 method는 405여야 한다 (SHALL).

#### Scenario: container healthcheck
- **WHEN** local container peer가 HTTPS `/healthz`를 GET한다
- **THEN** broker/config/journal 호출 없이 200과 고정 `ok`만 반환한다
