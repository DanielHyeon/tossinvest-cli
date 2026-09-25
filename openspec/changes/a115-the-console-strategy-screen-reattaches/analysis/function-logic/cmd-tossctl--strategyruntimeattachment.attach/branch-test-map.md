# Branch Test Map: `strategyRuntimeAttachment.attach`

인용 전용 — 기존 a109 시험 · a115 시험이 덮는다(편집 없음).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | happy path — 부팅 해석 결과를 자리에 앉힌다(attached=live, failed=!live, seat++) | `TestTheConsoleBootKeepsAnUnreachableEndpointUnreachable` · `TestTheConsoleStrategyScreenRecoversWhenTheEngineStartsLater` · `TestTheDaemonAttachesWhenTheEngineComesUpLater` | no — 무편집 | yes |
