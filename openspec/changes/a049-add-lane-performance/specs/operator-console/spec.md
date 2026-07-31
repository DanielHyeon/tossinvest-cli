## ADDED Requirements

### Requirement: 성과와 이력은 읽기 전용 카테고리에서 설명된다
콘솔은 a050의 `performance-history` 카테고리에서 lane/policy 성과와 설정 변경 이력을 읽기 전용으로 제공하고 각 metric의 쉬운 정의, 단위, 기간, 표본 수와 provenance를 표시해야 한다 (SHALL).

#### Scenario: 최초 조회
- **WHEN** 운영자가 별도 filter 저장 없이 카테고리를 연다
- **THEN** 최근 30일, 전체 시장, 전체 lane, 완전한 lineage만 포함하는 조회 기본값을 표시한다

#### Scenario: 누락 결과
- **WHEN** 한 거래의 결정적 lineage 또는 markout이 없다
- **THEN** 0으로 표시하지 않고 각각 `link_missing` 또는 `not_measured` 설명을 표시한다

#### Scenario: 표본 부족
- **WHEN** 선택 구간의 표본이 승인 최소치 미만이다
- **THEN** 관측값과 표본은 보여주되 `insufficient_sample · 추천 근거로 사용 불가`를 표시한다

### Requirement: 성과 화면은 거래나 설정 권한을 갖지 않는다
`performance-history`는 주문, lane toggle, LIVE approval 또는 설정 apply control을 제공해서는 안 된다 (MUST NOT).

#### Scenario: 성과 비교
- **WHEN** 운영자가 lane 두 개를 비교한다
- **THEN** 비교 결과만 표시하고 더 좋은 lane를 자동 활성화하거나 저장하지 않는다
