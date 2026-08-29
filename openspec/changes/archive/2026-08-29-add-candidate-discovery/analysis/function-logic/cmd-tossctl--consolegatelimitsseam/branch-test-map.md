# Branch Test Map: `consoleGateLimitsSeam`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | config 경로 미해석 → seam nil, 콘솔은 그래도 기동 | `TestTheConsoleComesUpWithoutTheLimitsSeam` + 개요의 미배선 렌더 | yes | yes |

정상 경로(구체 타입 반환·메서드 1개)는 `TestTheConsoleIsHandedTheLimitsAsNumbersAndNoWayToWriteThem`가 소유한다.
