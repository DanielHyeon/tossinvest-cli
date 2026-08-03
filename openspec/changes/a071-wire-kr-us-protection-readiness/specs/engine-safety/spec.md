## MODIFIED Requirements

### Requirement: ProtectionReady는 attestation 범위에서만 WIRED다
엔진은 ProtectionReady를 exact attestation 범위에서만 WIRED로 계산해야 한다 (SHALL). KR과
US 각각에 대해 현재 계좌·profile·시장·주문유형·수량·세션·trigger source,
atomic/continuous replace semantics 및 exact broker identity/query/dedup/idempotency capability가
strict versioned signed attestation과 일치하고 protection supervisor/saga가 production
assembly에 배선된 경우에만 해당 시장의 exposure-raising mutation을 허용해야 한다 (SHALL).
Attestation은 tool/build와 evidence digest, issued-at, expires-at, monotonic serial, key ID,
allowlisted algorithm과 signature에 묶이고 pin된 trust root, revocation, bounded rotation overlap,
maximum lifetime와 durable trusted-time floor를 통과해야 한다 (SHALL). legacy/unknown field,
capability 누락, 만료, signature·digest·경로·소유자·권한 불일치, serial/time rollback은 해당
시장에서 fail-closed해야 한다 (SHALL). 한 시장의 불일치는 다른 시장의 유효한 readiness를
낮춰서는 안 되지만 (MUST NOT), 실패한 시장의 신규 진입을 허용해서도 안 된다 (MUST NOT).
Readiness 판정은 운영 토글, lane, autostart, automation gate 또는 LIVE approval을
생성·변경해서는 안 된다 (MUST NOT). 외부 readiness enum은 정확히 `WIRED`와 `UNWIRED`만
허용해야 하며 (SHALL), 모든 실패 상세는 enum 변형 대신 stable typed refusal과 provenance로
제공해야 한다 (SHALL).

#### Scenario: 미검증 시장
- **WHEN** KR capability만 attested된 상태에서 US 자동 진입을 시도한다
- **THEN** entry는 protection_unwired로 거부되고 기존 US 보유의 protection, reconciliation과 reduce-only exit는 계속된다

#### Scenario: 유효 capability
- **WHEN** 현재 KR profile이 KR attestation과 일치하고 Guardian/gate가 유효하다
- **THEN** KR protection readiness clause는 `WIRED`지만 운영자가 승인하지 않은 KR lane는 여전히 OFF다

#### Scenario: US readiness와 KR 실패의 독립성
- **WHEN** US attestation과 wiring은 유효하지만 KR attestation은 만료됐다
- **THEN** KR 신규 진입만 protection_unwired로 거부되고 US readiness verdict는 유지되며 두 시장의 exit와 reconciliation은 계속된다

#### Scenario: decision 뒤 attestation 만료
- **WHEN** EntryDecision 생성 뒤 durable dispatch 직전에 해당 시장 attestation이 만료된다
- **THEN** broker request는 0건이고 해당 시장 readiness를 `UNWIRED`로 낮추며 기존 protection과 reduce-only exit를 유지한다

#### Scenario: signed evidence 부재
- **WHEN** 보호 saga 코드는 배선됐지만 current signed attestation이 없다
- **THEN** 두 시장 readiness는 `UNWIRED`이고 테스트 fixture 외의 live 보호주문으로 자동 검증하지 않는다

#### Scenario: readiness enum 계약
- **WHEN** attestation 또는 wiring 검증이 실패한다
- **THEN** readiness는 정확히 `UNWIRED`이고 `READY`, `DEGRADED`, `UNKNOWN` 같은 제3 enum 대신 typed refusal을 반환한다

#### Scenario: trust state 실패의 시장 격리
- **WHEN** KR attestation key는 revoked됐지만 US attestation은 별도 active key와 증가한 serial로 유효하다
- **THEN** KR만 `UNWIRED`와 typed key refusal이고 US는 `WIRED`를 유지하며 두 시장의 기존 protection과 exit는 계속된다
