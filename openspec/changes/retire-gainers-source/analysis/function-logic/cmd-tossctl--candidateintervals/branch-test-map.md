# Branch Test Map: `candidateIntervals`

분기가 없는 함수이므로 행은 happy path 하나다(ast.json `branches`: null).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 반환된 키 집합이 `candidatesrc.Panel`이 만들 수 있는 id 집합에 **포함**된다. 규칙은 등호가 아니다 — 패널에 있고 일정에 없는 원천은 위반이 아니라 `unconfiguredFloor`가 적용되는 설계된 경우다 | `TestNoIntervalNamesASourceNoPanelBuilds` · `TestTheScheduleGuardIsContainmentAndNotEquality` | **yes** — 1.1만 적용하고 1.2 적용 전에 `candidateIntervals sets a cadence for [official_rankings_top_gainers] and no market's panel builds those sources`로 실패. 변이(항목만 되살리기)에서 같은 실패 재현 | yes |
