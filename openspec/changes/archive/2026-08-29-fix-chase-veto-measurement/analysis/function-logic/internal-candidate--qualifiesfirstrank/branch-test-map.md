# Branch Test Map: `qualifiesFirstRank`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 분기 없음 — 두 절의 논리곱. 세 경우(둘 다 참 / 첫 절 거짓 / 둘째 절 거짓)를 아래 세 테스트가 각각 몬다 | `TestAQualifiedReadingTakesTheFirstSightingFromAPanelEarlierUnqualifiedOne` · `TestWhenNoReadingInTheTickCanQualifyThePositionIsHeld` · `TestAReadingThatNeverRecordedItsRequestIsHeldToo` | yes | yes |

세 경우를 풀어 적는다. 한 행에 묶은 이유는 AST에 분기가 없기 때문이고, 그것이 이 술어의
요점이기도 하다 — 조건이 하나로 모여 있어야 두 호출자가 갈라질 수 없다.

| 경우 | 테스트 | RED 확인 (2026-07-28 실행) |
|---|---|---|
| 두 사실이 모두 있다 → 자격 있음 | `TestAQualifiedReadingTakesTheFirstSightingFromAPanelEarlierUnqualifiedOne` · `TestATruncatedReadingReachesTheVerdictAsTruncated`(절단은 자격을 잃지 않는다는 대조군) | `Collect`의 선점 규칙을 옛 형태로 되돌리면 2건 실패 |
| 직전 읽기가 없다(`NewlyListed unknown`) → 자격 없음 | `TestASessionStartDoesNotStampThePanelAsSeenLate` · `TestWhenNoReadingInTheTickCanQualifyThePositionIsHeld` | `FirstRanksHeld++`를 지우면 4건 실패 |
| 요청 행 수가 기록되지 않았다(`RankRequested == 0`) → 자격 없음 | `TestAReadingThatNeverRecordedItsRequestIsHeldToo` | `&& r.RankRequested > 0`을 지우면 1건 실패 |

**정직한 커버리지 기록**: 커버되지 않는 조합은 없다. 둘 다 거짓인 경우는 별도로 몰지
않는다 — 논리곱이므로 첫 절이 거짓이면 결과가 결정되고, 그 경우는 두 번째 행이 덮는다.
