# sdd-workflow Specification (delta)

## ADDED Requirements

### Requirement: 최상위 안전 불변식
docs/WORKFLOW.md §0의 안전 불변식은 모든 방법론·스펙보다 우선한다(SHALL). 특히: 개발·테스트 중 승인 없는 LIVE 주문 side-effect 금지, 토글 OFF 시 upstream 동작 보존, 손절·비상 청산 즉시성 약화 금지, 손절·익절·사이징 변경은 보수 방향만 허용(불명확 시 변경 금지), 운영 토글 flip은 사람 승인 필수.

#### Scenario: 사이징 로직 완화 변경 제출
- **WHEN** 위험 기반 수량 계산을 더 공격적으로 바꾸는 변경이 명확한 근거 없이 제출되면
- **THEN** 안전 불변식 §0.9 위반으로 반려된다

### Requirement: High-risk Pre-Edit 선언
High-risk 경로(주문 제출·취소·정정, 손절/사이징, Guardian, journal·원장, reconciliation, retry matrix, 인증·세션, 체결 감지)의 기존 코드를 수정하기 전에 Teammate는 Pre-Edit 선언(change/task id, 대상 심볼, 기존 동작 근거, upstream 테스트 영향, 실패 테스트 선행 여부, §0 검토)을 구현 보고에 기록해야 한다(SHALL). 근거 없는 기존 함수 내부 수정은 금지된다(SHALL NOT).

#### Scenario: 선언 없는 High-risk 수정
- **WHEN** Pre-Edit 선언 없이 주문 경로 코드 수정이 보고되면
- **THEN** Manager 리뷰에서 반려되고 선언 후 재작업한다

### Requirement: 완료 보고 조건
Teammate의 완료 보고에는 실행한 테스트 명령과 실제 결과, 변경 파일 요약, DoD 충족 여부, High-risk 영향 여부, upstream 테스트 회귀 여부, 남은 위험이 포함되어야 한다(SHALL). 하나라도 없으면 완료로 취급하지 않는다.

#### Scenario: 테스트 결과 없는 완료 보고
- **WHEN** 테스트 실행 결과가 없는 완료 보고가 제출되면
- **THEN** 완료로 인정되지 않고 검증 후 재보고를 요구한다
