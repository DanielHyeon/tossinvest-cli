# a051 · 모바일용 HTTP API daemon 추가

## 0. 백로그 계층 추적

- **Initiative**: `INIT-TOS-002`
- **Epic**: `EPIC-TOS-003`
- **Feature**: `FEAT-TOS-012`
- **Story**: `STORY-TOS-a051`

## Why

현재 Go 콘솔은 HTML/meta-refresh 중심이라 VPN 모바일 클라이언트가 안정적인 schema와 실시간 상태를 사용할 수 없다. no-token local/VPN 읽기와 신원이 검증된 제한 쓰기를 분리해야 한다.

## What Changes

- `/api/v1` 아래 engine, positions, orders, candidates, performance, settings, optimization REST를 정의한다.
- `/api/v1/stream` SSE와 schema/sequence/reconnect 계약을 추가한다.
- local/VPN read-only REST/SSE는 애플리케이션 토큰을 요구하지 않되 private bind와 TLS termination을 요구한다.
- 쓰기 origin은 `scheme + host + port`만 비교하고 path는 비교하지 않는다.
- browser write는 기존 session+CSRF+origin을 유지하고 VPN native write는 mTLS identity 또는 서명된 단기 capability, idempotency, `If-Match`와 audit를 요구한다.
- LIVE/gate/kill-switch/protection 약화/activation-manifest mutation은 remote API에 제공하지 않는다.
- optimization API는 a050과 동일한 여섯 category ID·순서, field 설명, 기본 상태/값, desired/effective 값, 제약, 적용 시점과 provenance를 반환해 모바일에서도 웹과 같은 의미를 사용한다.
- **비목표**: 공용 인터넷 노출, 사용자 계정 시스템, 모바일 앱 자체 구현.

## Capabilities

### New Capabilities

- `http-api-service`: versioned REST/SSE, 오류 schema와 private-network 운영 계약.

### Modified Capabilities

- `console-request-origin`: 브라우저와 API 쓰기의 canonical origin 비교를 공유하도록 확장한다.

## Impact

- 신규 `internal/httpapi`, daemon command/compose service, console adapters와 API contract tests.
- no-token read boundary와 authenticated write boundary를 분리하고 console/httpapi의 journal read-only 불변을 유지한다.
