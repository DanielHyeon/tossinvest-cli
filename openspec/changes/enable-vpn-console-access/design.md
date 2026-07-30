## Context

현재 `internal/console.Listen`은 `127.0.0.1`을 코드에 고정하고 `Serve`가 실제
listener를 다시 검사한다. 초기 설계는 local URL token과 remote application token을
각각 cookie/session으로 교환했다. 사용자는 2026-07-31에 단일 운영자 환경의 host
loopback 또는 인증된 VPN membership을 application 접근 신뢰 경계로 사용하고,
별도 token/login/session을 제거하도록 명시 승인했다. 상태 변경 route의 CSRF,
원격 Host/Origin, action audit와 엔진/검증 interlock은 그대로 유지한다.

사용자는 VPN에 연결된 모바일 브라우저에서 조회뿐 아니라 현재 콘솔의 설정,
엔진 제어, 검증 작업도 수행하려 한다. 호스트에는 Docker/Compose가 있고 OpenVPN
서비스가 설치돼 있지만 현재 활성 tunnel 주소는 없으므로 특정 VPN 제품·interface
이름·주소를 저장소가 추측할 수 없다.

## Goals / Non-Goals

**Goals:**

- 아무 플래그가 없는 native 실행은 기존 loopback HTTP + 터미널 possession token을
  보존한다.
- 원격 trusted-network 모드는 VPN CIDR, 명시적 bind, canonical HTTPS origin,
  인증서/키와 명시적 trust 선택이 모두 있어야만 시작한다.
- local/VPN browser는 token/login/session 없이 같은 콘솔 기능을 사용한다.
- 상태변경은 기존 CSRF, 원격 same-origin, action audit, engine/verify interlock을
  그대로 사용하며 별도 권한 확장을 만들지 않는다.
- Docker Compose는 host의 VPN 주소에만 port를 publish하고 image/compose에 secret을
  넣지 않는다.

**Non-Goals:**

- VPN 자체의 설치·사용자 등록·접근정책 관리.
- 인터넷 공개, reverse proxy 신뢰, `X-Forwarded-*` 처리, 다중 사용자/RBAC.
- 원격 route에서 직접 주문을 발주·정정·취소하는 새 기능.
- 기존 automation gate, trading policy, verify nonce, engine interlock의 승인 의미 변경.
- container 내부 self-update. container는 새 image로 교체한다.

## Decisions

### D1. 두 개의 명시적 listener mode

`Options.Remote`가 비어 있으면 기존 `Listen(port)`, `loopbackOnly`와 local session
credential을 그대로 사용한다. 원격 모드는 `Bind`,
`AllowedCIDRs`, `PublicURL`, `TLSCertFile`, `TLSKeyFile`, `AccessAudit`와
`TrustedNetwork=true`가 모두 유효할 때만 credential-free 접근으로 활성화한다.
기존 token-auth mode는 rollback/호환을 위해 선택적으로 유지하되 두 접근 모드를
동시에 구성할 수 없다.
부분 구성, HTTP public URL, path/query/userinfo가 있는 URL, IP literal이 아닌 bind,
빈/전역 CIDR, public URL host를 검증하지 못하는 인증서는 기동 전에 거부한다.

원격 listener는 bind IP와 실제 `net.TCPAddr`가 일치해야 한다. `0.0.0.0`/`::`는
container 내부 bind를 위해 허용하지만, 모든 요청의 실제 peer IP를
`AllowedCIDRs`에 다시 대조한다. `X-Forwarded-For`는 신뢰하지 않는다. Compose는
별도로 host publish를 `TOSSOS_VPN_BIND_IP`에 고정하므로 wildcard는 container
network 안에만 남는다.

대안으로 loopback server 앞에 Caddy/Nginx를 두는 방식을 검토했다. proxy
identity/Host 전달 계약과 새 운영 의존성이 생기므로 단일 Go server의 계약을
검증하는 쪽을 선택한다.

### D2. loopback/VPN membership을 application 접근 경계로 사용

사용자의 명시적 운영 결정에 따라 host loopback publish 또는 인증된 VPN membership을
application 접근 경계로 사용한다. trusted-network mode에는 token file, login form,
session cookie와 logout lifecycle을 만들지 않는다. Compose는 이 선택을 명시 플래그로
전달하며, 누락 시 원격 listener를 열지 않는다.

이 결정은 network 접근자가 콘솔의 전체 단일-운영자 권한을 가진다는 뜻이다. 따라서
VPN 계정/기기 관리와 host bind/firewall가 identity/access owner다. 애플리케이션은
여전히 실제 peer CIDR, canonical Host, TLS와 상태변경 Origin/Referer+CSRF를
검사한다. 재시작 handoff token은 사용자 인증이 아니라 프로세스 교체 무결성을 위한
내부 단발성 자격으로만 남긴다.

대안인 password hash DB, Passkey와 OAuth provider는 단일 operator 도구에 사용자
저장소, RP domain 또는 외부 availability를 추가한다. 사용자는 이 추가 복잡도를
원하지 않고 VPN 인증을 재사용하도록 선택했다.

