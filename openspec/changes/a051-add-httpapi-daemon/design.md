## Context

현재 console은 동일 프로세스의 HTML handler와 meta refresh를 사용한다. 모바일/VPN 소비자는 안정적인 versioned JSON과 stream이 필요하다.

## Goals / Non-Goals

**Goals:** 기존 domain service를 재사용하는 private read REST/SSE daemon과 authenticated 제한 write contract를 제공한다.

**Non-Goals:** 공용 인터넷, 사용자 계정·로그인 UI, 별도 모바일 앱, domain logic 복제, remote LIVE/gate/protection 약화 mutation.

## Decisions

1. `/api/v1` adapter는 console과 같은 domain/query seams를 사용하고 journal/broker를 직접 중복 조립하지 않는다.
2. SSE event ID는 process epoch와 epoch-local monotonic sequence를 포함한다. 재시작 또는 unknown
   `Last-Event-ID`는 새 epoch의 full snapshot으로 수렴한다.
3. 기존 token/session mode는 유지한다. 명시적으로 설정한 no-token mode는 loopback/configured VPN CIDR의
   read-only routes만 연다. VPN native write는 enrolled mTLS identity 또는 local human approval channel이
   발급한 one-time nonce를 actor/method/resource/canonical body digest/expiry에 묶어 서명한 60초 단기
   capability가 있을 때만 제한 route를 연다. nonce는 한 번 사용되면 즉시 폐기한다.
4. write canonical origin은 scheme+host+effective port만 비교하고 path/query는 무시한다.
5. browser write는 session+CSRF+origin, native write는 mTLS/서명 capability를 요구한다. 둘 다
   `actor/client + method + resource + canonical body digest + idempotency key` scope, `If-Match`, audit와
   narrow commander를 사용한다. 동일 key의 다른 body는 409다. LIVE/gate/kill-switch/protection 약화와
   activation-manifest write는 route 자체를 제공하지 않는다.
6. optimization resource는 a050 `OptimizationCategory`와 `OptimizationFieldDescriptor`를 공통 DTO로 사용한다. 웹 HTML과 모바일 API가 category 이름·기본값·설명을 별도 상수로 복제하지 않는다.
7. category 목록은 `overview`, `exit-protection`, `position-management`, `candidate-filters`, `strategy-runtime`, `performance-history` 고정 순서이며 응답은 default state/value, desired/effective, constraint, apply timing, help와 provenance를 포함한다.
8. 기본 운영 한도는 SSE client 32, client queue 64 event, heartbeat 15초, queue-full 즉시 disconnect,
   request body 256 KiB, header/read timeout 5초다. 값은 server-owned config이며 더 넓은 값은 startup에서 거부한다.
9. `internal/httpapi`는 journal/broker writer를 직접 갖지 않고 a043/a050 read model과 narrow commander만 사용한다.
10. Origin/Host/trusted-network 판정은 console에서 `internal/networkboundary`로 추출해 둘이 공유한다.
    strict Origin precedence, opaque-origin 거부, configured trusted proxy의 정확한 hop만 forwarded header
    사용을 허용하는 기존 fail-closed 의미를 보존한다.

## Risks / Trade-offs

- [VPN 오설정으로 외부 노출] → startup에서 bind/CIDR/TLS forwarding 설정을 fail-closed 검증한다.
- [HTML과 API drift] → 공통 DTO/domain adapters와 contract test를 사용한다.
- [SSE 폭주] → 32 clients, 64-event bounded queue, 15초 heartbeat, queue-full disconnect를 적용한다.
- [웹과 모바일의 기본값·설명 drift] → 공통 descriptor registry와 golden contract fixture를 사용한다.

## Migration Plan

daemon/compose service를 기존 console과 함께 배포하고 API consumer가 없을 때 동작 무변경을 보장한다. rollback은 API-owned pending control command가 0임을 확인한 뒤 API service만 중지하며 engine/console은 유지한다.

## Open Questions

모바일 앱 구현은 별도 Story에서 API 승인 후 다룬다.
