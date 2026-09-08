## MODIFIED Requirements

### Requirement: Full SDD 권위 계층

모든 신규 기능·동작 변경은 OpenSpec 계약, 현재 HEAD와 CodeGraph hard evidence, CodeGraphContext 보조 문맥, 기존 함수 내부 변경 시 Go AST·ast-grep Function Logic Map, Superpowers TDD, gstack 게이트 순서로 수행되어야 한다(SHALL). CodeGraphContext·GBrain·기억·관측 그래프는 advisory이며 OpenSpec, 현재 HEAD, 테스트, gstack을 대체해서는 안 된다(SHALL NOT).
일반 변경의 함수 분석 비교 기준은 불변 `base-commit.txt`여야 한다(SHALL).

함수 분석 비교의 **대상 쪽 끝**은 그 change 의 작업이 착지한 지점이어야 하며(SHALL),
그 지점이 기록된 change 에 대해 워킹트리를 대상으로 삼아서는 안 된다(SHALL NOT).
착지 지점이 기록되지 않은 change 는 워킹트리를 대상으로 한다(SHALL) — 작업 중인
change 의 판정은 바뀌지 않아야 한다.

이 규칙이 필요한 이유는 병합이다. 대상이 워킹트리이면, base 와 워킹트리 사이에 들어온
다른 change 의 함수가 이 change 가 고친 함수로 집계되어, 어떤 change 도 답할 수 없는
질문이 된다. 특히 배포 후 실측 태스크는 정의상 다른 작업이 착지한 뒤에 닫히므로 그런
태스크를 가진 모든 change 가 이 상태에 빠진다.

착지 지점의 기록은 위조 가능해서는 안 된다(SHALL NOT). 그 값은 비교 대상을 정하는
유일한 손잡이이므로, 유효한 값의 조건과 기록 시점이 명시되어야 한다(SHALL).

아래 명시적 legacy 실행 기준선 이관의 모든 조건을 만족하는 a063만 고정된 실행
기준 E로 판정할 수 있으며, 원래 구현 전 증거 규칙을 소급 충족했다고 보고해서는
안 된다(SHALL NOT). 원래 계획 기준 P는 계속 보존해야 한다(SHALL).

#### Scenario: 병합 뒤에 닫히는 배포 후 실측 태스크

- **WHEN** 어떤 change 의 작업이 착지하고 그 지점이 기록된 뒤, 다른 change 들이 트리에
  들어오고, 그 change 의 배포 후 실측 태스크를 닫아 완료 게이트를 실행하면
- **THEN** 5단계는 그 change 가 실제로 고친 함수에 대해서만 증거를 요구하고, 사이에
  들어온 다른 change 의 함수를 요구하지 않는다

#### Scenario: 착지 지점이 없는 작업 중 change

- **WHEN** 착지 지점이 기록되지 않은 change 로 완료 게이트를 실행하면
- **THEN** 5단계는 지금과 같이 `base-commit.txt` 와 워킹트리를 비교한다

#### Scenario: 위조된 착지 지점

- **WHEN** 착지 지점 기록이 유효 조건을 만족하지 않는 값을 담고 있으면
- **THEN** 5단계는 통과하지 않고 그 사유를 이름으로 말한다
