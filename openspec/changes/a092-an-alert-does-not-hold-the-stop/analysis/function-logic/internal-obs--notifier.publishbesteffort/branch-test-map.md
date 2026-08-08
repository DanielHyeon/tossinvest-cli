# Branch Test Map: `Notifier.publishBestEffort`

Source: `internal/obs/notifier.go` (138-150). AST 기준 분기 2 / 이탈 1 / defers 0.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:139` publisher 없음 → 무동작·무로그 | `TestObservationFailureAlertNeverReachesTheGateOrTheMode` (`measurement_test.go:85`) | no | yes |
| B2 | `:142` publish 오류 → warn 한 줄, 사건화 안 함 | `TestObservationFailureIsLoggedEvenWhenUndeliverable` (`measurement_test.go:166`) · `TestOrdinaryAlertsAreBestEffort` (`obs_test.go:527`) | no | yes |
| — | 성공(최대 10s 체류) | `TestNtfyPublishesToTheTopic` (`obs_test.go:208`)은 발송을 단언하지만 **체류를 단언하지 않는다** | no | yes |

## 필요한 RED

| # | Scenario | 기대 |
|---|---|---|
| R10 | normal 이벤트, publish가 블록 | 호출자가 즉시 반환한다 |
| R17 | 유계 큐가 가득 참 | 이벤트가 버려지고 **그 사실이 기록된다** — 조용한 유실 금지 |

R17이 없으면 이 change가 B1의 "조용한 구멍"을 큐로 옮겨 심는 것에 불과하다.
