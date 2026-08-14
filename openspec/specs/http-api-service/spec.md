# http-api-service Specification

## Purpose
TBD - created by archiving change a051-add-httpapi-daemon. Update Purpose after archive.
## Requirements
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

### Requirement: 제한 API 쓰기는 capability-only 신원·감사·동시성 계약을 유지한다
a051 daemon의 mutation endpoint는 local human approval channel이 발급한 signed capability만 인증 수단으로 요구해야 한다 (SHALL). shared mutation guard가 지원하는 browser session+CSRF+origin 및 enrolled mTLS identity는 이 daemon에 wiring하거나 OpenAPI 인증 수단으로 광고해서는 안 된다 (MUST NOT). `actor/client + method + resource + canonical body digest + idempotency key` scope, `If-Match`, audit와 narrow command service를 적용해야 한다 (SHALL). signed capability는 one-time nonce와 최대 60초 expiry 및 canonical HTTPS audience에 묶여야 한다 (SHALL).
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

#### Scenario: daemon에 wiring하지 않은 인증 mode
- **WHEN** client가 signed capability 없이 browser session/CSRF 또는 client certificate만 제시한다
- **THEN** a051 daemon은 mutation identity를 인정하지 않고 401을 반환하며 command를 실행하지 않는다

### Requirement: SSE와 HTTP resource는 정량 한도를 가진다
daemon은 기본 최대 SSE client 32, client당 queue 64 event, heartbeat 15초와 queue-full disconnect를 강제해야 한다 (SHALL). request body 256 KiB와 header/read timeout 5초도 강제해야 한다 (SHALL).

#### Scenario: 느린 SSE client
- **WHEN** 한 client의 queue가 64 event를 넘는다
- **THEN** 다른 client나 producer를 막지 않고 해당 client만 끊으며 재연결 시 full snapshot으로 수렴한다

### Requirement: positions와 optimization API는 웹과 같은 adoption/reconcile 사실을 사용한다
`/api/v1/positions`와 `/api/v1/optimization`은 웹 position-management와 동일한 registry default, config desired, running-engine effective, candidate와 adoption-blocking tracker projector를 사용해야 한다 (SHALL). block DTO는 모든 journal cause가 아니라 adoption driver와 같은 runtime projection임을 source로 명시해야 한다 (SHALL). read API는 reconcile resolution capability 또는 mutation route를 노출해서는 안 된다 (MUST NOT).

positions item은 stable `adoptionStatus` enum(`UNKNOWN`, `MANAGED`, `EXCLUDED`, `RECONCILE_BLOCKED`, `ADOPTION_PENDING`, `UNMANAGED`), `statusKnown`, `adoptionLabel`, typed/sanitized `adoptionReason`, `included`, `excluded`, `candidate`, `designationKnown`과 nullable covering block(scope/market/symbol/reason/startedAt)을 반환해야 한다 (SHALL). optimization position-management는 desired/effective adoption blocks와 `effectiveKnown`을 반환해야 한다 (SHALL).

#### Scenario: 미국 include 보유분이 영구 차단으로 대기한다
- **WHEN** 미국 보유분이 include됐고 account-wide permanent quantity-mismatch block이 active다
- **THEN** positions API는 eligible false, candidate true, adoptionStatus `RECONCILE_BLOCKED` 및 sanitized block reason을 반환하고 optimization API는 actual desired/effective adoption 값을 반환한다

#### Scenario: managed와 exclude가 함께 있다
- **WHEN** managed position의 symbol이 exclude에도 있다
- **THEN** positions API의 adoptionStatus는 `MANAGED`이고 included/excluded 사실은 별도 boolean으로 보존된다

#### Scenario: released lifecycle은 raw adoption eligibility보다 우선한다
- **WHEN** journal position에는 adoption ID가 남아 있지만 authoritative lifecycle은 `RELEASED`다
- **THEN** positions API는 `UNMANAGED`, `OPERATOR_RELEASED`를 반환하고 `MANAGED`, `ADOPTION_PENDING`, `RECONCILE_BLOCKED`로 오표시하지 않는다

