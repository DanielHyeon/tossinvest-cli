## ADDED Requirements

### Requirement: 공통 정책 설정은 기동 시 fail-closed 검증된다
엔진은 non-empty 공통 policy ID가 registry에 없거나 policy가 ordering/ratio/runner 조건을 위반하면 exit observer 기동을 거부해야 한다 (SHALL).

#### Scenario: 손상된 common policy 설정
- **WHEN** config가 알 수 없는 common policy ID를 포함한다
- **THEN** 엔진은 조용히 RATCHET으로 후퇴하지 않고 이유를 포함해 기동을 거부한다

### Requirement: 공통 정책은 위험 축소 권한만 사용한다
공통 정책이 만드는 모든 주문 proposal은 기존 Guardian의 reduce-only issuance와 execution gateway를 거쳐야 하며 설정 승인이 LIVE order 승인이나 exposure 증가 권한으로 해석되어서는 안 된다 (MUST NOT).

#### Scenario: HYBRID_50 부분익절
- **WHEN** HYBRID_50 rung이 부분익절을 제안한다
- **THEN** 기존 reduction decision, reservation, idempotency, submit 경로를 사용하고 신규 buy 권한을 만들지 않는다

#### Scenario: automation gate OFF
- **WHEN** 공통 정책이 저장돼 있지만 automation gate가 OFF다
- **THEN** 설정은 보존되지만 unattended exit observer는 기존 interlock에 따라 기동하지 않는다
