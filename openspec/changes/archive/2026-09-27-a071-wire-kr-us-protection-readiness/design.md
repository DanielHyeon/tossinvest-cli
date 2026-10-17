## Context

Production engine은 `ProtectionReady`를 요구하지만 현재 profile은 보호 배선을
`UNWIRED`로 고정한다. 기존 보호 saga와 official conditional-order adapter는 출발점일
뿐, KR과 US의 주문 유형·세션·수량·trigger·replace 동작이 현재 build에서 검증됐다는
권한은 아니다. 이 change는 보호 실행 코드, 외부에서 생성된 capability evidence와
production assembly 사이의 신뢰 경계를 닫는다.

보호 mutation은 공식 Open API만 사용하며 사람 승인 없는 live 검증은 수행하지 않는다.
배포 기본값은 attestation이 없는 `UNWIRED`이고, readiness는 lane·automation·LIVE 승인을
대신하지 않는다.

## Goals / Non-Goals

**Goals:**

- signed current attestation의 정확한 scope에서만 KR·US별 readiness를 계산한다.
- 보호 등록·교체·취소·복구가 재시도와 프로세스 손실 뒤에도 한 broker-resident 보호로
  수렴한다.
- 한 시장의 capability 실패가 다른 시장의 유효한 readiness를 오염시키지 않게 한다.
- live endpoint를 사용하지 않는 결정적 fixture와 격리 broker test로 배선을 검증한다.

**Non-Goals:**

- 운영 토글, lane, autostart, automation gate 또는 LIVE approval 활성화.
- 사람 승인 없는 capability 실측이나 live 보호주문 제출.
- attestation이 증명하지 않은 주문 유형·세션·수량 또는 replace 의미의 추정 지원.
- 기존 exit, flatten 또는 reconciliation 권한의 축소.

## Decisions

### 1. Readiness는 global boolean이 아니라 market-scoped capability verdict다

검증 입력은 account/profile, market, order type, session, quantity range, trigger source,
replace semantics, exact broker conditional-order identity/query/dedup/idempotency capability,
tool/build digest, evidence digest, issued/expires-at, monotonic serial, key ID, signature
algorithm과 signature를 포함하는 strict versioned document다. Broker capability에는 최소한
client operation key의 전달·echo 여부, 그 key 또는 broker order ID로 하는 exact lookup 필드,
order identity uniqueness scope, pending/terminal 상태 조회, cancel 결과 조회와 중복 제출 시
broker 동작이 명시돼야 한다. 하나라도 없거나 모호하면 해당 scope는 `UNWIRED`다.

Loader의 trust root는 배포 manifest에 pin된 root key set이다. 허용 signature algorithm은
명시적 allowlist만 사용하고 key ID가 active/rotation-overlap key에 속하는지 확인한다.
서명 시점 또는 현재 시점에 revoked key인 문서, 이전에 수락한 serial 이하의 문서,
설정된 최대 수명을 넘는 문서와 trusted-time source보다 미래인 문서는 거부한다. Key rotation은
명시된 짧은 overlap window에서 old/new key를 함께 허용하되 serial은 계속 단조 증가해야 한다.
Trusted time이 unavailable하거나 마지막 trusted time보다 후퇴하면 신규 readiness를
`UNWIRED`로 내리고 기존 broker protection과 reduce-only exit만 보존한다. 마지막 수락 serial과
trusted-time floor는 durable anti-rollback state로 저장한다. 파일의 canonical body,
signature, 소유자·권한도 함께 검증하며 unknown/legacy field와 digest 불일치를 거부한다.

대안인 `KR=true`/`US=true` 설정값이나 코드 존재 여부 기반 판정은 세션·수량·replace
의미를 잃고 운영자가 증명하지 않은 권한을 만들기 때문에 배제한다.

### 2. KR과 US readiness snapshot은 독립적으로 계산한다

하나의 attestation set에서 시장별 verdict와 typed refusal을 만들고 immutable runtime
snapshot으로 publish한다. 해당 시장의 exact attestation과 supervisor wiring이 모두
유효할 때만 `WIRED`다. 외부 enum은 정확히 `WIRED` 또는 `UNWIRED`만 사용하고, 실패 원인은
별도 typed refusal code와 evidence provenance로 제공한다. KR 실패는 US verdict를 낮추지 않고 그 반대도 같지만, 실패한
시장의 exposure-raising dispatch는 거부한다.

결합 `KR_US_READY` 값은 어느 시장의 근거가 사용됐는지 숨기므로 만들지 않는다.

### 3. 보호 lifecycle은 durable desired-state convergence로 배선한다

