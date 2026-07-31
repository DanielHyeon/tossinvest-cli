## ADDED Requirements

### Requirement: 자동 진입은 모든 안전 권한의 교집합이다
엔진은 automation gate, operating mode, lane state, Guardian, reconciliation health와 ProtectionReady가 모두 허용할 때만 strategy entry를 제출해야 한다 (SHALL).
이 조건들은 동일한 immutable activation manifest에 version/digest/expiry로 결합돼야 하며 (SHALL), durable dispatch 직전에 전부 재검증되지 않으면 신규 진입을 제출해서는 안 된다 (MUST NOT).

#### Scenario: protection 미배선
- **WHEN** 다른 조건이 허용돼도 ProtectionReady가 UNWIRED다
- **THEN** buy는 거부되고 reduce-only exit는 계속된다

#### Scenario: kill switch
- **WHEN** kill switch가 활성화된다
- **THEN** 신규 entry를 즉시 중지하고 기존 보호·청산 감독은 유지한다

#### Scenario: 승인 manifest 불일치
- **WHEN** decision 뒤 dispatch 전에 threshold, settings, attestation, Guardian, scheduler 또는 build digest가 바뀐다
- **THEN** 신규 attempt는 제출되지 않고 effective entry는 OFF와 구체적 refusal reason을 기록한다

#### Scenario: manifest 만료 뒤 재시작
- **WHEN** 저장된 desired state는 ON이지만 activation manifest가 만료됐다
- **THEN** 재시작은 승인 상태를 재구성하지 않고 entry OFF를 유지하며 exit/reconcile은 계속한다
