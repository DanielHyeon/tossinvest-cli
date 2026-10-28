# Branch Test Map: `a165BoundaryBroker.CancelPendingOrder`

| Branch | Coordinate | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `if at 42:2` | 첫 cancel | `TestA165RealFirstDispatchDelayDoesNotDriveRetryWindowJudgement` | D5 이전 wall 판정 기대 | AP 409 반환 |
| B2 | `if at 43:3` | `firstDelay` 25ms | 같은 시험 | 실제 지연이 창을 닫음 | Fake Clock은 전진하지 않음 |
