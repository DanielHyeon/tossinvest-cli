## ADDED Requirements

### Requirement: 브로커 보호 설정과 상태는 한 카테고리에서 설명된다
콘솔은 a050의 `exit-protection` 카테고리에서 capability, activation, current effective trigger, protected quantity, broker identifier, updated-at과 reconcile reason을 표시해야 한다 (SHALL). 각 항목은 label, 쉬운 설명, 기본값, desired/effective 값, 적용 시점과 provenance를 가져야 한다 (SHALL).

#### Scenario: attestation 미완료
- **WHEN** 현재 시장의 조건주문 capability가 확인되지 않았다
- **THEN** 화면은 `OFF / 지원 확인 전 사용 불가`를 기본·실효 상태로 표시하고 주문 유형이나 trigger 기본값을 임의 생성하지 않는다

#### Scenario: capability 확인 완료
- **WHEN** SINGLE+MARKET capability가 attested됐다
- **THEN** 지원 유형과 근거를 표시하되 activation 기본값은 OFF이고 운영자 승인 전 자동 활성화하지 않는다

#### Scenario: 활성 보호주문
- **WHEN** protection saga가 ACTIVE다
- **THEN** effective trigger, 수량, broker ID, 마지막 확인 시각과 다음 reconcile 설명을 한 section에서 읽을 수 있다

### Requirement: 보호 약화는 강화와 구분해 확인한다
콘솔은 trigger 하향, 보호 해제 또는 수량 감소처럼 보호를 약화하는 변경을 분류하고 before/after, 영향 포지션·수량, 보호 공백 가능성과 적용 시점을 표시한 뒤 3초 지연 확인을 요구해야 한다 (SHALL).

#### Scenario: trigger 하향 요청
- **WHEN** 운영자가 ACTIVE trigger보다 낮은 값을 preview한다
- **THEN** 위험 확인을 표시하더라도 domain contract에 따라 apply는 거부되고 현재 보호를 유지한다

#### Scenario: 보호 강화
- **WHEN** 새 trigger가 더 안전한 방향이고 모든 capability가 유효하다
- **THEN** 변경 범위와 적용 시점을 표시하되 약화 전용 경고 문구를 사용하지 않는다
