## ADDED Requirements

### Requirement: 보호 capability는 시장별 signed attestation으로 제한된다
시스템은 시장별 signed protection capability 계약을 강제해야 한다 (SHALL). 현재 account,
profile, market, order type, session, quantity range, trigger source, replace semantics, exact broker
conditional-order identity/query/dedup/idempotency capability, tool/build digest와 evidence digest가
current signed attestation에 정확히 포함된 경우에만 해당 KR 또는 US scope의 broker-resident
protection을 등록·교체해야 한다 (SHALL). Broker capability는 client operation key 전달/echo,
exact lookup field, identity uniqueness scope, pending/terminal status query, cancel-result query와
duplicate-submit behavior를 모두 명시해야 한다 (SHALL). unknown/legacy field, capability 누락,
만료, signature·digest·파일 소유자·권한 불일치는 해당 시장의 신규 보호 capability를
`UNWIRED`로 만들어야 한다 (SHALL). Attestation 검증은 운영 토글, lane, autostart,
automation gate 또는 LIVE approval을 변경해서는 안 된다 (MUST NOT).

#### Scenario: KR만 유효한 attestation
- **WHEN** current signed attestation이 KR regular-session protection만 포함한다
- **THEN** KR scope만 보호 등록 가능하고 US 신규 진입은 `protection_unwired`로 거부되며 기존 US protection과 reduce-only exit는 계속된다

#### Scenario: scope 밖 수량
- **WHEN** US protection 요청 수량이 attested quantity range를 벗어난다
- **THEN** broker mutation 없이 요청을 거부하고 해당 포지션의 신규 exposure latch를 닫는다

#### Scenario: 만료 또는 build drift
- **WHEN** ACTIVE 보호가 있는 동안 attestation이 만료되거나 current build digest와 달라진다
- **THEN** 기존 broker-resident 보호를 취소하지 않고 신규 등록·약화·exposure raising을 차단하며 reconciliation과 reduce-only exit를 계속한다

#### Scenario: broker dedup capability 누락
- **WHEN** US attestation이 주문 형식은 포함하지만 client operation key echo 또는 exact lookup/dedup semantics를 포함하지 않는다
- **THEN** US readiness는 `UNWIRED`이고 typed `broker_identity_capability_unattested` refusal을 반환하며 신규 broker mutation은 0건이다

### Requirement: attestation trust는 pin된 key와 단조 시간 계약을 따른다
시스템은 protection attestation을 pin된 trust root와 단조 시간 계약으로 검증해야 한다 (SHALL). Attestation은 active key ID, 명시적 signature algorithm allowlist, revocation 상태,
bounded key-rotation overlap, durable monotonic serial, configured maximum lifetime와 trusted-time
source를 모두 통과해야 한다 (SHALL). 이미 수락한 serial 이하, revoked key, 허용되지 않은
algorithm, overlap 밖 old key, 최대 수명 초과, 미래 issued-at, trusted-time unavailable 또는
durable trusted-time floor보다 후퇴한 clock은 해당 시장을 `UNWIRED`로 만들고 typed refusal을
반환해야 한다 (SHALL). 이 실패는 기존 broker-resident protection을 자동 취소해서는 안 된다
(MUST NOT).

#### Scenario: serial rollback
- **WHEN** 같은 market/key scope에서 마지막 수락 serial보다 작거나 같은 attestation을 로드한다
- **THEN** 해당 시장은 `UNWIRED`와 `attestation_serial_rollback` refusal을 반환하고 신규 보호 mutation은 0건이다

#### Scenario: revoked rotation key
- **WHEN** signature는 수학적으로 유효하지만 key ID가 revoked됐거나 rotation overlap을 벗어났다
- **THEN** attestation을 거부하고 기존 protection, reconciliation과 reduce-only exit를 유지한다

