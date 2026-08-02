## MODIFIED Requirements

### Requirement: 서비스는 versioned REST와 SSE를 제공한다
The daemon SHALL provide engine, positions, orders, candidates, performance, settings,
optimization resources and SSE under `/api/v1` with a stable JSON schema. It SHALL accept
bodyless HTTP/1.1 and HTTP/2 GET/HEAD requests equally, and SHALL fail closed when a read or
stream request carries a declared or unknown-length body.

#### Scenario: body 없는 HTTP/2 조회
- **WHEN** VPN 모바일 클라이언트가 body 없는 HTTP/2 GET 또는 HEAD로 고정 resource를 조회한다
- **THEN** body가 있다고 오판하지 않고 HTTP/1.1과 같은 resource 결과를 반환한다

#### Scenario: HTTP/2 body 거부
- **WHEN** HTTP/2 read 또는 stream 요청이 declared 또는 unknown-length body를 전송한다
- **THEN** stable `BODY_NOT_SUPPORTED` 오류를 반환하고 resource/stream handler를 실행하지 않는다
