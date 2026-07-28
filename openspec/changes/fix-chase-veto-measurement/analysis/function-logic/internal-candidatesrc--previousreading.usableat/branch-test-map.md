# Branch Test Map: `previousReading.usableAt`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 소스의 첫 읽기 — 기억이 nil이므로 답하지 않는다 | `TestASourcesFirstReadingHasNoAnswerAboutNewEntrants` · `TestOneSourceServingTwoMarketsDoesNotAnswerAboutTheWrongList` | yes (3-상태 이전에는 이 함수가 없었고 첫 읽기가 `false`로 접혔다) | yes |
| (꼬리) age < 0 | 시계가 뒤로 간 뒤의 읽기 | `TestAClockThatStepsBackwardsDoesNotMakeTheMemoryFresh` | yes | yes |
| (꼬리) age >= TTL | 장애·경계값 4종 | `TestTheMemoryOfAReadingBeforeAnOutageIsNotAnAnswer` · `TestTheMemoryExpiresAtTheStalenessTTLAndNotBefore`(one tick / full ladder / TTL-1ns / TTL / 4×TTL) | yes (F1이 실행 probe로 확인) | yes |

`p.at.IsZero()` 단독 arm(집합은 있는데 instant가 없는 기억)은 생산 경로에서 만들 수 없다 —
`rememberRead`가 `symbols`와 `at`을 같은 리터럴에서 쓴다. 전용 테스트는 **없고**, 그것이
정직한 기록이다. 방어적으로 남기는 이유는 zero value가 "무한히 신선함"으로 읽히면 안 되기
때문이다.
