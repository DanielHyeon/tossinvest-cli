## ADDED Requirements

### Requirement: exit_events 기존 행은 운영자 정비 명령으로만 바뀐다
exit_events 는 append-only 판정 이력이며 엔진 · 콘솔 · API 의 자동 경로가 기존 행을 고치거나 지워서는 안 된다(SHALL NOT).
예외는 사람이 실행하는 `mutating: true` 정비 명령 하나다(SHALL). 그 명령은 엔진 잠금을 잡지 못하면 거절하고(SHALL),
쓰기 전에 저널 백업을 만들며(SHALL), 대상 행 수와 전후 크기를 audit 에 남겨야 한다(SHALL).

#### Scenario: 엔진이 돌고 있다
- **WHEN** 엔진이 저널 잠금을 쥔 채 정비 명령을 적용하려 한다
- **THEN** 명령은 아무것도 쓰지 않고 거절한다

#### Scenario: 적용 없이 실행
- **WHEN** 운영자가 `--apply` 없이 정비 명령을 실행한다
- **THEN** 명령은 대상 집계만 출력하고 저널 데이터 · 백업 · audit 어느 것도 쓰지 않는다

#### Scenario: 적용 전 백업
- **WHEN** 운영자가 엔진을 멈추고 정비 명령을 적용한다
- **THEN** 어떤 행도 바뀌기 전에 백업이 만들어지고 실행 결과가 audit 에 남는다

### Requirement: 정비는 닫힌 포지션의 미연결 행의 판정 증거만 통째로 비운다
정비 대상은 CLOSED 이고 exit 상태가 completed 인 포지션의 행 가운데 브로커 주문 식별자를 가진 attempt 에 연결되지 않은 행뿐이며, 반복 거절 압축 행과 그 앵커 행은 대상에서 뺀다(SHALL).
그 행의 판정 증거 열(평가 tuple 과 세 snapshot JSON) 전부를 한 트랜잭션에서 함께 NULL 로 바꿔야 하며(SHALL — 일부만 비우면 리더가
그 행을 손상으로 읽는다), 행 자체와 action · proposed_intent_id · 가격 · 시각은 남긴다(SHALL). 비워진 행은 legacy 행처럼 unknown
근거로 읽혀야 하고 0 이나 재계산 값으로 채워져서는 안 된다(MUST NOT).

#### Scenario: 열린 포지션과 연결 행은 남는다
- **WHEN** 운영자가 정비 명령을 적용한다
- **THEN** 열린 포지션의 행 · 브로커 주문에 연결된 행 · 반복 거절 압축 행과 그 앵커의 증거 열은 바이트 그대로이고, 닫힌 포지션의 나머지 미연결 행만 증거 열 전부가 NULL 이 된다

#### Scenario: 정비 뒤 리더
- **WHEN** 정비된 저널을 엔진 · 콘솔 리더가 읽는다
- **THEN** 읽기 오류는 0 건이고 비워진 행은 legacy unknown 근거로, orders 연결 결과는 정비 전과 같게 나온다

#### Scenario: 두 번째 실행
- **WHEN** 같은 조건으로 정비 명령을 다시 적용한다
- **THEN** 바뀌는 행은 0 이고 압축 행과 앵커는 여전히 남는다
