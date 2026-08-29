## MODIFIED Requirements

### Requirement: 테스트 동반 구현
기능 구현은 해당 기능을 검증하는 테스트가 같은 change 안에 존재하고 통과하는 상태로만 완료될 수 있다(SHALL). 각 Requirement는 최소 1개의 테스트 또는 검증 명령으로 추적 가능해야 한다(SHALL).
여기서 "통과"는 **저장소 자신의 게이트가 실행한** 통과여야 한다(SHALL). build tag 뒤에 있어 완료 게이트도 CI도 실행하지 않는 테스트는 이 요구를 만족시키지 못한다(SHALL NOT). 따라서 완료 게이트와 CI는 무태그 실행과 `tossos_testseams` 태그 실행을 **둘 다** 수행해야 한다(SHALL).
테스트가 상수에서 유도되는 값을 단언할 때는 그 값을 상수에서 계산해야 하며(SHALL), 계산 결과를 리터럴로 굳혀서는 안 된다(SHALL NOT) — 상수를 바꾸는 change가 아무 신호도 받지 못하기 때문이다.

#### Scenario: 기능 커밋 리뷰
- **WHEN** 기능 커밋을 리뷰하면
- **THEN** 해당 기능의 테스트가 같은 change 안에 존재하고 통과한다

#### Scenario: 게이트가 실행하지 않는 테스트
- **WHEN** `tossos_testseams` 태그 뒤에만 존재하는 테스트가 실패하는 상태로 완료 게이트를 실행하면
- **THEN** 게이트가 실패한다

#### Scenario: 상수에서 유도되는 기대값
- **WHEN** 어떤 테스트가 두 런타임 상수의 비로 결정되는 호출 횟수를 단언하고 그 상수 중 하나가 바뀌면
- **THEN** 그 테스트는 새 상수로 계산한 값을 기대하고 통과하거나, 계약이 실제로 깨졌을 때 실패한다

### Requirement: 자동화된 완료 게이트
change 완료 선언은 `make gate CHANGE=<change-id>`(tasks.md 미완료 항목 0건 + review.md 존재 + test·test-seams·vet·validate 통과) 성공과 Manager의 diff 리뷰·독립 테스트 재실행 이후에만 가능하다(SHALL). task 완료 체크는 그 산출물을 만드는 커밋과 같은 커밋에서 수행해야 한다(SHALL). 완료된 change는 `openspec archive`로 확정 스펙에 반영한다.

#### Scenario: 미완료 task가 있는 완료 시도
- **WHEN** tasks.md에 미완료 체크박스가 남은 상태로 gate를 실행하면
- **THEN** 게이트가 실패하고 미완료 항목이 출력된다

#### Scenario: 태그 스위트가 깨진 상태의 완료 시도
- **WHEN** 무태그 스위트는 통과하지만 `tossos_testseams` 태그 스위트가 실패하는 상태로 gate를 실행하면
- **THEN** 게이트가 실패한다
