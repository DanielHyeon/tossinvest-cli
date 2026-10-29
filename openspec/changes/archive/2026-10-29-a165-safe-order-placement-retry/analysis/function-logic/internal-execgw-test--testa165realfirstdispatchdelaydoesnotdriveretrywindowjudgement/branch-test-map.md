# Branch Test Map: `TestA165RealFirstDispatchDelayDoesNotDriveRetryWindowJudgement`

| Branch | Coordinate | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `if at 94:2` | 실제 25ms firstDelay | 함수 자신 | IN_DOUBT 기대 실패 | CONFIRMED |
| B2 | `if at 97:2` | AP + retry-after 1ms | 함수 자신 | 호출 1 기대 실패 | 호출 2 |
| B3 | `if at 101:2` | retry ACK 정착 | 함수 자신 | IN_DOUBT 기대 실패 | journal CONFIRMED |
| B4 | `if at 104:2` | Fake Clock elapsed | 함수 자신 | wall 지연 반영 | 1ms |

대조군: `TestA165WindowStartsBeforeFirstDispatch`가 주입 Clock을 7초 전진시킨 뒤 다음 1초 경계를 `stop=window`로 잡음.
