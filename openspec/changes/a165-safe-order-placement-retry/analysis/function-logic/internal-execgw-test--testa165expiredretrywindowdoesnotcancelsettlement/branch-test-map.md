# Branch Test Map: `TestA165ExpiredRetryWindowDoesNotCancelSettlement`

| 분기 | 시나리오 | D5 4판 GREEN |
|---|---|---|
| B1 | 첫 실제 왕복 25ms, Fake Clock 0, retry-after 1ms | retry CONFIRMED |
| B2 | retry 호출 | `broker.calls == 2` |
| B3 | 정착 기록 | `StateConfirmed` |

양성 대조군: Fake Clock을 실제로 Window 경계까지 전진시키는 `TestA165WindowStartsBeforeFirstDispatch`는 여전히 `stop=window`/첫 AMBIGUOUS를 고정함.
