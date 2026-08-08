# official-account-scope Specification

## Purpose
공식 클라이언트의 기본 계좌 선택과 account sequence cache의 일관성 계약을 정의한다.

## Requirements

### Requirement: 공식 클라이언트의 기본 계좌 scope
공식 클라이언트가 명시적 account sequence 없이 계좌 scope가 필요한 요청을 수행할 때 기본 계좌는 공식 계좌 목록의 첫 레코드여야 한다(SHALL). 첫 레코드의 sequence가 양수가 아니면 요청을 거부해야 하며(SHALL), 뒤 레코드를 암묵적으로 선택해서는 안 된다(SHALL NOT). `WithAccountSeq`로 구성된 양수 sequence는 명시적 선택이며 목록 조회 결과로 덮어써서는 안 된다(SHALL NOT). `WithAccountSeq(0)`은 기존 unresolved sentinel이며 명시적 선택이 아니다. 음수 sequence는 유효한 선택이 아니며 account header로 전송해서는 안 된다(SHALL NOT).

#### Scenario: 첫 레코드의 양수 sequence
- **WHEN** 명시적 sequence가 없는 클라이언트가 첫 레코드 sequence 7인 계좌 목록을 성공적으로 읽으면
- **THEN** 다음 account-scoped 요청은 `X-Tossinvest-Account: 7`을 사용한다

#### Scenario: 명시적 sequence 우선
- **WHEN** `WithAccountSeq(99)`로 구성된 클라이언트가 첫 레코드 sequence 7인 목록을 읽으면
- **THEN** 다음 account-scoped 요청은 `X-Tossinvest-Account: 99`를 사용한다

#### Scenario: 음수 명시값 거부
- **WHEN** 클라이언트가 `WithAccountSeq(-1)`로 구성되면
- **THEN** account-scoped 요청은 계좌 header를 보내지 않고 유효하지 않은 sequence 오류로 거부된다

#### Scenario: 유효하지 않은 첫 sequence
- **WHEN** 첫 계좌 레코드의 sequence가 0 또는 음수이면
- **THEN** cache는 unresolved로 남고 account-scoped 요청은 잘못된 header를 보내지 않고 거부된다

### Requirement: 계좌 sequence cache의 완결성과 직렬화
공식 클라이언트는 완전히 성공한 계좌 목록 응답만 sequence cache에 반영해야 한다(SHALL). 오류·취소·빈 목록은 cache를 변경해서는 안 된다(SHALL NOT). public `Accounts`와 lazy account resolution의 첫 discovery는 같은 mutex로 직렬화되어야 하고(SHALL), recursive lock 없이 완료되어야 한다(SHALL). unresolved 클라이언트의 동시 account-scoped 첫 사용은 성공한 첫 discovery를 공유해 `/api/v1/accounts`를 한 번만 호출해야 한다(SHALL). public `Accounts` 요청이 먼저 진행 중이면 뒤따른 unresolved account-scoped 요청은 그 discovery의 sequence를 재사용해야 한다(SHALL). sequence가 이미 양수로 선택된 account-scoped 요청은 진행 중인 public 목록 I/O를 기다려서는 안 된다(SHALL NOT). 암묵적으로 선택된 sequence와 이후 public 목록의 첫 레코드가 달라지거나 유효하지 않으면 목록 결과를 오류로 거부하고 기존 선택을 조용히 변경해서는 안 된다(SHALL NOT). lazy sequence-only discovery 뒤에 시작된 public 목록 조회까지 한 번으로 합치는 account-list cache는 이 요구사항의 범위가 아니다. 실패·취소된 discovery가 mutex를 영구 점유해서는 안 된다(SHALL NOT).

#### Scenario: 동시 첫 사용
- **WHEN** 여러 account-scoped 요청이 unresolved 클라이언트에서 동시에 시작되면
- **THEN** 계좌 목록은 한 번만 성공적으로 조회되고 모든 요청은 같은 sequence를 사용한다

#### Scenario: public 조회와 scoped 요청의 경합
- **WHEN** public `Accounts` HTTP 요청이 진행 중인 동안 account-scoped 요청이 시작되면
- **THEN** scoped 요청은 discovery 완료 후 같은 sequence를 사용하며 두 번째 계좌 목록 요청을 만들지 않는다

#### Scenario: 선택 완료 후 public 조회와 scoped 요청의 경합
- **WHEN** 양수 sequence가 이미 선택된 상태에서 public `Accounts` HTTP 요청이 진행 중이면
- **THEN** account-scoped 요청은 public 목록 I/O를 기다리지 않고 선택된 sequence를 사용한다

#### Scenario: 암묵적 선택 이후 계좌 목록 drift
- **WHEN** 암묵적으로 sequence 7을 선택한 뒤 public 계좌 목록의 첫 sequence가 8 또는 유효하지 않은 값으로 바뀌면
- **THEN** public 목록 조회는 identity drift 오류로 거부되고 선택된 sequence 7은 유지된다

#### Scenario: 실패 후 mutex 해제
- **WHEN** 계좌 목록 요청이 오류 또는 context 취소로 끝난 뒤 다음 요청이 시작되면
- **THEN** 다음 요청은 deadlock 없이 새 discovery를 수행할 수 있고 실패 응답의 sequence는 사용되지 않는다
