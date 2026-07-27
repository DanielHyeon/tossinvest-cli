# sdd-workflow Specification (delta)

## ADDED Requirements

### Requirement: OpenSpec 기반 변경 관리
신규 기능·동작 변경·안전 경로(주문 실행, 위험관리, 원장, reconciliation) 수정은 OpenSpec change로 정의된 후에만 구현되어야 하며(SHALL), 구현 전 `openspec validate <change> --strict`를 통과해야 한다. 오탈자·주석·문서 수정, 동작 불변 의존성 patch, 테스트만 추가하는 변경은 change 없이 가능하다. 보안·자금 위험의 긴급 수정은 즉시 구현하되 24시간 내 사후 change로 문서화해야 한다(SHALL).

#### Scenario: 스펙 없는 기능 구현 시도
- **WHEN** openspec change 없이 신규 기능 코드가 제출되면
- **THEN** Manager 리뷰에서 반려된다

#### Scenario: 긴급 보안 수정
- **WHEN** 자금 위험이 있는 결함을 즉시 수정하면
- **THEN** 24시간 내 해당 수정을 기술하는 사후 change가 생성된다

### Requirement: 역할 분리
Manager(총괄 아키텍트)는 스펙 작성·검토·검증만 수행하고 구현 코드는 Teammate(구현 에이전트)가 작성해야 한다(SHALL). 구현을 생성한 컨텍스트와 이를 검증하는 컨텍스트는 별도 세션으로 분리되어야 한다(SHALL). Teammate는 blocking 수준 스펙 결함 발견 시 구현을 중단하고 `openspec/changes/<id>/issues.md`에 기록 후 Manager에 보고해야 한다(SHALL).

#### Scenario: blocking 스펙 결함 발견
- **WHEN** Teammate가 안전·동작에 영향을 주는 스펙 모순을 발견하면
- **THEN** 구현을 중단하고 issues.md에 기록하며, Manager가 스펙을 수정한 뒤 재개한다

### Requirement: 등급제 문서 리뷰 게이트
change의 proposal·design·spec 델타는 첫 구현 task 착수 전(proposal-freeze) gstack 리뷰를 1회 통과해야 하며(SHALL), 이후 Requirement 수준의 스펙 수정은 수정분에 대한 리뷰를 재실행해야 한다(SHALL). 리뷰 결과(일시, 보이스 구성, 발견 요약, 수용/거절과 근거)는 `openspec/changes/<change-id>/review.md`에 기록되어야 한다(SHALL). tasks.md 상태 갱신·오탈자·리뷰 결정 반영은 면제된다.

#### Scenario: 리뷰 기록 없이 구현 착수 시도
- **WHEN** review.md가 없는 change에 대해 `make gate CHANGE=<id>`를 실행하면
- **THEN** 게이트가 실패한다

### Requirement: 테스트 동반 구현
기능 구현은 해당 기능을 검증하는 테스트가 같은 change 안에 존재하고 통과하는 상태로만 완료될 수 있다(SHALL). 각 Requirement는 최소 1개의 테스트 또는 검증 명령으로 추적 가능해야 한다(SHALL).

#### Scenario: 기능 커밋 리뷰
- **WHEN** 기능 커밋을 리뷰하면
- **THEN** 해당 기능의 테스트가 같은 change 안에 존재하고 통과한다

### Requirement: 자동화된 완료 게이트
change 완료 선언은 `make gate CHANGE=<change-id>`(tasks.md 미완료 항목 0건 + review.md 존재 + test·vet·validate 통과) 성공과 Manager의 diff 리뷰·독립 테스트 재실행 이후에만 가능하다(SHALL). task 완료 체크는 그 산출물을 만드는 커밋과 같은 커밋에서 수행해야 한다(SHALL). 완료된 change는 `openspec archive`로 확정 스펙에 반영한다.

#### Scenario: 미완료 task가 있는 완료 시도
- **WHEN** tasks.md에 미완료 체크박스가 남은 상태로 gate를 실행하면
- **THEN** 게이트가 실패하고 미완료 항목이 출력된다

### Requirement: 실계좌 보호의 기계적 강제
자동 테스트는 실계좌 주문을 발생시켜서는 안 되며(SHALL NOT), 이는 규칙이 아니라 테스트 인프라로 강제되어야 한다(SHALL): 테스트는 격리된 임시 config 디렉터리에서 실행되고, 실 endpoint 호출은 httptest 대체 없이는 구성될 수 없어야 한다. 실계좌 검증은 사용자 승인 하의 수동 절차로만 수행한다.

#### Scenario: 주문 로직 테스트 실행
- **WHEN** 주문 관련 테스트가 실행되면
- **THEN** 격리된 config 경로와 httptest 서버만 사용되며 실제 API 주문 호출은 발생하지 않는다