position generation과 protection revision이 operation identity를 소유한다. 등록, 더 안전한
trigger/수량으로의 교체, 취소와 복구는 journal의 desired/observed broker identity를 기준으로
재시도한다. 재기동은 저장된 broker ID를 먼저 조회하며 시간·심볼 추정으로 새 주문을 만들지
않는다.

Submit 결과가 timeout/transport loss로 unknown이면 attestation이 증명한 client operation key와
exact lookup을 사용해 accepted/not-found/terminal을 판별한다. Attestation이 broker-side
idempotency와 exact dedup을 모두 증명한 경우에만 동일 operation key 재제출을 허용한다.
그렇지 않으면 상태를 `RECONCILE_REQUIRED`로 고정하고 신규 제출과 exposure를 차단한다.
Cancel 결과가 unknown이면 exact broker ID로 ACTIVE/CANCELED/FILLED를 조회하며, 확인 전에는
취소됐다고 간주하거나 대체 보호를 제출하지 않는다.

Broker 조회에서 durable desired state에 없는 orphan conditional order가 발견되면 해당
시장/position entry latch를 닫고 typed orphan refusal을 발행한다. Attested identity와
account/market/position generation까지 정확히 귀속할 수 있을 때만 adoption/cancel policy를
적용한다. 귀속 불가능한 orphan은 추정 취소하지 않고 safety reconciliation과 사람 개입에
남긴다.

교체는 attested atomic 또는 continuous-coverage 의미가 oversell과 무보호 window를 모두
막는 경우에만 허용한다. 그 의미를 증명하지 못하면 현재 ACTIVE 보호를 유지하고 신규 노출을
차단한다. `cancel then place`를 일반 fallback으로 사용하는 대안은 보호 공백을 만들기 때문에
배제한다.

### 4. Runtime loss와 capability drift는 entry를 닫고 보호 감독을 보존한다

broker-resident ACTIVE 주문은 engine process와 독립적으로 유지된다. 재기동 시 observed
broker state와 durable desired state가 수렴하기 전 entry latch는 닫힌다. 실행 중 attestation
만료·build drift가 감지되면 해당 시장의 신규 노출만 즉시 차단하고 기존 protection,
reconciliation과 reduce-only exit는 계속된다.

### 5. 검증은 live transport를 구조적으로 제외한다

signature, expiry, scope matrix, duplicate event, crash point와 recovery test는 고정 fixture와
`httptest` official broker로 수행한다. 실제 Toss hostname에 대한 mutation은 transport guard가
실패시킨다. 새로운 실측 evidence가 필요하면 기존 사람 승인 verification workflow가 생성하고
이 change의 loader는 그 결과만 소비한다.

## Risks / Trade-offs

- [Broker replace 의미가 시장별로 다름] → exact attested semantics 외에는 `UNWIRED`로
  유지하고 일반 fallback을 만들지 않는다.
- [Attestation 만료로 운용 중 entry가 닫힘] → 시장별 typed refusal과 expiry를 노출하고
  protection/exit/reconciliation은 계속한다.
- [재시도 중 중복 매도 청구권] → generation/revision operation identity와 broker ID 조회,
  holdings/reservation 합계 검증으로 제출 전에 차단한다.
- [Submit/cancel 결과 unknown] → attested exact identity/query로만 대사하고 broker
  idempotency가 증명되지 않으면 재제출하지 않는다.
- [Trust key 교체 또는 clock rollback] → pin된 root, allowlist, revocation, overlap,
  monotonic serial과 durable trusted-time floor로 fail-closed한다.
- [Rollback binary가 새 attestation schema를 모름] → unknown schema를 fail-closed하고 기존
  broker-resident 보호를 자동 취소하지 않는다.

## Migration Plan

1. attestation loader와 시장별 readiness projector를 dormant 상태로 배포한다.
2. durable protection supervisor를 격리 broker fixture로 검증하고 production assembly에
   연결한다.
3. 기존 설치는 attestation 부재로 두 시장 모두 `UNWIRED`를 유지한다.
4. 별도 사람 승인 workflow가 pin된 trust root와 key lifecycle을 따라 생성한 current signed
   attestation만 운영 profile에 설치한다.
5. rollback은 신규 entry를 OFF로 유지하고 이미 broker에 상주하는 보호를 취소하지 않은 채
   이전 read/reconcile 경로로 복귀한다.

## Open Questions

없음. 아직 생성되지 않은 실제 KR/US evidence는 구현을 추정 지원으로 바꾸지 않으며 해당
scope를 `UNWIRED`로 유지한다.
