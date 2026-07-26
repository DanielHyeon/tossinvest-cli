# position-ledger Specification (delta)

## MODIFIED Requirements

### Requirement: Position 투영과 단일 권위
Position은 journal의 체결 이벤트와 조정 이벤트의 **투영**이며(SHALL) 직접 변이 API를 노출하지 않는다(SHALL NOT). 투영은 심볼·시장 단위 집계, position-instance 식별자(CLOSED 후 재진입은 새 인스턴스), 평균단가, 수량을 산출하고, 엔진 발 포지션은 진입 결정 참조(`entry_decision_id`)를 가진다(SHALL). 외부·수동 취득 포지션의 결정 참조는 편입 전 NULL이며 exit 정책 대상이 아니다(SHALL); **ADOPTION 결정이 영속되면 결정 참조가 그 결정을 가리키고 exit 정책 대상이 된다**(SHALL — exit-policy "외부 취득 포지션의 자동 편입"; 결정 class로 엔진 진입과 편입의 provenance를 구분한다). 방향은 intent side에서 재도출한다(SHALL — 이 change의 범위에서 모든 로컬 체결은 intent가 있는 주문에서 온다; 발동 주문의 방향 출처는 보호주문 도입 change가 정의한다). 포지션 진실은 하나이므로 reconciliation의 로컬 상태는 이 투영을 소비한다(SHALL — reconciliation delta). decimal 문자열 산술을 사용한다(SHALL NOT float 누적).

#### Scenario: 체결 반영으로만 포지션 변화
- **WHEN** 체결 delta가 반영되면
- **THEN** Position 투영이 갱신되고, 그 외 어떤 코드 경로도 Position 수량을 직접 쓰지 못한다

#### Scenario: 청산 후 재진입
- **WHEN** CLOSED된 심볼에 새 진입이 발생하면
- **THEN** 새 position-instance 식별자와 새 진입 결정 참조로 시작하고 이전 인스턴스 기록은 보존된다

#### Scenario: 편입에 의한 결정 참조 부여
- **WHEN** 무결정 포지션에 대한 ADOPTION 결정이 영속되면
- **THEN** 포지션의 결정 참조가 그 결정을 가리키고, 참조 부여는 수량·평균단가를 변경하지 않는다
