# Branch Test Map: `TestA165RealFirstDispatchDelayDoesNotDriveRetryWindowJudgement`

| 분기 | 입력 | GREEN 결과 |
|---|---|---|
| B1 | 실제 firstDelay 25ms, `Window=10ms` | wall delay로 IN_DOUBT 전환하지 않음 |
| B2 | AP 409 뒤 retry-after 1ms | broker 호출 2회 |
| B3 | retry ACK | journal CONFIRMED |
| B4 | Fake Clock | 1ms만 경과 |

대조군: `TestA165WindowStartsBeforeFirstDispatch`가 주입 Clock을 7초 전진시킨 뒤 다음 1초 경계를 `stop=window`로 잡음.
