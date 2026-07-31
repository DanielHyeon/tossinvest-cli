## ADDED Requirements

### Requirement: 정책 lifecycle은 generation에 귀속된다
journal은 override, release와 re-adopt event를 position/adoption generation에 귀속하고 과거 generation의 상태를 새 lifecycle에 적용해서는 안 된다 (MUST NOT).

#### Scenario: 늦게 도착한 과거 event
- **WHEN** release 후 새 generation이 열린 뒤 과거 generation의 event가 도착한다
- **THEN** event를 격리하고 새 exit state를 변경하지 않는다
