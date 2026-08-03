# Change: add position campaign and leg core

## Why

기존 position 투영은 체결 결과와 exit 상태를 권위 있게 보존하지만, 한 전략 결정이 여러 번의 계획·부분체결·재시작을 거쳐 하나의 포지션을 형성하는 과정을 표현하지 못한다. 8:4:2, 2:4:8, 주 1회 최대 7회처럼 서로 다른 진입 방식을 안전하게 병행하려면 전략 상수와 분리된 campaign/leg 수명주기와 재구성 가능한 lineage가 먼저 필요하다.

## What Changes

- 하나의 account/symbol/market 포지션 생성을 소유하는 `PositionCampaign`과 순서가 있는 `CampaignLeg` 도메인을 추가한다. 첫 fill 전에는 현재 Position generation/version을 compare-and-swap해 prospective generation token을 유일하게 예약하고, 첫 fill의 실제 successor generation에 원자 결합한다.
- 코어는 leg 순서·계획·제출 참조·누적 체결·취소/종결 상태만 표현하고, 8:4:2·2:4:8·7-leg 같은 전략별 비율과 cadence는 포함하지 않는다.
- 결정적 idempotency key, 완전한 Campaign/Leg 상태표와 broker order identity별 cumulative fill watermark를 사용해 retry, partial fill, amend/replacement, duplicate observation 및 process restart에서도 계획과 체결 적용이 중복되지 않게 한다.
- replaced/cancelled predecessor에서 late positive fill delta가 도착해도 fill과 Position 적용을 버리거나 leg cap에 맞춰 자르지 않는다. immutable broker-order watermark와 Position을 같은 transaction에서 exactly once 전진하고 replacement remaining 및 leg aggregate를 다시 계산한 뒤 campaign을 `RECONCILE`로 격리해 신규 exposure만 차단한다.
- EXIT FIRST를 전이 우선순위로 고정하고, campaign의 effective stop은 이미 저장된 보호선보다 불리한 방향으로 후퇴할 수 없게 한다.
- production entry caller를 연결하기 전에 append-only journal evidence로 campaign과 leg를 offline 재구성하고 불일치를 보고하는 read/replay 경로를 제공한다.

## Capabilities

### New Capabilities

- `position-campaigns`: 전략 중립 campaign/leg 상태기계, 멱등 전이, EXIT FIRST, 단조 보호선 및 offline reconstruction 계약

### Modified Capabilities

- `position-ledger`: Position 변화의 기존 단일 권위를 유지하면서 campaign/leg와 decision, intent, fill, position generation 사이의 명시적 lineage 및 additive persistence를 확장

## Impact

- journal에는 additive-nullable campaign/leg/prospective-generation/order-watermark 테이블·참조와 재시작 투영이 추가되며 기존 position 수량 권위는 바뀌지 않는다.
- fill apply/reconstruction 경계와 lineage 조회가 영향을 받으므로 crash, 중복, 부분체결, replacement 뒤 late fill, cap 초과 및 schema-too-new 회귀 테스트가 필요하다.
- 신규 진입 caller, live 주문, lane 활성화는 이 change 범위 밖이다. 기존 stop·emergency exit·reconciliation·fill detection의 즉시성과 우선순위를 유지한다.
