# Branch Test Map: `strategyRuntimeAttachment.attempt`

인용 전용 — 기존 a109 시험 · a115 시험이 덮는다(편집 없음).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 실패(!live) — 침묵, **붙어 있는 자리 불변**(live→sentinel 격하 금지) | `TestAFailedAttemptDoesNotClobberTheCurrentScreen` · `TestTheConsoleStrategyScreenReattachesAfterTheEngineRestarts` | no — 무편집(a109 원장이 변이로 잰다) | yes |
| B2 | 빈 자리(nil)에 sentinel 승격(a109 G1), seat++ — 영구 NOT_CONFIGURED 탈출로 | `TestAnEmptySeatTakesTheUnavailableSentinel` | no — 무편집(a109 원장이 변이로 잰다) | yes |
| B3 | 부착 전이 1회 로그(announce). 밀려난 값은 잠금 밖에서 Close(G5) | `TestTheAttachmentReportsOnlyTransitions` · `TestTheReplacedReaderIsClosed` | no — 무편집(a109 원장이 변이로 잰다) | yes |
