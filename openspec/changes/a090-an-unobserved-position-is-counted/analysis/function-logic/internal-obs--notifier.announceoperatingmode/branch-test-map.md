# Branch Test Map: `Notifier.AnnounceOperatingMode`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:49` nil Notifier 무동작 | `TestAnnouncingWithoutANotifierIsSafe` (`internal/obs/mode_test.go:252`) | no | yes |

> a090 은 이 함수를 편집하지 않는다(FLM 「a092 이후 바뀐 것」). 설계 단계의 B2(완화 방향 표기)는 a092 가 `operatingModeEvent` 로 옮겼다 —
> `TestARelaxationIsAnnouncedAsARelaxation` (`mode_test.go:135`)이 그 경로를 이 함수를 거쳐 잰다.
