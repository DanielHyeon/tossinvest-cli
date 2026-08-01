## ADDED Requirements

### Requirement: exit 기준선은 하나의 권위 snapshot으로 계산된다
시스템은 entry, initial stop, current protection, high-water, active rung, next target, next protection, proposal action·ratio·projected quantity를 하나의 immutable decimal snapshot으로 계산해야 한다 (SHALL). 실행과 화면은 이 결과를 별도로 재계산해서는 안 된다 (MUST NOT).
snapshot은 immutable policy ID/version/digest, position generation, observation identity, deterministic snapshot ID와 decision ID를 포함해야 한다 (SHALL).

#### Scenario: 다음 익절과 보호선 계산
- **WHEN** 관리 포지션이 현재 rung과 high-water를 가진 채 평가된다
- **THEN** snapshot은 현재 보호선과 다음 도달 목표·다음 보호선을 함께 반환한다

#### Scenario: 단조 보호선
- **WHEN** 새 후보 보호선이 저장된 보호선보다 낮다
- **THEN** snapshot은 기존 보호선을 유지하고 낮은 값을 적용하지 않는다

### Requirement: 1주는 중간 익절 없이 끝까지 보호한다
whole-share 보유 수량이 정확히 1주일 때 시스템은 중간 부분익절 주문을 생성해서는 안 되며 (MUST NOT), rung과 보호선 승격은 계속 적용해야 한다 (SHALL). 최종 전량익절 또는 보호선 breach는 정확히 1주 전량 청산을 제안해야 한다 (SHALL).

#### Scenario: 1주 중간 목표 도달
- **WHEN** 1주 포지션이 partial ratio 0.25인 중간 rung에 도달한다
- **THEN** 주문 proposal은 없고 active rung과 보호선만 상승하며 수량은 1주로 남는다

#### Scenario: 1주 최종 목표 도달
- **WHEN** 같은 포지션이 final take-full rung에 도달한다
- **THEN** 정확히 1주 전량익절 proposal을 생성한다

#### Scenario: 1주 보호선 이탈
- **WHEN** 같은 포지션의 현재가가 active protection 아래로 내려간다
- **THEN** 중간 익절 이력과 무관하게 정확히 1주 전량보호 proposal을 생성한다

### Requirement: 0수량 주문은 존재할 수 없다
부분익절 projected quantity가 최소 주문 단위보다 작으면 시스템은 state-only transition으로 처리하고 0수량 intent·reservation·broker request를 만들어서는 안 된다 (MUST NOT).

#### Scenario: 내림 결과 0주
- **WHEN** 잔량과 partial ratio의 곱을 whole-share로 내림한 결과가 0이다
- **THEN** exit state는 승격될 수 있지만 journal mutation attempt와 broker call은 0건이다

### Requirement: 공통 정책 descriptor는 계산 기본값과 설명을 제공한다
시스템은 `COMMON_LADDER_BALANCED`, `COMMON_LADDER_RUNNER`, `COMMON_LADDER_HYBRID_50`의 label, summary, recommended 여부, rung별 target/stop/partial, runner gap, 단위와 1주 projection을 server-authoritative descriptor로 제공해야 한다 (SHALL). UI가 descriptor 밖의 기본 수치를 발명해서는 안 된다 (MUST NOT).

#### Scenario: 미승인 공통 정책
- **WHEN** 운영자가 공통 정책을 아직 승인하지 않았다
- **THEN** effective 상태는 `기존 RATCHET 유지`, 추천 선택은 `COMMON_LADDER_HYBRID_50`으로 서로 구분된다

#### Scenario: 1주 descriptor preview
- **WHEN** descriptor를 1주 포지션에 적용해 preview한다
- **THEN** 모든 intermediate partial은 `매도 0주 · 보호선 승격`, 최종 take-full과 protection breach는 `1주 전량`으로 설명된다

### Requirement: 설정 descriptor는 transport-neutral이고 유한한 선택만 허용한다
시스템은 UI나 HTTP 타입에 의존하지 않는 field key/type, control kind, finite stable option ID, default/effective state, apply timing, safety direction과 provenance 계약을 제공해야 한다 (SHALL). 동일 policy ID/version의 canonical digest가 달라지면 해당 descriptor와 snapshot을 거부해야 한다 (SHALL).

#### Scenario: policy identity 충돌
- **WHEN** 같은 policy ID/version으로 서로 다른 rung digest가 로드된다
- **THEN** registry와 exit evaluation은 fail-closed하고 기존 identity의 의미를 덮어쓰지 않는다

#### Scenario: 자유 입력 없는 정책 선택
- **WHEN** transport가 공통 정책 descriptor를 렌더링한다
- **THEN** finite stable option ID만 선택할 수 있고 transport가 별도 수치나 기본값을 발명하지 않는다
