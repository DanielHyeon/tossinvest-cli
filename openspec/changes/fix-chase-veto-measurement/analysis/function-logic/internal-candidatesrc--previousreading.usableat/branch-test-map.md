# Branch Test Map: `previousReading.usableAt`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 소스의 첫 읽기 — 기억이 nil이므로 답하지 않는다 | `TestASourcesFirstReadingHasNoAnswerAboutNewEntrants` · `TestOneSourceServingTwoMarketsDoesNotAnswerAboutTheWrongList` | yes (3-상태 이전에는 이 함수가 없었고 첫 읽기가 `false`로 접혔다) | yes |
| (꼬리) age < 0 | 시계가 뒤로 간 뒤의 읽기 | `TestAClockThatStepsBackwardsDoesNotMakeTheMemoryFresh` | yes | yes |
| (꼬리) age >= TTL | 장애·경계값 4종 | `TestTheMemoryOfAReadingBeforeAnOutageIsNotAnAnswer` · `TestTheMemoryExpiresAtTheStalenessTTLAndNotBefore`(one tick / full ladder / TTL-1ns / TTL / 4×TTL) | yes (F1이 실행 probe로 확인) | yes |

**2026-07-28 부분 해소**: `TestAMemoryWithNoInstantOnEitherSideIsNotAnAnswer`가 zero
instant 네 조합과 정상 대조군을 직접 몬다. `p.at.IsZero() || at.IsZero()` 두 절을 지우면
**한 case**가 실패한다 — 양쪽 다 zero인 경우다. 나머지 둘은 뺄셈이 각각 큰 음수와 56년을
내므로 부호 검사와 TTL이 먼저 거부한다. 그 사실을 테스트 주석이 말한다.

이 arm은 여전히 생산 경로에서 만들 수 없다 — `rememberRead`가 `symbols`와 `at`을 같은
리터럴에서 쓴다. 방어적으로 남기는 이유는 zero value가 "무한히 신선함"으로 읽히면 안 되기
때문이고, 도달 불가한 가드는 지워도 아무것도 실패하지 않는 가드라는 것이 이제 테스트로
막혀 있다.
