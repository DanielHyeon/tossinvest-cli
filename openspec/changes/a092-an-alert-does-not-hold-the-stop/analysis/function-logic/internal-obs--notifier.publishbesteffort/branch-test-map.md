# Branch Test Map: `Notifier.publishBestEffort`

Source: `internal/obs/notifier.go` (155-167). AST 기준 분기 2 / 이탈 1 / defers 0.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:156` publisher 없음 → 무동작·무로그 | `TestTheTransitionLogLineIsCountable` (`mode_test.go:88`) | no | yes |
| B2 | `:159` publish 오류 → warn 한 줄, 사건화 안 함 | `TestObservationFailureIsLoggedEvenWhenUndeliverable` (`measurement_test.go:166`) · `TestOrdinaryAlertsAreBestEffort` (`obs_test.go:527`) | no | yes |
| — | 성공(최대 10s 체류) | `TestCriticalWithoutAJournalIsLoudRatherThanSilent` (`obs_test.go:600`)은 발송을 단언하지만 **체류를 단언하지 않는다** | no | yes |

> **18라운드 B-P3이 바꾼 두 칸 — 그리고 리뷰가 절반만 맞았다.**
>
> B1이 이름 붙이던 `TestObservationFailureAlertNeverReachesTheGateOrTheMode`
> (`measurement_test.go:85`)는 `&failingPublisher{fail: true}`를 넣는다 — publisher가
> **nil이 아니므로 B1을 타지 않는다**. 성공 이탈이 이름 붙이던 `TestNtfyPublishesToTheTopic`
> (`obs_test.go:208`)은 `(*Ntfy).Publish`를 직접 부른다 — `publishBestEffort`를 **거치지
> 않는다**. 두 지목 다 거짓이었고 리뷰가 그것을 맞게 잡았다.
>
> **그러나 "그러므로 미검증"은 틀렸다.** 두 분기 다 다른 테스트가 덮는다:
> `EventOperatingMode`는 critical 등급표에 있고(`event.go:290`), `mode_test.go:88`의
> `Notifier`는 `Log`만 채운다. 그래서 `AnnounceOperatingMode`(`mode.go:57`) →
> `Notify` → `notifyCritical` B1(`Journal == nil` `:171`) → `publishBestEffort`
> `:180`으로 **nil publisher를 들고 B1에 도달한다**. 성공 이탈은 `obs_test.go:600`의
> `&failingPublisher{}`가 `fail`을 켜지 않아 **발송이 성공하고**, 같은 테스트가
> `callCount() == 1`을 단언한다.
>
> 지목만 지웠으면 없는 결함 둘을 만들어 냈을 것이다. **정정의 단위는 이름이 아니라 값이다.**

## 필요한 RED

| # | Scenario | 기대 |
|---|---|---|
| R10 | normal 이벤트, publish가 블록 | 호출자가 즉시 반환한다 |
| R17 | 유계 큐가 가득 참 | 이벤트가 버려지고 **그 사실이 기록된다** — 조용한 유실 금지 |

R17이 없으면 이 change가 B1의 "조용한 구멍"을 큐로 옮겨 심는 것에 불과하다.
