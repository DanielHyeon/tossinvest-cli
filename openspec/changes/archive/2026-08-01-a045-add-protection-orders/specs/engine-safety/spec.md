## ADDED Requirements

### Requirement: ProtectionReady는 attestation 범위에서만 WIRED다
엔진은 현재 계좌·profile·시장·주문유형·수량·세션·trigger source와 atomic/continuous replace semantics가 strict versioned attestation과 일치하고 protection saga가 배선된 경우에만 exposure-raising mutation을 허용해야 한다 (SHALL). attestation은 tool/build와 evidence digest에 묶이고 legacy/unknown field, 만료, 경로·소유자·권한 불일치는 fail-closed해야 한다 (SHALL).

#### Scenario: 미검증 시장
- **WHEN** KR capability만 attested된 상태에서 US 자동 진입을 시도한다
- **THEN** entry는 protection_unwired로 거부되고 기존 US 보유의 reduce-only exit는 계속된다

#### Scenario: 유효 capability
- **WHEN** 현재 profile이 attestation과 일치하고 Guardian/gate가 유효하다
- **THEN** protection readiness clause는 충족되지만 운영자가 승인하지 않은 lane는 여전히 OFF다
