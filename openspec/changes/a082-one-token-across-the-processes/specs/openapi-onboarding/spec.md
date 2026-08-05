# openapi-onboarding — a082 delta

> **MODIFIED 1건.** 기존 요구사항이 "Short-lived access-token expiry SHALL be
> handled by the existing official client renewal behavior"라고 말하는데, 그
> "existing behavior"가 프로세스 하나를 가정하고 있었다. TossOS는 config
> 디렉터리 하나를 세 상주 프로세스가 공유하므로 그 가정이 참이 아니다. 새
> 요구사항을 더하면 이 문장과 모순되는 SHALL이 둘이 되므로 MODIFIED로 쓴다.
>
> 절 하나를 더하고 scenario 셋을 더한다. 나머지 절과 기존 scenario 6건은 글자
> 그대로 보존한다.

## MODIFIED Requirements

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

토큰 캐시 파일 하나를 여러 프로세스가 공유할 때, 갱신은 그 파일에 대해 수렴해야
한다(SHALL). 인증 거부를 받은 프로세스는 새 토큰을 교환하기 전에 캐시 파일을 다시
읽어야 하며(SHALL), 거기에 자신이 방금 제시한 것과 다른 유효한 토큰이 있으면 그것을
채택하고 교환하지 않아야 한다(SHALL NOT — 교환은 다른 프로세스가 쓰고 있는 토큰을
무효화할 수 있고, 그러면 그 프로세스가 다시 교환해 이쪽을 무효화한다). 메모리에 든
토큰이 아직 만료되지 않았더라도 캐시 파일이 그 뒤에 바뀌었으면 프로세스는 그것을
그대로 신뢰해서는 안 된다(SHALL NOT).

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

#### Scenario: 캐시 파일이 바뀌면 메모리의 토큰을 다시 확인한다
- **WHEN** 메모리에 든 토큰이 아직 만료되지 않았지만 그 토큰을 읽거나 쓴 뒤에 캐시 파일이 바뀌었으면
- **THEN** 그 프로세스는 캐시 파일을 다시 읽어 현재 토큰을 쓰고, 인증 거부를 한 번 받고 나서야 알아차리지 않는다

#### Scenario: 인증 거부 보고가 상태 코드를 남긴다
- **WHEN** 브로커가 401 또는 403으로 요청을 거부하면
- **THEN** 보고된 오류는 그 상태 코드를 담고 응답 본문은 담지 않으며, 그 오류가 인증 거부로 분류되는 것은 달라지지 않는다
