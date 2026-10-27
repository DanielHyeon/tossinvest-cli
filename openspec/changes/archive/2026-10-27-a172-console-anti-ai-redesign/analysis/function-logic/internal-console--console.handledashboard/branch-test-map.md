# Branch Test Map: `Console.handleDashboard`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | run 없음이면 `Run`을 비우고, 있으면 기존 스냅숏·2초 새로고침 조건을 유지함 | `TestStartupReadinessAppearsOnThreeScreensWithoutControls` 및 `go test ./internal/console/ -count=1` 기존 회귀군 | RED 2.1: 준비도 section 부재로 실패, 기존 분기 변경 0 | GREEN 3.1: 두 시험 통과 |
