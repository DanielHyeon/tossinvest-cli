## ADDED Requirements

### Requirement: 보호주문 불일치는 신규 진입을 차단하고 수렴한다
reconciliation은 broker conditional orders와 local protection saga를 비교하고 missing, duplicate, orphan, quantity mismatch를 typed discrepancy로 격리해야 한다 (SHALL).

#### Scenario: broker orphan
- **WHEN** 계좌에 local saga가 모르는 활성 조건주문이 있다
- **THEN** 자동 취소하거나 귀속을 추정하지 않고 RECONCILE로 격리하며 신규 진입을 차단한다

#### Scenario: flatten
- **WHEN** 운영자가 포지션 전량 flatten을 승인한다
- **THEN** 2초 안에 관련 보호주문의 terminal cancel과 broker sellable quantity를 확인한 경우에만 기존 reduce-only liquidation을 실행한다

#### Scenario: flatten cancel이 모호함
- **WHEN** cancel 응답이 유실되거나 trigger 경합으로 2초 안에 terminal 상태를 확인할 수 없다
- **THEN** saga를 `IN_DOUBT`로 격리하고 신규 exposure를 차단하며 최우선 reconcile과 사람 emergency action을 요구하고 oversell 가능한 blind liquidation을 제출하지 않는다
