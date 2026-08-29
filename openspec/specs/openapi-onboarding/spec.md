# openapi-onboarding Specification

## Purpose
TBD - created by archiving change automate-soak-openapi-onboarding. Update Purpose after archive.
## Requirements
### Requirement: one-action soak onboarding

In configured HTTPS mode, the console SHALL make an explicit soak restart one
continuous operator action.
It SHALL validate persisted official Open API credentials with one read-only
probe before spawning. Ready credentials SHALL proceed directly to soak
restart. Missing or authentication-rejected credentials SHALL route to the
HTTPS Open API setup screen without spawning. A successful setup submission
SHALL validate and persist the submitted credentials and then start the soak
without requiring a second restart click. The ordinary restart-preflight path
and setup-save-continuation path SHALL be mutually exclusive from the first
credential check through the final restart result.

#### Scenario: saved credentials are ready
- **WHEN** the operator clicks soak restart and the persisted credentials pass the official read-only probe
- **THEN** the console invokes soak restart immediately and returns its truthful result without another operator action

#### Scenario: credentials are missing
- **WHEN** the operator clicks soak restart and no persisted or environment credentials are available
- **THEN** no soak process is spawned and the operator is redirected to an HTTPS setup page that explains the required key and secret

#### Scenario: credentials are rejected
- **WHEN** the operator clicks soak restart and the official probe classifies the persisted credentials as authentication-rejected
- **THEN** no soak process is spawned and the operator is redirected to the setup page with replacement guidance

#### Scenario: rejected credentials are environment-managed
- **WHEN** restart preflight rejects credentials supplied by the container environment
- **THEN** no soak process is spawned, no credential file replacement is offered, and the operator receives fixed guidance to update or recreate the container environment

#### Scenario: transient validation failure
- **WHEN** the explicit restart preflight encounters an IP allow-list, rate-limit, server, or transport failure
- **THEN** no soak process is spawned, the failure is classified for the operator, and the key is not misreported as missing

#### Scenario: successful first-time setup
- **WHEN** the operator submits a key and secret that pass the official read-only probe and protected persistence succeeds
- **THEN** the console invalidates the normal token cache, records a secret-free save-success audit event, waits for the old soak to exit, invalidates the normal cache again immediately before spawn, starts the soak in the same POST flow, and redirects to the dashboard only with the successful restart result

#### Scenario: old soak recreates a token during replacement
- **WHEN** an old soak can refresh and recreate the shared normal token cache after credential preflight or save has invalidated it
- **THEN** restart waits for that old process to exit and applies a second token-cache invalidation fence immediately before detached spawn; fence failure starts no new soak and returns a truthful retryable restart error

#### Scenario: concurrent setup and restart requests
- **WHEN** setup and restart requests overlap
- **THEN** one completes its credential check/save/restart sequence before the other begins, so a request cannot validate one credential generation and restart with another

#### Scenario: persistence or audit fails
- **WHEN** submitted credentials validate but protected persistence or the required save-success audit fails
- **THEN** the console reports the classified failure, does not start soak, does not claim success, and does not record save success before persistence; once persistence is attempted it retains a secret-free 0600 pending-generation marker across restarts even if the store returns an error, blocks ordinary restart preflight, reopens the blank setup screen, and clears the marker only after a later valid submission completes persistence, token invalidation, audit, and marker removal

#### Scenario: pending marker cannot be read or removed
- **WHEN** restart preflight cannot read the file-generation marker or completed setup cannot remove it
- **THEN** the console remains fail-closed, starts no soak, and reports fixed retry guidance without exposing credential-derived content

### Requirement: persistent credential and token lifecycle

The console SHALL persist a validated key and secret through the existing
protected Open API credential store and SHALL reuse that store across console
and container restarts. Short-lived access-token expiry SHALL be handled by the
existing official client renewal behavior and SHALL NOT require the operator to
re-enter a stored valid key and secret. The setup path SHALL be offered when
credentials are missing or authentication-rejected, not merely because the
cached access token expired. An incomplete-generation marker SHALL use the same
persistent config mount as the credential file at
`<config-dir>/openapi-onboarding.pending`.

