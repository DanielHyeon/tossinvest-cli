# Branch Test Map: `consoleGateSwitchSeam`

편집되지 않았으므로 분기 하나가 그대로다. 표는 그 사실의 기록이다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 프로필이 해석되면 seam을, 아니면 nil을 준다 | 기존 console 배선 테스트 | no | pass |

## 추가 시나리오 (분기 밖의 계약)

| 계약 | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| 무편집 | hunk가 이 함수 밖에서 시작한다 (삭제 0줄) | `git diff --unified=0` | no | pass |
