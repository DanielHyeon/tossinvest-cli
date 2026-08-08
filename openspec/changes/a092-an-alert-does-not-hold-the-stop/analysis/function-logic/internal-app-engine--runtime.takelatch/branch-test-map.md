# Branch Test Map: `Runtime.takeLatch`

a092는 이 함수를 편집하지 않는다. 표는 **인용한 분기가 실재함**을 AST로 고정하기 위한 것이고,
RED/GREEN은 해당 없음이다.

| Branch | 위치 | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:431` | 같은 루프가 두 번째로 임계를 넘어도 `false`를 돌려 승격을 막는다 | 기존 커버리지 — a092 편집 없음 | n/a | n/a |
| — | `:434-435` | 첫 초과는 래치를 세우고 `true` | 기존 커버리지 — a092 편집 없음 | n/a | n/a |
