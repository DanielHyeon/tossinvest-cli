## ADDED Requirements

### Requirement: 서비스는 versioned REST와 SSE를 제공한다
daemon은 `/api/v1` 아래 engine, positions, orders, candidates, performance, settings, optimization resource와 `/api/v1/stream` SSE를 안정된 JSON schema로 제공해야 한다 (SHALL).

#### Scenario: positions 조회
- **WHEN** VPN 모바일 클라이언트가 `/api/v1/positions`를 조회한다
- **THEN** schema version과 a043 exit-line fields를 포함한 JSON을 반환한다

#### Scenario: SSE 재연결
- **WHEN** stream 연결이 끊긴 뒤 클라이언트가 재연결한다
- **THEN** 같은 process epoch의 연속 event면 다음 sequence를 받고, epoch mismatch·gap·재시작이면 새 epoch의 full snapshot으로 최신 상태에 수렴한다

### Requirement: optimization API는 웹과 같은 메뉴·기본값·설명을 제공한다
optimization resource는 a050과 동일한 여섯 category ID와 순서, 각 field의 label, description, type, unit, default state/value, desired/effective 값, range/choices, owner, apply timing, safety direction과 provenance를 반환해야 한다 (SHALL).

#### Scenario: 모바일 메뉴 구성
- **WHEN** VPN 모바일 클라이언트가 optimization schema를 조회한다
- **THEN** `overview`, `exit-protection`, `position-management`, `candidate-filters`, `strategy-runtime`, `performance-history`를 웹과 같은 순서와 사용자 설명으로 반환한다

#### Scenario: 외부 매수 자동편입 기본값
- **WHEN** 저장 설정이 없는 환경에서 position-management descriptor를 조회한다
- **THEN** adoption OFF, default stop 5%, range 2~20%/step 0.5%, 빈 include/exclude와 exclude 우선 설명을 반환한다

#### Scenario: 미승인 후보 필터
- **WHEN** threshold evidence가 승인되지 않았다
- **THEN** default state는 `unapproved`이고 임의 숫자 0을 default value로 반환하지 않는다

#### Scenario: 웹·API drift 검사
- **WHEN** category 또는 descriptor golden contract를 검증한다
- **THEN** HTML adapter와 API adapter는 같은 registry를 사용하고 별도 기본값·도움말 상수를 갖지 않는다

### Requirement: local/VPN no-token 접근은 읽기 전용이다
서비스는 configured loopback/VPN private network의 read-only resource와 SSE에 별도 access token·login을 요구해서는 안 되며 (MUST NOT), no-token mode에서 mutation route를 제공해서는 안 된다 (MUST NOT).
public bind 또는 신뢰하지 않은 forwarded origin은 fail-closed로 거부해야 한다 (MUST).

#### Scenario: VPN 접근
- **WHEN** configured VPN CIDR의 클라이언트가 TLS endpoint에 접근한다
- **THEN** app token 없이 읽기 REST/SSE를 사용할 수 있고 mutation은 404/405다

#### Scenario: public bind 설정
- **WHEN** service가 private boundary 없이 wildcard public interface에 bind되도록 설정된다
- **THEN** startup을 거부하고 수정 방법을 기록한다

### Requirement: 제한 API 쓰기는 신원·감사·동시성 계약을 유지한다
mutation endpoint는 browser session+CSRF+origin 또는 mTLS identity/서명된 단기 capability를 요구해야 한다 (SHALL). `actor/client + method + resource + canonical body digest + idempotency key` scope, `If-Match`, audit와 narrow command service를 적용해야 한다 (SHALL). signed capability는 local human channel의 one-time nonce와 60초 expiry에 묶여야 한다 (SHALL).
LIVE/gate/kill-switch/protection 약화와 activation-manifest mutation endpoint를 제공해서는 안 된다 (MUST NOT).

#### Scenario: stale settings write
- **WHEN** 인증된 native client가 오래된 `If-Match`로 허용된 설정을 저장한다
- **THEN** 412와 current version을 반환하고 자동 재시도하거나 부분 저장하지 않는다

#### Scenario: LIVE toggle
- **WHEN** 모바일이 engine LIVE state를 변경하려 한다
- **THEN** route가 존재하지 않아 404/405이고 local human approval channel만 사용할 수 있다

#### Scenario: idempotency body 충돌
- **WHEN** 같은 actor/client와 idempotency key를 다른 canonical body로 다시 사용한다
- **THEN** 409를 반환하고 두 번째 command를 실행하지 않는다

#### Scenario: native capability 재사용
- **WHEN** 이미 소비했거나 60초가 지난 signed capability를 다시 제출한다
- **THEN** 인증을 거부하고 mutation과 audit commit을 만들지 않는다

### Requirement: SSE와 HTTP resource는 정량 한도를 가진다
daemon은 기본 최대 SSE client 32, client당 queue 64 event, heartbeat 15초와 queue-full disconnect를 강제해야 한다 (SHALL). request body 256 KiB와 header/read timeout 5초도 강제해야 한다 (SHALL).

#### Scenario: 느린 SSE client
- **WHEN** 한 client의 queue가 64 event를 넘는다
- **THEN** 다른 client나 producer를 막지 않고 해당 client만 끊으며 재연결 시 full snapshot으로 수렴한다
