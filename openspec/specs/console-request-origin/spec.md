# console-request-origin Specification

## Purpose

원격 콘솔 상태 변경 요청의 canonical HTTPS origin 판정, 명시적 헤더 우선순위,
privacy-header 환경의 direct TLS+Host fallback, 그리고 기존 독립 안전 gate 보존 계약.

## Requirements

### Requirement: 콘솔 상태 변경은 canonical request origin 증거를 가져야 한다

원격 콘솔의 상태 변경 POST는 configured canonical HTTPS origin과 일치하는 요청에서만 handler에 도달해야 한다(SHALL).
origin 증거는 다음 순서로 평가해야 한다(SHALL):
명시적 `Origin`, `Origin`이 없을 때 유효한 `Referer`의 scheme·host·port,
두 헤더가 모두 없을 때 직접 TLS 연결과 요청 `Host`로 계산한 HTTPS origin.
명시적 헤더가 있으면 더 낮은 우선순위의 증거로 불일치를 덮어서는 안 된다(SHALL NOT).
현재 우선순위에서 평가하는 명시적 헤더 key가 존재하지만 값이 비었거나 공백뿐이거나
여러 값이거나 파싱할 수 없으면 요청을 거부해야 한다(SHALL).
`Origin` key가 존재하면 `Referer`는 값의 유효성과 무관하게 평가하지 않아야 한다(SHALL NOT).
URL path·query·fragment는 origin 비교에 포함해서는 안 된다(SHALL NOT).

#### Scenario: privacy header 환경의 same-host POST
- **WHEN** `Origin`과 `Referer`가 없고 요청이 TLS로 canonical host와 port에 도착하며 유효한 CSRF를 가진다
- **THEN** origin gate와 CSRF gate를 통과해 대상 handler에 도달한다

#### Scenario: 명시적 cross-origin은 request host로 덮지 않는다
- **WHEN** 직접 TLS와 요청 Host는 canonical이지만 명시적 `Origin` 또는 `Referer`의 scheme·host·port가 다르다
- **THEN** 요청은 handler 전에 거부되고 request-host fallback은 적용되지 않는다

#### Scenario: 명시적 Origin이 Referer보다 우선한다
- **WHEN** `Origin`과 `Referer`가 모두 있고 두 값이 서로 다르다
- **THEN** `Origin`만 최종 증거로 평가하며 `Referer`로 결과를 덮지 않는다

#### Scenario: 유효한 Origin과 잘못된 Referer가 함께 있음
- **WHEN** canonical `Origin`과 비어 있거나 malformed이거나 cross-origin인 `Referer`가 함께 있다
- **THEN** `Referer`를 평가하지 않고 canonical `Origin`에 따라 같은 origin으로 판정한다

#### Scenario: 잘못된 Origin과 유효한 Referer가 함께 있음
- **WHEN** 비어 있거나 malformed이거나 cross-origin인 `Origin`과 canonical `Referer`가 함께 있다
- **THEN** `Referer`로 fall through하지 않고 `Origin`에 따라 거부한다

#### Scenario: 명시적 헤더 key의 값이 유효하지 않음
- **WHEN** 우선순위상 평가되는 `Origin` 또는 `Referer` key가 있지만 값이 없거나 공백뿐이거나 여러 값이거나 파싱할 수 없다
- **THEN** 요청은 handler 전에 거부되고 TLS+Host fallback은 적용되지 않는다

#### Scenario: headerless 비TLS 요청
- **WHEN** `Origin`과 `Referer`가 없고 요청이 TLS 연결이 아니다
- **THEN** 요청은 handler 전에 거부된다

#### Scenario: Referer의 하위 경로
- **WHEN** `Referer`의 scheme·host·port는 canonical이고 path가 `/settings`, `/optimization` 또는 `/restart`다
- **THEN** path와 무관하게 같은 origin으로 판정한다

### Requirement: origin fallback은 기존 독립 안전 gate를 유지해야 한다

상태 변경 POST의 headerless TLS+Host fallback은 기존 독립 안전 gate를 우회해서는 안 된다(SHALL NOT).
독립 안전 gate에는 peer CIDR, exact Host, POST method, process-local CSRF와 기존 action audit이 포함된다.
forwarding header는 origin
증거로 사용해서는 안 된다(SHALL NOT).
호환 로그인 `POST /login`은 TLS+Host fallback을 사용해서는 안 되며(SHALL NOT),
기존의 명시적 Origin/Referer 검사, 로그인 token, rate limit, audit 순서를 유지해야 한다(SHALL).

#### Scenario: canonical TLS Host지만 CSRF가 틀림
- **WHEN** headerless POST의 TLS와 Host는 canonical이지만 CSRF가 없거나 틀리다
- **THEN** 대상 handler에 도달하지 않고 기존 CSRF 오류로 거부된다

#### Scenario: forwarding header만 canonical
- **WHEN** 직접 요청 Host 또는 TLS가 canonical 조건을 충족하지 않고 `X-Forwarded-Host` 또는 `X-Forwarded-Proto`만 configured origin을 가리킨다
- **THEN** 요청은 handler 전에 거부된다

#### Scenario: headerless 호환 로그인
- **WHEN** `POST /login`에 `Origin`과 `Referer`가 없고 login token이 맞더라도
- **THEN** credential parsing, rate-limit mutation, session 발급과 audit 전에 origin 오류로 거부된다

### Requirement: 콘솔 문서는 same-origin 폼의 canonical origin 증거를 보존해야 한다

The console SHALL send `Referrer-Policy: same-origin` on remote console
responses and SHALL declare the same policy in every HTML policy surface used
by normal console pages and the restart interstitial. 이 정책은 같은 origin
내부에서만 referrer 정보를 유지하고 cross-origin 목적지에는 이를 보내지 않아야 한다.
The console SHALL use this consistent document policy so a state-changing form
submitted from the canonical HTTPS console provides canonical origin evidence
instead of an explicit opaque `Origin: null`. The server SHALL NOT accept an
explicit opaque origin to obtain this compatibility. The change SHALL NOT alter
the existing deny-by-default CSP, peer CIDR, exact Host, TLS, CSRF, action audit,
or handler order.

#### Scenario: 응답과 문서 정책의 일치

- **WHEN** remote wrapper, normal page renderer, shared head, 또는 restart interstitial이 콘솔 문서를 제공한다
- **THEN** 모든 활성 `Referrer-Policy` 값은 정확히 `same-origin`이고 `no-referrer`로 덮어쓰지 않는다

#### Scenario: canonical Chrome 폼 POST

- **WHEN** Chrome이 canonical HTTPS 콘솔 문서에서 상태 변경 폼을 제출한다
- **THEN** 요청은 canonical `Origin` 증거로 origin gate를 통과하고 기존 CSRF gate에서 검증된다

#### Scenario: opaque origin은 계속 거부함

- **WHEN** 상태 변경 요청이 explicit `Origin: null`을 보낸다
- **THEN** TLS, Host 또는 CSRF 값과 무관하게 handler 전에 origin 오류로 거부된다

#### Scenario: cross-origin referrer 비공개

- **WHEN** 콘솔 문서에서 cross-origin 목적지로 요청이 만들어진다
- **THEN** 콘솔 URL referrer 정보는 목적지에 전송되지 않는다

#### Scenario: DevTools well-known probe

- **WHEN** Chrome DevTools가 deny-by-default CSP 아래에서 선택적 well-known endpoint 연결을 시도한다
- **THEN** CSP는 완화되지 않으며 해당 브라우저 진단은 콘솔 폼 기능과 독립적으로 차단될 수 있다
