# Branch Test Map: `publishHTTPAPISnapshots`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | tick 마다 본문이 돈다 | `TestTheReattachWakeSurvivesABrokenAggregate` (10ms 주기) | yes (시도 0회) | yes |
| B2 | ctx 종료면 루프가 끝난다 | 같음 (cancel 후 `<-done` 대기) | no | yes |
| B3 | **집계가 실패해도 재부착은 깨어난다** | 같음 | yes (2s 안에 시도 0회) | yes (시도 1회 + 부착) |
| B4 | 같은 digest 는 다시 싣지 않는다 | 기존 `internal/httpapi` stream 테스트 | no | yes |
| B5 | `Publish` 실패는 루프를 끝낸다 | 기존 | no | yes |

뮤테이션 M31(tick 첫 줄 삭제) → **CAUGHT**. B3 의 방어가 실제로 그 줄에 얹혀 있다.