토큰 캐시 파일 하나를 여러 보유자가 공유할 때, 갱신은 그 파일에 대해 수렴해야
한다(SHALL). 인증 거부를 받은 보유자는 새 토큰을 교환하기 전에 캐시 파일을 다시
읽어야 하며(SHALL), 거기에 **자신이 거부당한 바로 그 토큰과 다른** 유효한 토큰이
있으면 그것을 채택하고 교환하지 않아야 한다(SHALL NOT — 교환은 다른 보유자가 쓰고
있는 토큰을 무효화할 수 있고, 그러면 그 보유자가 다시 교환해 이쪽을 무효화한다).
거부당한 토큰이 무엇인지는 호출자가 알려야 하며 캐시 상태에서 추론해서는 안
된다(SHALL NOT — 같은 프로세스의 다른 실행 흐름이 그 사이에 캐시를 바꿀 수 있고,
그 창에서의 추론은 같은 결함을 프로세스 안에서 되풀이한다).

**채택한 토큰은 검증된 토큰이 아니다.** 그 생존은 만료 시각에서 추론한 것이므로,
채택한 토큰이 다시 거부당하면 요청은 새로 발급한 토큰으로 한 번 더 시도해야
한다(SHALL). 재시도 예산을 추측에 쓰고 그대로 인증 실패를 올려서는 안 된다
(SHALL NOT — 그 실패는 진입 게이트를 잠그고 그 잠금은 재시작으로 풀리지 않으며,
그것을 올린 판정 주기는 손절 판정을 하지 않는다). 갱신 시도 횟수에는 상한이
있어야 한다(SHALL).

캐시 파일 쓰기는 원자적이어야 한다(SHALL — 자리에서 잘라 쓰면 그 사이에 읽는
보유자가 토큰이 없다고 판단해 하나를 사고, 방금 쓴 보유자의 토큰을 무효화한다.
채택하는 읽기는 다른 보유자가 방금 썼을 때 정확히 일어난다).

토큰 획득 경로는 손절·비상 청산을 포함한 모든 브로커 읽기가 지나가므로, 이 수렴을
위해 프로세스 간 블로킹 대기를 도입해서는 안 된다(SHALL NOT — 한 프로세스의 지연이
다른 프로세스의 판정 간격에 더해지는 결합을 만든다).

인증 거부를 보고할 때 상태 코드를 함께 남겨야 한다(SHALL — 401과 403은 다른 원인을
가리키고, 이 둘을 구분하지 못하면 토큰 경합과 권한·IP 문제가 같은 한 줄로 보인다).
응답 본문은 남기지 않는다(SHALL NOT — 계좌 식별자나 자격증명 파생값이 들어올 수
있다). 이 보고는 기존 분류 결과를 바꾸지 않는다(SHALL — 어떤 오류가 인증 거부로
분류되는지는 그대로다).

#### Scenario: environment pair while a file generation is pending
- **WHEN** both environment credential variables are non-empty while a file-managed pending marker exists
- **THEN** the environment pair remains authoritative exactly as the existing loader defines, file onboarding is not offered, and the dormant marker is neither rewritten nor cleared nor allowed to affect the environment generation; if the environment pair is later removed, the marker blocks file-mode restart until setup completes

#### Scenario: environment credentials replace a cached generation
- **WHEN** a complete environment credential pair is checked while the normal token cache may contain a valid token issued for an older pair
- **THEN** preflight validates the environment pair through an isolated temporary 0600 token cache, removes that temporary cache, invalidates the normal cache, and requires the post-exit pre-spawn invalidation fence so the child cannot reuse the old generation

#### Scenario: replacement target was permissive
- **WHEN** protected persistence replaces an existing credential path whose mode is more permissive than 0600
- **THEN** the store atomically replaces it with a verified regular 0600 file or reports failure while the pending marker remains and no soak starts

#### Scenario: container replacement
- **WHEN** a validated credential file was saved through the console and the Docker container is recreated with the existing config bind mount
- **THEN** the next soak restart reuses the saved credentials without showing the setup form

#### Scenario: cached access token expires
- **WHEN** the cached official access token is expired but the persisted key and secret remain valid
- **THEN** the official client renews the token during the restart preflight and the soak starts without credential re-entry

#### Scenario: 다른 프로세스가 먼저 갱신한 토큰을 채택한다
- **WHEN** 한 프로세스가 인증 거부를 받았고 공유 캐시 파일에 자신이 제시한 것과 다른 유효한 토큰이 이미 들어 있으면
- **THEN** 그 프로세스는 교환하지 않고 파일의 토큰을 채택해 재시도하며, 다른 프로세스가 쓰고 있는 토큰은 계속 유효하다

