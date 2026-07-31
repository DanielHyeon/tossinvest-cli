## ADDED Requirements

### Requirement: 후보 필터는 판정 의미와 근거를 함께 표시한다
콘솔은 a050의 `candidate-filters` 카테고리에서 `seen_late`, `extended`, `near_high`를 시장·세션별로 구분하고 각 필터의 쉬운 설명, 판정 방향, 단위, 기본 상태, desired/effective 값, 범위, 표본과 evidence provenance를 표시해야 한다 (SHALL).

#### Scenario: 최초 미승인 상태
- **WHEN** 승인된 threshold set이 없다
- **THEN** 화면은 `미승인 · passed 구조적 0 · verdict 비활성`을 표시하고 숫자 0을 기본 threshold처럼 보여주지 않는다

#### Scenario: evidence 불완전
- **WHEN** sample count 또는 evidence digest가 누락됐다
- **THEN** 관련 입력은 read-only이고 누락 항목과 승인에 필요한 다음 행동을 설명한다

#### Scenario: 승인된 시장별 값
- **WHEN** KR regular-session threshold set이 승인됐다
- **THEN** 각 metric의 값·단위·방향·표본·version을 표시하고 US에는 같은 값을 기본값으로 복제하지 않는다

### Requirement: threshold 승인은 변경 영향 preview를 선행한다
콘솔은 threshold apply 전에 before/after, 대상 시장·세션, 예상 verdict count 변화, evidence version과 적용 시점을 preview해야 한다 (SHALL).

#### Scenario: 승인 preview
- **WHEN** 운영자가 완전한 threshold set을 승인하려 한다
- **THEN** 후보 판정만 활성화되고 주문·RiskIntent·LIVE 상태는 바뀌지 않음을 확인 전에 설명한다