### D3. 동일 origin을 server가 결정한다

원격 mode의 모든 요청은 `Host == PublicURL.Host`여야 한다. 상태 변경 POST는 기존
CSRF 검사 전에 `Origin`이 canonical origin과 exact match해야 하며, Origin이 없는
클라이언트만 same-origin Referer로 보완한다. 둘 다 없거나 다른 경우 거부한다.
local native mode는 loopback listener 자체가 경계이므로 이 추가 검사를 적용하지 않는다.

응답에는 CSP(`default-src 'self'; form-action 'self'; frame-ancestors 'none'`),
`X-Content-Type-Options: nosniff`, `Referrer-Policy: no-referrer`,
`X-Frame-Options: DENY`, 민감 화면 `Cache-Control: no-store`를 붙이고 원격 mode에는
HSTS를 추가한다. HTTP server는 TLS 1.3 minimum, header/read/idle timeout과 header
size limit를 갖는다.

### D4. network boundary와 action audit는 fail-closed다

trusted-network에는 login endpoint가 없다. 실제 peer CIDR과 exact Host는 handler보다
먼저 검사하고 forwarding header는 신뢰하지 않는다. 설정·engine·verify 등 기존
상태변경 action audit는 유지하며, application login 제거가 LIVE 승인 또는 운영
토글 승인으로 해석되지 않는다.

`/healthz`는 container healthcheck를 위한 exact GET/HEAD 예외다. 고정된 `ok`만
반환하고 account/config/session 상태를 노출하지 않으며, 상태 변경 route가 아니다.

### D5. Compose는 host VPN IP publish와 최소 container 권한을 강제한다

multi-stage Go build 후 Alpine runtime에는 binary와 CA certificate만 둔다. runtime은
non-root numeric UID/GID, read-only root filesystem, `cap_drop: ALL`,
`no-new-privileges`, PID/resource/log limit, init, tmpfs `/tmp`를 사용한다.

Compose의 `ports`는
`${TOSSOS_VPN_BIND_IP:?required}:${TOSSOS_CONSOLE_PORT:-37085}:37085`이며
wildcard fallback이 없다. config/data는 명시적 host directory에 persist하고,
broker session·TLS cert/key는 host secret file에서 read-only mount한다.
container의 system-update는 사용하지 않고 `docker compose build/pull && up -d`로
교체한다.

## Threat Model

상세 DFD와 STRIDE/DREAD 표는 `analysis/threat-model.md`가 정본이다. DREAD 7 이상
항목의 mitigation owner는 다음과 같다.

- network membership theft: VPN admin의 계정/기기 폐기 + console TLS/CIDR/Host gate
- unintended exposure: operator(VPN bind/CIDR) + Compose required-variable guard
  + console peer-CIDR gate
- Host/CSRF tampering: console exact Host/Origin + existing CSRF
- secret/image disclosure: operator TLS/broker 0600 secret files + container read-only mounts
  + repository secret scan
- container privilege escalation: Dockerfile/Compose non-root, capability drop,
  no-new-privileges

## Risks / Trade-offs

- [VPN 계정 또는 VPN 접속 기기가 탈취되면 console 전체 권한을 얻는다] → 사용자가
  VPN을 sole access identity로 명시 승인했다. VPN 계정/기기 폐기와 host bind/firewall를
  운영 통제로 사용하며 public/wildcard publish는 계속 금지한다.
- [인증서가 모바일에서 신뢰되지 않으면 접속할 수 없다] → VPN DNS 이름/IP가 든
  운영자 CA 또는 VPN 제공 인증서를 사용한다. `--insecure` 우회는 제공하지 않는다.
- [container port NAT가 peer IP를 보존하지 않는 환경] → health 외 요청이 CIDR
  검사에서 fail-closed된다. 해당 플랫폼에서는 host의 VPN IP에 native bind하거나
  firewall가 보장된 별도 배포 설계를 사용한다.
- [application identity가 없어 세밀한 사용자 attribution이 없다] → 현재
  single-operator 범위만 지원하고 audit subject는 VPN/OS operator + peer IP다.
  다중 사용자는 별도 IdP/RBAC change다.
- [container self-update가 비활성] → image provenance와 rollback이 더 명확한
  Compose image 교체를 사용한다.

## Migration Plan

1. local/remote no-login RED tests를 먼저 작성한다.
2. explicit trusted-network 선택과 session bypass를 최소 구현한다.
3. Compose에서 remote-token secret을 제거하고 인증서·VPN CIDR만 준비한다.
4. Compose config/build/health/dashboard smoke test를 broker account mutation 없이 수행한다.
5. 실제 운영에서는 host loopback과 VPN client에서 `/healthz`와 dashboard direct access를
   검증한다. rollback은 trusted-network 플래그를 제거하고 기존 token-auth image/config로
   되돌린다.

## Open Questions

없음. 실제 VPN bind IP/CIDR, public HTTPS URL과 인증서는 배포 시 운영자가 제공하는
환경값이며 저장소가 추측하지 않는다.
