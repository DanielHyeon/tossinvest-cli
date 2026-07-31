## ADDED Requirements

### Requirement: 전역 정책 변경은 활성 포지션을 재해석하지 않는다
시스템은 common policy 변경만으로 기존 exit state의 policy ID/version을 변경해서는 안 된다 (MUST NOT).

#### Scenario: 공통값 변경
- **WHEN** 활성 BALANCED 포지션이 있는 동안 공통값을 RUNNER로 저장한다
- **THEN** 기존 포지션은 BALANCED snapshot을 유지하고 새 lifecycle만 RUNNER를 사용한다