#### Scenario: trusted time rollback
- **WHEN** trusted time을 얻을 수 없거나 현재 trusted time이 durable time floor보다 과거다
- **THEN** 신규 readiness와 protection mutation을 `UNWIRED`로 차단하고 clock 추정으로 만료를 연장하지 않는다

### Requirement: 보호 lifecycle은 하나의 broker-resident 보호로 멱등 수렴한다
시스템은 보호 lifecycle을 하나의 broker-resident protection으로 멱등 수렴시켜야 한다 (SHALL). Position generation과 protection revision에 결합된 stable operation identity로 등록,
수량 증가, 더 안전한 trigger 교체, 취소, 재기동 복구와 reconciliation을 영속해야 한다
(SHALL). duplicate event와 crash retry는 저장된 broker identifier를 먼저 조회해 같은 desired
protection으로 수렴해야 하며 (SHALL), symbol/time 근사로 두 번째 보호를 생성해서는 안 된다
(MUST NOT). 교체는 attested semantics가 무보호 window와 보유 수량 초과 매도 청구권을 모두
방지하는 경우에만 수행해야 한다 (SHALL).

Submit/cancel 결과가 unknown이면 시스템은 attested exact identity/query로 broker 상태를
대사해야 한다 (SHALL). 동일 operation key 재제출은 attestation이 broker-side idempotency와
exact dedup을 증명할 때만 허용해야 하며 (SHALL), 그 증명이 없으면 no-resubmit
`RECONCILE_REQUIRED`로 유지해야 한다 (SHALL). Durable desired state에 없는 orphan protection은
entry를 차단해야 하며 (SHALL), account/market/position generation 소유권을 exact identity로
확인하지 못하면 adopt, cancel 또는 replace해서는 안 된다 (MUST NOT).

#### Scenario: 등록 응답 전 프로세스 손실
- **WHEN** 보호 등록 요청이 broker에 도달한 뒤 응답을 영속하기 전에 엔진 프로세스가 종료된다
- **THEN** 재기동은 stable operation identity와 attested broker 조회로 기존 주문을 귀속하고 exact query/dedup이 없으면 재제출하지 않는다

#### Scenario: 교체 의미 미검증
- **WHEN** 현재 ACTIVE 보호는 있으나 attestation이 atomic 또는 continuous-coverage replace 의미를 증명하지 못한다
- **THEN** 현재 보호를 유지하고 broker mutation과 추가 exposure를 0건으로 만들며 typed reconcile reason을 기록한다

#### Scenario: 중복 fill event
- **WHEN** 같은 position generation과 filled quantity를 알리는 이벤트가 반복 전달된다
- **THEN** protection revision과 broker-reserved sell claim은 한 번만 전진하고 보호 수량은 보유 수량을 초과하지 않는다

#### Scenario: 엔진 프로세스 손실
- **WHEN** ACTIVE protection을 확인한 뒤 엔진 프로세스가 사라진다
- **THEN** attested broker-side 주문은 계속 상주하고 재기동 전까지 신규 exposure는 발생하지 않는다

#### Scenario: submit outcome unknown without attested idempotency
- **WHEN** protection submit 뒤 transport가 끊기고 attestation이 exact lookup 또는 broker-side idempotency를 증명하지 않는다
- **THEN** operation은 `RECONCILE_REQUIRED`로 남고 재제출과 신규 exposure는 0건이며 기존 safety reconciliation을 계속한다

#### Scenario: cancel outcome unknown
- **WHEN** broker cancel 요청 결과를 받지 못했다
- **THEN** exact broker ID로 ACTIVE/CANCELED/FILLED를 확인하기 전에는 취소 성공으로 간주하거나 replacement를 제출하지 않는다

#### Scenario: 소유권 불명 orphan protection
- **WHEN** broker 조회에서 durable desired state에 없고 position generation으로 exact 귀속할 수 없는 conditional order가 발견된다
- **THEN** 해당 scope의 entry를 차단하고 typed orphan refusal을 기록하며 주문을 추정 취소하지 않는다
