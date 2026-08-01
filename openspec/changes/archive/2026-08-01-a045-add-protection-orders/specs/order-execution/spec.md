## ADDED Requirements

### Requirement: 조건주문 mutation은 durable execution contract를 따른다
조건주문 create/replace/cancel은 canonical body, serializer version, client order identity, mutation attempt와 broker identifier를 submit 전에 영속해야 한다 (SHALL).

#### Scenario: create 응답 유실
- **WHEN** broker가 create를 처리했으나 응답이 유실된다
- **THEN** attempt는 IN_DOUBT로 남고 attested idempotency/reconciliation 절차 외 재제출을 금지한다

#### Scenario: replace가 새 ID를 반환
- **WHEN** 보호 정정이 새 conditional ID를 만든다
- **THEN** old/new identifier와 유효성 전환을 한 saga generation에 기록한다
