## Why

TossOS 콘솔은 `127.0.0.1` 전용이어서 VPN에 연결된 모바일 브라우저에서는 사용할 수
없다. 단일 운영자는 호스트 loopback 또는 이미 인증된 VPN을 접근 신뢰 경계로
사용하기로 승인했으므로, 별도 application token/login을 반복하지 않고도 콘솔을
열되 네트워크 노출·TLS·요청 출처·상태변경 안전을 명시적으로 강제해야 한다.

## What Changes

- **BREAKING** 현재 Compose의 application token 로그인과 remote session 요구를 제거하고,
  사용자가 승인한 명시적 trusted-network 모드로 전환한다.
- 기존 native 기본값은 loopback 전용으로 보존하고, 명시적으로 구성된 원격 모드만
  non-loopback listener를 허용한다.
- trusted-network 원격 모드는 bind 주소, 허용 VPN CIDR, HTTPS 인증서/키,
  canonical public URL이 모두 유효해야 기동하며 token/login/session을 요구하지 않는다.
- 모든 원격 요청의 Host를 canonical URL에 고정하고, 상태 변경 요청은 기존 CSRF에
  더해 Origin/Referer가 같은 HTTPS origin인지 확인한다.
- 기존 콘솔의 설정·엔진 시작/중지·검증 등 전체 기능을 원격 세션에도 제공하되,
  LIVE 주문·운영 토글의 기존 사람 승인과 엔진 interlock은 변경하지 않는다.
- multi-stage Dockerfile과 Docker Compose 예제를 추가한다. Compose는 필수
  `TOSSOS_VPN_BIND_IP`에만 host port를 게시하며 wildcard host publish 기본값을
  제공하지 않고, secret/certificate/session은 image나 저장소에 포함하지 않는다.
- 컨테이너의 프로세스·파일 권한·healthcheck·resource/log 제한과 이미지 기반
  업데이트 경로를 문서화한다.

## Capabilities

### New Capabilities

- `vpn-console-access`: VPN 경계의 원격 listener, TLS, trusted-network 접근,
  Host/Origin 및 fail-closed 기동 계약.
- `container-deployment`: TossOS 콘솔/엔진의 non-root image, VPN-IP 전용 Compose
  publish, secret mount, persistence 및 container lifecycle 계약.

### Modified Capabilities

- `operator-console`: Compose의 loopback/VPN trusted-network 접근에는 application
  login을 요구하지 않고, 동일한 기능·CSRF·감사 경계를 적용한다.

## Impact

- `internal/console`: listener policy, TLS serving, trusted-network middleware,
  Host/Origin 검증, 보안 응답 헤더와 health endpoint.
- `cmd/tossctl`: `console`의 명시적 원격 접근 플래그와 fail-closed 조립.
- `Dockerfile`, `.dockerignore`, `compose.yaml`, `.env.example`: reproducible
  non-root container 및 VPN 주소 한정 배포.
- `docs/operations.md`: OpenVPN 등 VPN 주소 선택, 인증서 준비, Compose
  기동·교체·복구 절차.
- 새 외부 서비스나 인증 제공자 의존성은 추가하지 않는다.

PM mapping: `STORY-TOS-031`과 이 change는 1:1로 연결된다. bootstrap 예외를 사용하지
않는다.
