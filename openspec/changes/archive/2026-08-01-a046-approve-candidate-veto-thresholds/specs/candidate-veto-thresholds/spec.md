## ADDED Requirements

### Requirement: veto threshold set은 근거와 승인 범위를 가진다
각 threshold set은 market, session, metric definition, decimal values, sample window/count, missing rate, evidence digest, version과 approver를 포함해야 한다 (SHALL).

#### Scenario: 완전한 승인 set
- **WHEN** KR regular-session set이 모든 필드와 evidence digest를 가진다
- **THEN** KR 정규장 verdict는 해당 version으로 세 veto를 평가할 수 있다

#### Scenario: 다른 시장 재사용
- **WHEN** KR set만 존재하는 상태에서 US 후보를 평가한다
- **THEN** verdict는 unmeasured/fail-closed이고 KR 값을 재사용하지 않는다

### Requirement: 승인 전에는 candidate pass를 만들지 않는다
시스템은 `seen_late`, `extended`, `near_high` 중 하나라도 미승인·누락이면 cleared-all verdict를 생성해서는 안 된다 (MUST NOT).

#### Scenario: 일부 threshold 누락
- **WHEN** near_high만 있고 seen_late와 extended가 비어 있다
- **THEN** passed count는 0이고 누락 reason을 각 후보에 기록한다

#### Scenario: 승인 후 pass
- **WHEN** 완전한 set 아래 모든 관측이 veto 조건을 벗어난다
- **THEN** candidate-life ID, threshold version, set/evidence digest와 approved-at을 포함한 ApprovedCandidate verdict를 만들되 주문·RiskIntent는 만들지 않는다

#### Scenario: 승인 set 아래 veto 또는 미측정
- **WHEN** 완전한 set을 사용했지만 하나 이상의 veto가 raised 또는 unmeasured다
- **THEN** ApprovedCandidate는 zero value이고 typed refusal을 반환한다

#### Scenario: candidate-life provenance
- **WHEN** 같은 normalized market/symbol과 같은 UTC instant의 FirstSeenAt을 다시 평가한다
- **THEN** raw symbol을 포함하지 않는 동일한 deterministic candidate-life ID를 만들고, FirstSeenAt이 달라지면 다른 ID를 만든다