#### Scenario: API sidecar는 별도 network namespace에서 runtime을 읽는다
- **WHEN** console/engine과 HTTP API가 Compose의 서로 다른 network namespace에서 같은 private engine directory를 mount한다
- **THEN** HTTP API는 command loopback이 아니라 authenticated runtime-only Unix endpoint로 effective와 block projection을 읽고 Preview/Apply 권한을 얻지 않는다

#### Scenario: engine runtime을 읽지 못한다
- **WHEN** config desired는 읽히지만 engine control plane이 unavailable이다
- **THEN** optimization API는 effectiveKnown false를 반환하고 registry 기본값을 effective로 위장하지 않으며, non-managed positions item은 statusKnown/designationKnown false와 typed runtime-unavailable reason을 반환한다

#### Scenario: read API의 mutation 표면은 그대로다
- **WHEN** a052가 배포된다
- **THEN** HTTP API mutation allowlist에는 기존 optimization preview/application만 남고 reconcile 해제 endpoint는 없다

#### Scenario: raw exit evidence와 effective line
- **WHEN** legacy exit state에 raw t0/baseline은 있지만 canonical effective snapshot이 없다
- **THEN** positions API의 exitLine actionable 가격은 unknown을 유지하고 storedExitEvidence는 raw 값과 effectiveKnown false를 반환한다

### Requirement: 컨테이너 entrypoint mode는 checkout filesystem과 독립적이다
배포 image는 source checkout이나 Git executable bit에 의존하지 않고 entrypoint를 실행 가능한 mode로 설치해야 한다 (SHALL). non-root runtime identity는 Compose 재생성 후 entrypoint를 실행할 수 있어야 한다 (SHALL).

#### Scenario: NTFS checkout에서 image 재빌드
- **WHEN** entrypoint source가 `0644`인 checkout에서 image를 재빌드한다
- **THEN** image의 `/usr/local/bin/tossos-entrypoint`는 `0755`이고 service는 exit 126 없이 healthcheck까지 기동한다

### Requirement: positions API는 비실효 기준선 참조를 별도 계약으로 제공한다
`GET /api/v1/positions`는 actionable `exitLine`과 별도로 nullable `exitLineReference`를 제공해야 한다 (SHALL). reference는 `LEGACY_RAW`, `ADOPTION_PLAN` 또는 generation/runtime/lifecycle unknown 상태를 typed kind로 표시하고 항상 `effectiveKnown=false`여야 한다 (SHALL). lifecycle generation이 다르거나 lifecycle을 검증할 수 없으면 이전 가격이나 identity를 반환해서는 안 된다 (MUST NOT).

#### Scenario: legacy raw evidence
- **WHEN** KR 또는 US 포지션에 same-generation raw exit state만 있다
- **THEN** `exitLine`의 current/next 가격은 `—`이고 `exitLineReference`와 호환 `storedExitEvidence`에 non-effective 원장 근거가 반환된다

#### Scenario: US adoption plan
- **WHEN** US candidate가 pending 또는 reconcile-blocked이고 running effective stop percentage가 알려져 있다
- **THEN** `exitLineReference.kind=ADOPTION_PLAN`, stop percentage와 가격 미확정 설명을 반환하며 계산된 가격은 반환하지 않는다

#### Scenario: generation mismatch
- **WHEN** current lifecycle generation이 stored exit generation과 다르다
- **THEN** `exitLine`, `storedExitEvidence`, `exitLineReference` 어디에도 과거 가격이나 snapshot identity가 없다

#### Scenario: corrupt 또는 lifecycle-unverified evidence
- **WHEN** raw exit tuple의 snapshot 상태가 partial/invalid/corrupt이거나 요구된 lifecycle lookup이 현재 generation을 증명하지 못한다
- **THEN** API는 raw 가격을 반환하지 않고 typed unknown reason만 제공한다

