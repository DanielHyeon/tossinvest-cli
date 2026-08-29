# container-deployment Specification

## Purpose
TBD - created by archiving change enable-vpn-console-access. Update Purpose after archive.
## Requirements
### Requirement: image는 재현 가능한 non-root multi-stage build다

TossOS container image는 pinned-version multi-stage build여야 한다 (SHALL). Go builder와 최소 runtime의 multi-stage
build여야 하며(SHALL), runtime에 source/build tool/secret을 포함해서는 안 된다
(SHALL NOT). runtime process는 non-root이고 CA certificate와 healthcheck client만
가져야 한다 (SHALL).

#### Scenario: image metadata와 사용자
- **WHEN** build한 image를 inspect한다
- **THEN** non-root USER, console port, healthcheck가 선언되고 source·VCS·secret file은 image layer에 없다

### Requirement: Compose는 VPN host IP에만 publish한다

Compose의 host publish 주소는 required `TOSSOS_VPN_BIND_IP`여야 하며(SHALL),
unset/blank일 때 config가 실패해야 한다 (SHALL). `0.0.0.0`, 빈 host 또는 public
fallback을 제공해서는 안 된다(SHALL NOT). application command는 container 내부
wildcard bind와 required VPN CIDR/public URL/TLS/token 구성을 함께 전달해야 한다
(SHALL).

#### Scenario: VPN bind IP 누락
- **WHEN** `TOSSOS_VPN_BIND_IP` 없이 `docker compose config`를 실행한다
- **THEN** port wildcard로 대체되지 않고 명시적 오류로 실패한다

#### Scenario: 유효한 VPN 구성
- **WHEN** 모든 required path/IP/CIDR/URL을 제공해 Compose config를 렌더한다
- **THEN** host port는 제공한 VPN IP 하나에만 publish되고 container command에는 완전한 remote-mode 인자가 있다

### Requirement: secret과 state는 image 밖에서 최소 권한으로 제공된다

access token, TLS key/cert, broker session은 image 밖 host file에서 read-only secret으로 mount해야 한다 (SHALL).
mount해야 하며(SHALL), Compose environment나 image build arg에 secret value를
넣어서는 안 된다(SHALL NOT). config/data는 명시적 host directory에 persist하고
그 밖의 host path, Docker socket, device를 mount해서는 안 된다(SHALL NOT).

#### Scenario: Compose secret 검사
- **WHEN** Compose와 env example을 정적으로 검사한다
- **THEN** secret 값은 없고 required file path 변수와 read-only secret target만 존재한다

### Requirement: container runtime은 최소 권한과 bounded resource를 가진다

Compose service는 최소 권한과 bounded resource로 실행해야 한다 (SHALL). read-only root filesystem, tmpfs `/tmp`, `cap_drop: ALL`,
`no-new-privileges`, init, PID/CPU/memory/log limit, healthcheck, restart policy를
가져야 한다 (SHALL). container update는 image 교체로 수행하고 in-container
self-update를 운영 절차로 사용해서는 안 된다(SHALL NOT).

#### Scenario: hardened Compose inspect
- **WHEN** 유효한 Compose config를 검사한다
- **THEN** 모든 runtime hardening과 bounded log/resource 설정이 존재하고 privileged/host network/Docker socket이 없다

#### Scenario: health smoke test
- **WHEN** dummy credential과 self-signed test certificate로 container를 시작한다
- **THEN** live account 요청 없이 healthcheck가 healthy가 되고 종료 후 persistent host state 외 잔여 secret이 image에 남지 않는다
