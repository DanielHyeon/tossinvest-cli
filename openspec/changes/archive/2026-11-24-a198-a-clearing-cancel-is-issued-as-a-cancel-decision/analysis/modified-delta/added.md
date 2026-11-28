## ADDED Requirements

### Requirement: 취소가 이미 끝난 주문을 이유로 거절되면 확정 거절로 분류한다

엔진 게이트웨이의 취소가 브로커 응답 본문 code `already-filled`·`already-canceled`·`already-rejected` 로 거절되면 그 취소 attempt는 IN_DOUBT로 가지 않고 확정 거절로 종결해야 한다(SHALL — 그 주문은 브로커에서 이미 종결돼 후계 주문이 없고 취소가 성립하지 않았으며, 주문의 운명은 체결 감지가 기록한다). 판정은 「브로커가 이름을 준 거절은 code가 분류한다」와 같은 규칙을 따라야 한다(SHALL). `already-modified` 와 CLI flatten 의 취소, 발주·정정 응답에는 적용해서는 안 된다(SHALL NOT — 정정의 후계 주문은 원장이 모른다).

#### Scenario: 이미 체결된 주문의 취소
- **WHEN** 엔진 게이트웨이의 취소가 409와 `error.code` `already-filled` 로 거절되면
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

#### Scenario: 정정된 주문의 취소
- **WHEN** 엔진 게이트웨이의 취소가 409 `already-modified` 로 거절되면
- **THEN** 종전대로 결과 불명(IN_DOUBT)으로 기록된다

#### Scenario: flatten 의 취소가 이미 취소된 주문을 만난다
- **WHEN** CLI flatten 의 취소가 409 `already-canceled` 로 거절되면
- **THEN** 이 분류는 적용되지 않고 flatten 의 종전 처리를 따른다

### Requirement: 세션은 IN_DOUBT 로 남은 취소를 대상 주문의 종결 기록으로 닫는다

엔진이 낸 취소가 세션 중 결과 불명(IN_DOUBT)으로 남으면, 대상 주문의 소유 발주·정정 범위(계좌·시장·거래일·종목·방향, 소유 intent가 하나일 때만)에서 길 치우기와 같은 정의의 종결 체결 기록이 fail-closed 없이 있을 때 브로커를 부르지 않고 닫혀야 한다(SHALL): 대상이 취소로 끝났으면 접수 확정(대상 번호), 체결·거부로 끝났으면 확정 거절. 판정은 기동 해소의 같은 상태 대응과 같아야 하고(SHALL), 판독과 종결은 한 트랜잭션이어야 한다(SHALL). 대상이 살아 있거나 후계로 넘어갔거나 판정할 수 없으면 attempt를 바꾸거나 park해서는 안 된다(SHALL NOT).

#### Scenario: 대상이 취소로 끝났다
- **WHEN** 세션 중 IN_DOUBT 인 취소의 대상 주문에 취소됨 종결 기록이 원장에 들어오면
- **THEN** 다음 스윕에 그 취소 attempt는 대상 번호로 접수 확정 종결되고 그 종목의 변경 차단이 풀린다

#### Scenario: 대상이 체결로 끝났다
- **WHEN** 세션 중 IN_DOUBT 인 취소의 대상 주문에 전량 체결 종결 기록이 원장에 들어오면
- **THEN** 그 취소 attempt는 확정 거절로 종결된다

#### Scenario: 대상이 살아 있다
- **WHEN** 세션 중 IN_DOUBT 인 취소의 대상 주문이 원장에서 미체결·부분 체결·판정 불능이면
- **THEN** 그 취소 attempt는 IN_DOUBT 로 남는다

#### Scenario: 범위가 다른 같은 번호
- **WHEN** 대상 번호와 같은 번호의 종결 기록이 다른 계좌·거래일·종목·방향의 것뿐이면
- **THEN** 그 취소 attempt는 바뀌지 않는다

#### Scenario: flatten 이 낸 취소
- **WHEN** 수동 flatten 이 낸 취소가 IN_DOUBT 로 남고 대상 주문의 종결 기록이 원장에 있으면
- **THEN** 엔진의 실행자는 그 취소를 닫지 않는다

#### Scenario: 브로커 요청 없음
- **WHEN** 실행자가 대상 주문의 종결 기록으로 취소 attempt 를 닫으면
- **THEN** 그 취소의 판독·종결은 브로커에 아무 요청도 보내지 않는다(같은 스윕의 IN_DOUBT 발주 해소는 종전대로 조회한다)
