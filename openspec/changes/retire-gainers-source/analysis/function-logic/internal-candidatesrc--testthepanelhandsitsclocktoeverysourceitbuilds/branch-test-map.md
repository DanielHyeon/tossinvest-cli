# Branch Test Map: `TestThePanelHandsItsClockToEverySourceItBuilds`

이 함수 자체가 테스트이므로 "Test" 열은 그 분기를 실행시키는 구성 또는 변이를 적는다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 패널이 만든 원천의 id를 모은다 | 자기 자신 (KR, reader 모두 non-nil) | no | yes |
| B2 | 패널이 비면 아래 루프가 공허하게 통과하는 것을 막는다 | 자기 자신. `official`·`wts` 모두 nil일 때 도달하며, 같은 성질을 `TestThePanelWithoutAnOfficialClientIsNotSilentlyWTSOnly`가 별도로 지킨다 | no | yes (도달하지 않는 경로로서 통과) |
| B3 | id 집합이 기대 집합과 **정확히** 같다. 1.1 적용 직후·2.1 적용 전에는 이 자리의 기존 길이 단언이 깨졌고, 새 집합 단언은 같은 변이에서 다시 깨진다 | 변이: `Panel`의 리터럴에 `RankingTopGainers` 복원 | **yes** — 기존 단언: `clock_wiring_test.go:65: the KR panel has 3 sources, want 4 (three rankings and the popularity list)`. 새 단언: `the KR panel builds [... official_rankings_top_gainers ...], want [official_rankings_trading_value official_rankings_trading_volume wts_popular]` | yes |
| B4 | 원천마다 subtest가 돈다 | 자기 자신 | no | yes (subtest 3개) |
| B5 | 첫 `Read`가 실패하면 멈춘다 | 자기 자신 (fake는 실패하지 않는다) | no | yes |
| B6 | 두 번째 `Read`가 실패하면 멈춘다 | 자기 자신 | no | yes |
| B7 | 두 번째 reading에 행이 없으면 멈춘다 | 자기 자신 | no | yes |
| B8 | 한 tick 뒤 모든 행이 known으로 답한다 | 자기 자신 | no | yes |
| B9 | 세 번째 `Read`가 실패하면 멈춘다 | 자기 자신 | no | yes |
| B10 | TTL을 넘긴 뒤에는 어떤 행도 known으로 답하지 않는다 — 주입된 시계를 쓴다는 증거. 이 change는 이 분기를 건드리지 않았다 | 자기 자신 | no | yes |
