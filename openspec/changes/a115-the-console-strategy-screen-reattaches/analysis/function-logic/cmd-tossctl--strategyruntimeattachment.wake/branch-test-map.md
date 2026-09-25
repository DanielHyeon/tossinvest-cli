# Branch Test Map: `strategyRuntimeAttachment.wake`

인용 전용 — 기존 a109 시험 · a115 시험이 덮는다(편집 없음).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | single-flight(trying)·rate limit(tooSoon)·수명 종료(ctx.Err) 중 하나면 early return; 아니면 trying=true·lastTry=now·`go attempt()` | `TestTheAttemptIsSingleFlight` · `TestTheAttemptIsRateLimited` · `TestTheConsoleStrategyPumpStopsWithTheConsole` | no — 무편집(a109 원장이 변이로 잰다) | yes |
