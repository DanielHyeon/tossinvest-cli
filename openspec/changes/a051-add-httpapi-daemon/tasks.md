## 1. API 계약과 RED

- [x] 1.1 console/domain seams, remote origin, engine control과 compose 영향 지도를 작성한다.
- [x] 1.2 REST schema/error, optimization golden contract, SSE epoch/sequence/restart/gap/32-client/64-queue/15초 heartbeat,
  private bind, no-token read-only VPN, unauth mutation 404/405, daemon signed capability-only auth와 shared guard browser/mTLS 보존, origin,
  one-time 60초 nonce, actor/method/resource/body-digest idempotency/If-Match와 body/timeout RED 테스트를 추가한다.

## 2. Service adapter

- [x] 2.1 `internal/httpapi` DTO/router를 공통 domain seams 위에 구현한다.
- [x] 2.2 `/api/v1` read resources와 stable error/schema version을 구현하고 optimization에 a050 공통 category·descriptor DTO를 노출한다.
- [x] 2.3 SSE epoch+sequence, full snapshot, heartbeat, bounded queue, slow-consumer disconnect와 reconnect를 구현한다.

## 3. 안전한 writes와 배포 문서

- [x] 3.1 shared guard의 browser session+canonical origin+CSRF 및 enrolled mTLS 지원은 보존하되 a051 daemon에는
  signed capability mode만 wiring하고, actor-scoped idempotency, CAS, audit와 narrow commander 및 remote
  LIVE/gate/protection 약화 route 부재를 정적 검증한다.
- [x] 3.2 console origin logic을 shared `internal/networkboundary`로 추출하고 loopback/VPN bind와 exact trusted-proxy-hop/TLS startup validation을 구현한다.
- [x] 3.3 daemon/compose/autostart/rollback 운영 문서를 추가하되 공용 인터넷 exposure를 만들지 않는다.

## 4. 검증

- [x] 4.1 contract/race/load/full test·vet·validate와 security/adversarial review를 통과한다.
- [x] 4.2 `make gate CHANGE=a051-add-httpapi-daemon`을 통과한다.
