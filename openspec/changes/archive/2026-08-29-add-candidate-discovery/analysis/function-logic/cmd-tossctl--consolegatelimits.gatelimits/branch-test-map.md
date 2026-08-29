# Branch Test Map: `consoleGateLimits.GateLimits`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | config 읽기 실패 → 에러가 그대로 개요까지 간다(0이 아니다) | 개요의 미측정 패널 + `TestTheLimitsReadIsBounded`(경계가 존재함을 소스로 고정) | yes (`context.Background()` 직접 사용 시 FAIL) | yes |

정상 경로는 `TestTheConsoleIsHandedTheLimitsAsNumbersAndNoWayToWriteThem`(메서드 1개·필드 배선)가 소유한다.
