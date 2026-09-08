# Branch Test Map: `TestEveryRouteGoesThroughTheSessionGate`

분기 셋 그대로다. a075는 B3의 임계값을 27에서 30으로 올렸고 조건의 모양은
바꾸지 않았다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 등록된 모든 라우트를 순회한다 | 자기 자신 | no | pass |
| B2 | session0 없는 라우트를 잡는다 | 자기 자신 (기존 회귀 이력) | no | pass |
| B3 | 표가 30개 미만이면 추출기가 멈춘 것 | 자기 자신 | no | pass |

## 추가 시나리오 (분기 밖의 계약)

| 계약 | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| 하한 | a075의 세 라우트가 하한에 반영됐다 | 자기 자신 | no | pass |
| 실제 수 | 등록 수 49 > 하한 30 — 격차는 선행 change에서 온 것이다 | `grep -c mux.HandleFunc` | no | 기록됨(I4) |
