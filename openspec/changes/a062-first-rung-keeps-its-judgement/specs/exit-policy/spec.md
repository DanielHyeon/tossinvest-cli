# exit-policy — a062 delta

> **ADDED만 쓴다.** 기존 "복구된 기준선은 낮아질 수 없다" 요구사항은 그대로 참이고
> a062는 그 문장이 전제하면서도 적어 두지 않은 것을 명시한다 — **무엇이 "안전한 후보
> 하나를 결정할 수 없다"에 해당하는가**. 기존 SHALL과 충돌하지 않으며 격리 조건을
> 넓히지도 좁히지도 않는다.

## ADDED Requirements

### Requirement: 같은 정책 안의 rung 전진은 격리 사유가 아니다

시스템은 policy identity·position·generation·entry·initial stop이 모두 일치하는 두 복구 후보에 대해, 한쪽만 rung을 활성화했다는 이유로 그 쌍을 비교 불가로 판정해서는 안 된다(SHALL NOT).
rung 미활성은 ladder 단계의 **최하위**이며, rung 미활성에서 rung n으로의 이동은 정상적인 전진으로 비교해야 한다(SHALL). ladder 정책은 rung으로, ratchet 정책은 ratchet level로 단계를 판정하며, 두 후보가 모두 rung 미활성이면 기존 ratchet level 순위 비교를 그대로 사용해야 한다(SHALL).

이 요구사항은 격리 자체를 약화해서는 안 된다(SHALL NOT). policy identity가 다르거나, 해석할 수 없는 ratchet level이거나, protection·high-water·단계가 서로 다른 방향으로 움직여 하나의 검증된 후보를 고를 수 없으면 기존대로 해당 포지션을 격리해야 한다(SHALL). 재계산 후보가 저장 후보보다 뒤쳐진 단계이면 저장 후보를 유지해야 하며 기준선을 낮춰서는 안 된다(MUST NOT).

#### Scenario: 첫 rung 활성화
- **WHEN** 저장된 canonical snapshot이 rung 미활성이고 같은 정책의 재계산 후보가 rung 0을 활성화하며 protection과 high-water도 함께 올랐다
- **THEN** 재계산 후보가 통째로 선택되고 포지션은 격리되지 않으며 판정이 계속된다

#### Scenario: rung을 잃은 재계산
- **WHEN** 저장된 후보가 rung n을 보유하고 재계산 후보가 rung 미활성으로 돌아갔다
- **THEN** 저장된 후보가 유지되고 기준선이 낮아지지 않는다

#### Scenario: 여전히 엇갈리는 축
- **WHEN** 재계산 후보의 rung은 더 높은데 protection은 저장 후보보다 낮다
- **THEN** 하나의 후보를 고르지 않고 해당 포지션을 격리한다

#### Scenario: 정책이 다른 두 후보
- **WHEN** 두 후보의 policy identity·entry·initial stop 중 하나라도 다르다
- **THEN** rung 상태와 무관하게 정체성 불일치로 거부한다

#### Scenario: 해석할 수 없는 ratchet level
- **WHEN** 두 후보가 모두 rung 미활성이고 한쪽의 ratchet level이 알려진 순위에 없다
- **THEN** 기존대로 정체성 불일치로 거부한다

#### Scenario: 판정이 멈추지 않는다
- **WHEN** 관리 중인 ladder 포지션이 첫 익절선을 넘는다
- **THEN** 그 포지션은 이후 관측 주기에서도 계속 판정 대상이며 손절 평가가 유지된다
