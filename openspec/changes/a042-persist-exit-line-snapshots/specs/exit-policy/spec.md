## ADDED Requirements

### Requirement: 복구된 기준선은 낮아질 수 없다
exit recovery는 저장 snapshot과 현재 재계산 후보 중 더 안전한 기준만 채택하고 baseline과 high-water를 낮춰서는 안 된다 (MUST NOT).
선택은 검증된 coherent snapshot 단위로 수행해야 하며 (SHALL), 서로 다른 policy version/rung/target의
field별 최댓값을 조합해 합성 snapshot을 만들어서는 안 된다 (MUST NOT).
protection/high-water 비교는 같은 policy digest 안에서만 허용하고 (SHALL), rung과 next target/protection은 선택된 immutable policy에서 파생해야 한다 (SHALL). policy digest가 다르거나 안전한 후보 하나를 결정할 수 없으면 해당 포지션을 격리해야 한다 (SHALL).

#### Scenario: 재시작 후 낮은 현재가
- **WHEN** 재시작 시 재계산 후보가 저장된 active protection보다 낮다
- **THEN** 저장된 protection을 유지하고 포지션을 더 낮은 손절선으로 재해석하지 않는다

#### Scenario: 손상 snapshot
- **WHEN** 한 포지션의 snapshot이 invalid decimal 또는 unknown policy version을 포함한다
- **THEN** 해당 포지션의 자동 판정은 fail-closed이고 운영 경고를 남기며 다른 포지션의 emergency exit는 계속한다

#### Scenario: 더 높은 값이 서로 다른 후보에 분산됨
- **WHEN** saved 후보는 더 높은 protection을, recomputed 후보는 더 높은 rung을 가지지만 tuple 조합이 정책상 불가능하다
- **THEN** 검증된 후보 하나를 통째로 선택하거나 해당 포지션을 격리하고 field별 max snapshot을 만들지 않는다
