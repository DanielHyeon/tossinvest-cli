# position-ledger Specification (delta)

## MODIFIED Requirements

### Requirement: Position 투영과 단일 권위
Position은 journal의 체결 이벤트와 조정 이벤트의 **투영**이며(SHALL) 직접 변이 API를 노출하지 않는다(SHALL NOT). 투영은 심볼·시장 단위 집계, position-instance 식별자(CLOSED 후 재진입은 새 인스턴스), 평균단가, 수량을 산출하고, 엔진 발 포지션은 진입 결정 참조(`entry_decision_id`)를 가진다(SHALL — 설정 후 인스턴스 수명 동안 변경·NULL화되지 않는다(SHALL NOT); 외부·수동 취득 포지션은 NULL로 남는다). 외부 취득 포지션의 편입은 `positions.adoption_id`(additive v7, `position_adoptions` 참조)로 기록되며 **set-once**다(SHALL — 전용 tx API로만 기입하고 그 외 UPDATE의 언급은 정적 스캔이 거부한다). exit 정책 대상 자격은 `entry_decision_id 또는 adoption_id`가 설정된 포지션이며(SHALL — 자격 판정은 단일 술어 함수로 모은다), adoption_id 부여는 수량·평균단가를 변경하지 않는다(SHALL NOT). 방향은 intent side에서 재도출한다(SHALL — 이 change의 범위에서 모든 로컬 체결은 intent가 있는 주문에서 온다; 발동 주문의 방향 출처는 보호주문 도입 change가 정의한다). 포지션 진실은 하나이므로 reconciliation의 로컬 상태는 이 투영을 소비한다(SHALL — reconciliation delta). decimal 문자열 산술을 사용한다(SHALL NOT float 누적 — 편입가·원가는 브로커 원문 decimal 문자열을 보존하며 float 경유를 금지한다).

편입 포지션의 lineage 형태(SHALL 명시): `ADOPTION → POSITION → EXIT_EVENT …`이며 intent·MutationAttempt·Fill arm이 비는 것이 정상이다. 조정 이벤트로 편입 포지션의 수량이 0이 되면 exit_state는 completed 처리되고 trade_outcome이 ADJUSTMENT_CLOSED provenance로 동결된다(SHALL — 고아 exit_state 금지).

#### Scenario: 체결 반영으로만 포지션 변화
- **WHEN** 체결 delta가 반영되면
- **THEN** Position 투영이 갱신되고, 그 외 어떤 코드 경로도 Position 수량을 직접 쓰지 못한다

#### Scenario: 청산 후 재진입
- **WHEN** CLOSED된 심볼에 새 진입이 발생하면
- **THEN** 새 position-instance 식별자와 새 진입 결정 참조로 시작하고 이전 인스턴스 기록은 보존된다

#### Scenario: 편입 기록의 set-once
- **WHEN** adoption_id가 이미 설정된 포지션에 재기입이 시도되면
- **THEN** 전용 API가 거부하고 기존 참조는 변하지 않는다

#### Scenario: 외부 매도로 수량 0
- **WHEN** 편입 포지션의 수량이 조정 이벤트로 0이 되면
- **THEN** exit_state가 completed 처리되고 trade_outcome이 ADJUSTMENT_CLOSED provenance로 동결된다
