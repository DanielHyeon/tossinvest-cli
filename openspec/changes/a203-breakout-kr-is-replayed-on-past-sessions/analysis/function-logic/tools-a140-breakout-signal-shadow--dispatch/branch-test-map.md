# Branch Test Map: `dispatch`

편집 전(base `8877d7e01`) 갈래. 편집 뒤 표는 구현 커밋에서 갱신함.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 인자 없음 → 오류 | (편집 뒤 추가) | no | no |
| B2 | switch 59 `args[0]` — 이름별 갈래 선택 | 아래 B3~B5 가 함께 잼 | no | no |
| B3 | `run` | `TestA166Breakout*`(runCommand 직접 — dispatch 경유 아님) | no | no |
| B4 | `report` | (편집 뒤 추가) | no | no |
| B5 | 모르는 이름 → 오류 | (편집 뒤 추가) | no | no |