#### Scenario: 채택할 토큰이 없으면 교환한다
- **WHEN** 인증 거부를 받았고 공유 캐시 파일에 방금 제시한 것과 같은 토큰만 있거나 유효한 토큰이 없으면
- **THEN** 그 프로세스는 교환하고 결과를 캐시 파일에 남긴다

#### Scenario: 채택한 토큰도 거부당하면 새로 발급해 다시 시도한다
- **WHEN** 거부당한 보유자가 캐시 파일의 토큰을 채택했는데 브로커가 그 토큰도 거부하면
- **THEN** 요청은 새로 발급한 토큰으로 한 번 더 시도해 완료되고, 인증 실패가 호출자에게 올라가지 않는다

#### Scenario: 쓰는 중인 캐시 파일을 읽어도 토큰이 없다고 판단하지 않는다
- **WHEN** 한 보유자가 캐시 파일을 갱신하는 동안 다른 보유자가 그 파일을 읽으면
- **THEN** 읽는 쪽은 이전 토큰이나 새 토큰 중 하나를 온전히 보고, 빈 파일이나 잘린 내용을 보지 않는다

#### Scenario: 인증 거부 보고가 상태 코드를 남긴다
- **WHEN** 브로커가 401 또는 403으로 요청을 거부하면
- **THEN** 보고된 오류는 그 상태 코드를 담고 응답 본문은 담지 않으며, 그 오류가 인증 거부로 분류되는 것은 달라지지 않는다

### Requirement: credential ingress is secret-safe

The Open API setup GET SHALL return blank credential fields and SHALL NOT expose
stored credential values. Credential submission SHALL be accepted only through
a bounded POST body over direct TLS and behind the configured HTTPS peer, exact Host/origin,
configured access mode, method, CSRF, and audit boundaries. Token-authenticated
mode SHALL require its application session. Trusted-network mode SHALL continue
to use authenticated VPN membership as the application access boundary and
SHALL NOT add a separate application login. The console SHALL NOT place the key
or secret in HTML responses, redirect URLs, error text, application logs, audit
details, retained memory, or test output. Failed validation or persistence SHALL
not start a soak and SHALL not persist a rejected pair.

#### Scenario: setup form render
- **WHEN** an authenticated operator opens the Open API setup page
- **THEN** the page contains blank password-type key and secret inputs and contains no stored credential material

#### Scenario: plaintext loopback request
- **WHEN** an otherwise authenticated request opens or submits the Open API setup route without direct TLS
- **THEN** the console returns 403 before parsing or exposing any credential field and performs no validation, persistence, audit, or soak restart

#### Scenario: missing CSRF
- **WHEN** a credential submission has valid field values but lacks the current CSRF token
- **THEN** the request is rejected before validation, persistence, audit, or soak restart

#### Scenario: cross-origin submission
- **WHEN** a credential submission originates outside the configured exact HTTPS origin
- **THEN** the request is rejected before credential parsing or any side effect

#### Scenario: oversized submission
- **WHEN** a credential submission exceeds the setup request-body limit
- **THEN** the route-specific body limit deterministically returns 413 before mutation middleware parses the form and without bypassing peer, Host/origin, configured access mode, method, CSRF, or authorization checks and without validation, persistence, audit, or soak restart

#### Scenario: deployed trusted-network submission
- **WHEN** an allowed VPN peer submits from the exact configured HTTPS origin in trusted-network mode with a valid CSRF token
- **THEN** the request reaches credential validation without an application login, while a wrong peer, Host/origin, or CSRF value remains rejected before the save seam

#### Scenario: invalid credentials
- **WHEN** submitted credentials are rejected by the official read-only probe
- **THEN** neither field is echoed or persisted, no soak starts, and the page shows fixed replacement guidance

#### Scenario: replacement credentials with an existing valid token
- **WHEN** the operator submits a replacement key and secret while the normal token cache still contains a valid token for the old pair
- **THEN** validation uses an isolated temporary 0600 token-cache path, the old token cannot authenticate the submitted pair, and the temporary cache is removed

#### Scenario: saved-token refresh fails
- **WHEN** access-token refresh fails with an authentication or transient official-client classification
- **THEN** soak remains stopped and only fixed classified guidance, with no credential-derived content, is returned

