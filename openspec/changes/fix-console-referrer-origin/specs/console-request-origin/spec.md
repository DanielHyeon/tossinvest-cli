## ADDED Requirements

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