### Requirement: positions API는 flat 관측의 canonical freshness를 console과 공유한다
`GET /api/v1/positions`는 console과 같은 persisted effective snapshot 및 shared freshness 판정을 사용해야 한다 (SHALL): engine stopped가 확정되면 즉시 stale이고, running·unavailable·unwired에서는 integrity와 per-position 30초 age bound를 함께 적용한다. 성공한 flat observation refresh가 계속되는 managed position의 `exitLine`은 fresh/actionable이어야 하며 (SHALL), API adapter가 별도 시계나 raw seed 가격으로 line을 재계산해서는 안 된다 (MUST NOT).

#### Scenario: flat quote가 계속 관측되는 managed position
- **WHEN** position의 가격과 policy state는 변하지 않지만 마지막 성공 refresh가 freshness bound 안이다
- **THEN** API는 console과 같은 current protection, next target, next protection, freshness status와 evaluated-at을 반환한다

#### Scenario: seed에서 첫 flat 평가가 완료된다
- **WHEN** `SEED` position이 adoption t0와 같은 첫 공식 quote로 canonical evaluation을 영속한다
- **THEN** API는 `not_evaluated_yet`을 해제하고 완전한 actionable exitLine을 반환한다

#### Scenario: invalid quote만 도착한다
- **WHEN** 마지막 성공 evaluation 뒤 invalid 또는 stale quote만 도착해 persisted freshness가 age bound를 넘는다
- **THEN** API는 actionable 가격을 숨기고 console과 같은 typed stale/unknown reason을 반환한다

#### Scenario: console과 API의 30초 경계
- **WHEN** running·unavailable·unwired인 동일 snapshot을 29.999초, 정확히 30초와 30초 초과 시각에 두 adapter로 projection한다
- **THEN** 두 surface는 앞의 두 시각에 fresh, 초과 시각에만 stale이며 line visibility와 typed reason이 동일하다

#### Scenario: engine running과 stopped 판정
- **WHEN** 같은 snapshot과 runtime을 두 adapter가 running 또는 stopped로 확정한다
- **THEN** running도 30초 age bound를 적용하고 stopped는 둘 다 즉시 `engine_not_running`이다

#### Scenario: API blocking read 중 freshness 경계를 지난다
- **WHEN** 실제 positions 요청의 cache·journal·policy·runtime 또는 단일 engine-marker read가 진행되는 동안 snapshot age가 30초를 넘거나 marker가 stopped 경계를 지난다
- **THEN** API는 모든 blocking read 뒤의 한 response clock으로 판정해 즉시 stale/dash를 반환하며 broker cache를 다시 읽지 않는다

#### Scenario: stopped marker 뒤 wall clock이 rollback한다
- **WHEN** marker read는 engine을 stopped로 판정했지만 그 직후 response clock이 뒤로 움직여 marker refresh time이 다시 bound 안처럼 보인다
- **THEN** API는 pre-read stopped 판정을 running으로 승격하지 않고 `engine_not_running`과 dash를 유지한다

#### Scenario: invalid sibling은 running liveness를 빌리지 않는다
- **WHEN** running engine의 batch에서 다른 symbol은 유효하지만 이 position의 quote가 계속 invalid/missing이다
- **THEN** API와 console은 이 position의 마지막 성공 관측만 aging하여 30초 초과 뒤 같은 stale verdict로 숨긴다

### Requirement: positions API의 flat refresh는 read-only history 의미를 보존한다
API가 반환하는 최신 observed-at은 durable snapshot refresh 근거여야 하며 (SHALL), 의미가 동일한 refresh를 새 exit event, proposal 또는 order로 노출해서는 안 된다 (MUST NOT).

#### Scenario: 여러 flat refresh 뒤 API 조회
- **WHEN** 한 position이 동일 line으로 여러 번 refresh된 뒤 positions와 exit history를 조회한다
- **THEN** positions는 최신 observed-at을 반환하고 history는 첫 평가 이후 의미 없는 중복 event를 포함하지 않는다
