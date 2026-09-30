# multi-horizon-risk-buckets — a126 delta

> 사용량 정의(a066 D5 — "Position 에 귀속된 filled 노출")에 수명주기를 세운다. 떠나는 사건은 owner release
> receipt(Q1), 감소는 저장값을 바꾸지 않는 파생(Q2), 부분 종결은 감소하지 않음(Q4), latch 는 감소가 풀지 않음(Q3).
> 근거와 AST 인용은 design.md.

## ADDED Requirements

### Requirement: 체결 노출은 owner release receipt 와 함께 bucket 을 떠난다

owner release receipt(Position generation CLOSED·수량 0, clean reconciliation, 공식 broker 수량 0 관측, 미해소 fill·claim·saga·latch 부재가 증명된 뒤 한 번 기록되는 불변 영수증)가 커밋된 owner 에 귀속된 모든 monetary reservation 의 filled exposure 는 그 receipt 가 커밋된 시점부터 모든 적용 bucket 의 사용량 합과 한도 비교 모집단에서 함께 빠져야 하며(SHALL), 그 감소는 저장된 filled amount 를 바꾸지 않고 receipt 존재에서 파생되어야 한다(SHALL).

사용량 감소는 그 owner 의 receipt 외의 어떤 사건 — Position projection 단독 CLOSED, 부분 매도, reconciliation 단독,
운영자 요청, 시각 경과 — 으로도 일어나서는 안 된다(MUST NOT). receipt 뒤 그 owner scope 에 scope latch(late fill 의
ORPHAN_FILL 등)가 하나라도 기록되면 그 owner 의 filled exposure 는 다시 사용량 합과 한도 모집단에 계상되어야 한다(SHALL).
떠난 reservation 이 HELD 상태이거나 HELD 잔량을 가지면 사용량은 판독 불가로 거절되어야 한다(SHALL).

#### Scenario: 종결된 owner 의 release 가 사용량을 줄인다

- **WHEN** 한 owner 의 filled exposure 가 공유 sector bucket 에 쌓인 뒤 그 owner 의 release receipt 가 커밋된다
- **THEN** 그 sector bucket 을 포함한 모든 적용 bucket 의 사용량 합에서 그 owner 의 filled 몫이 빠지고, 같은 한도에서
  다음 진입의 여유가 그만큼 회복되며, 저장된 reservation 의 filled amount 는 바뀌지 않는다

#### Scenario: receipt 없는 종결은 사용량을 줄이지 않는다

- **WHEN** Position generation 이 CLOSED·수량 0 이지만 release 선결 조건(clean reconciliation, 공식 broker 0 관측, 미해소 claim 부재 등) 중 하나가 없어 receipt 가 없다
- **THEN** 그 owner 의 filled exposure 는 사용량 합과 한도 모집단에 그대로 남는다

#### Scenario: 부분 매도는 사용량을 줄이지 않는다

- **WHEN** 한 owner 의 보유 수량 일부가 매도 체결로 줄었지만 owner 는 해제되지 않았다
- **THEN** 그 owner 의 filled exposure 는 전액 사용량에 남는다

#### Scenario: 해제된 owner 의 late fill 은 떠남을 되돌린다

- **WHEN** release receipt 뒤 그 owner 의 주문에 late BUY 체결이 관측되어 그 owner scope 에 scope latch 가 기록된다
- **THEN** 그 owner 의 filled exposure 는 다시 모든 적용 bucket 의 사용량 합과 한도 모집단에 계상된다

#### Scenario: 손상된 떠난 행

- **WHEN** receipt 가 있는 owner 의 reservation 이 HELD 상태이거나 HELD 잔량을 가진다
- **THEN** 사용량은 판독 불가 오류로 거절되고 진입은 fail closed 한다

### Requirement: 사용량 감소는 overage latch 를 풀지 않는다

사용량 감소는 어떤 RISK_OVERAGE 또는 UNKNOWN_ACTUAL_RISK latch 도 해제하거나 latch 집계에서 제외해서는 안 되며(MUST NOT), 감소 뒤의 공유 bucket 합은 다음 fill transaction 의 overage 재계산과 다음 admission 의 여유 계산에만 반영되어야 한다(SHALL).

latch 가 남은 owner 는 release 될 수 없으므로 순서는 운영자 latch 해제 → owner release(=떠남) → 이후 fill 의 overage
재계산으로 정해져야 한다(SHALL).

#### Scenario: 떠남은 다른 owner 의 latch 를 풀지 않는다

- **WHEN** owner A 의 사용량 때문에 공유 bucket 에서 owner B 에 RISK_OVERAGE 가 latch 되어 있고 그 뒤 owner A 가 release 된다
- **THEN** owner B 의 RISK_OVERAGE 는 그대로 남아 신규 exposure 를 차단하고, 운영자 해제 뒤 B 의 다음 fill 은 A 가 빠진 합으로 overage 를 판정한다

#### Scenario: latch 된 owner 는 떠나지 못한다

- **WHEN** RISK_OVERAGE 또는 UNKNOWN_ACTUAL_RISK 가 latch 된 owner 에 release 가 시도된다
- **THEN** release 는 거절되고 그 owner 의 filled exposure 는 사용량에 남는다
