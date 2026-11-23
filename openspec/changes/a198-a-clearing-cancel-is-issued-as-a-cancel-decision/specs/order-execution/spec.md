## ADDED Requirements

### Requirement: 취소가 이미 끝난 주문을 이유로 거절되면 확정 거절로 분류한다

취소 요청의 응답 본문 code가 `already-filled`·`already-canceled`·`already-rejected`·`already-modified` 이면 그 취소 attempt는 IN_DOUBT로 가지 않고 확정 거절로 종결해야 한다(SHALL — 그 주문은 더 이상 취소할 수 있는 상태가 아니어서 취소가 성립하지 않았고, 주문의 운명은 체결 감지가 기록한다). 판정은 「브로커가 이름을 준 거절은 code가 분류한다」와 같은 규칙을 따라야 한다(SHALL). 이 분류는 취소에만 적용해야 하며(SHALL), 발주·정정 응답에 적용해서는 안 된다(SHALL NOT).

#### Scenario: 이미 체결된 주문의 취소
- **WHEN** 취소 요청이 409와 `error.code` `already-filled` 로 거절되면
- **THEN** 그 취소 attempt는 확정 거절로 종결되고 그 종목의 다른 mutation을 막지 않는다

#### Scenario: 같은 문자열이 발주 응답에 있다
- **WHEN** 발주 요청의 거절 응답이 `already-filled` code 를 실으면
- **THEN** 이 분류는 적용되지 않고 종전 분류를 따른다

#### Scenario: 처리 중인 동일 요청
- **WHEN** 취소 요청이 409 `request-in-progress` 로 거절되면
- **THEN** 종전대로 모호로 분류된다

#### Scenario: 두 code 자리의 값이 다르다
- **WHEN** 취소 거절 본문의 최상위 `code` 가 `already-filled` 이고 `error.code` 가 다른 값이면
- **THEN** 모호로 강제된다
