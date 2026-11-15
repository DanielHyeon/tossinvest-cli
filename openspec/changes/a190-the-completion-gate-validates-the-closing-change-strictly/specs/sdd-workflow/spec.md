## MODIFIED Requirements

### Requirement: 자동화된 완료 게이트
change 완료 선언은 `make gate CHANGE=<change-id>`(tasks.md 미완료 항목 0건 + review.md 존재 + test·test-seams·vet 통과 + 닫는 change 의 `openspec validate <change-id> --strict` 통과 + 저장소 전체 `openspec validate --all` 오류 0) 성공과 Manager의 diff 리뷰·독립 테스트 재실행 이후에만 가능하다(SHALL). 판정할 수 없는 validate는 실패로 친다(SHALL). 저장소 전체 strict 결과는 판정이 아니지만 gate 출력에 합계가 남아야 한다(SHALL). task 완료 체크는 그 산출물을 만드는 커밋과 같은 커밋에서 수행해야 한다(SHALL). 완료된 change는 `openspec archive`로 확정 스펙에 반영한다.

#### Scenario: 미완료 task가 있는 완료 시도
- **WHEN** tasks.md에 미완료 체크박스가 남은 상태로 gate를 실행하면
- **THEN** 게이트가 실패하고 미완료 항목이 출력된다

#### Scenario: 태그 스위트가 깨진 상태의 완료 시도
- **WHEN** 무태그 스위트는 통과하지만 `tossos_testseams` 태그 스위트가 실패하는 상태로 gate를 실행하면
- **THEN** 게이트가 실패한다

#### Scenario: 남의 strict 경고만 있는 완료 시도
- **WHEN** 닫는 change 는 strict 로 valid 이고 다른 change 나 main spec 에만 strict 경고가 있는 상태로 gate를 실행하면
- **THEN** validate 단계는 통과하고, 저장소 전체 strict 합계가 판정이 아니라는 표시와 함께 출력된다

#### Scenario: 닫는 change 의 strict 경고
- **WHEN** 닫는 change 에 strict 경고가 있는 상태로 gate를 실행하면
- **THEN** 게이트가 validate 단계에서 실패한다

#### Scenario: 저장소 어딘가의 오류
- **WHEN** 닫는 change 는 valid 이지만 다른 change 나 spec 이 strict 없이도 invalid 인 상태로 gate를 실행하면
- **THEN** 게이트가 validate 단계에서 실패한다

#### Scenario: 판정할 수 없는 validate
- **WHEN** openspec 을 실행할 수 없거나, 닫는 change 를 openspec 이 찾지 못하거나, 저장소 전체 검증이 0 항목을 검사했으면
- **THEN** 게이트가 validate 단계에서 실패한다
