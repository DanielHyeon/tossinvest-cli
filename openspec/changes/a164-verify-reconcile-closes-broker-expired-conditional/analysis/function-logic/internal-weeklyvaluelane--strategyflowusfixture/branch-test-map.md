# Branch Test Map: StrategyflowUSFixture

- Source: `internal/weeklyvaluelane/strategyflow_testseam.go` (45-47). AST branches 0.
- 분기 없음. GREEN: 2026-10-28 `go test -tags tossos_testseams -count=1 -run 'TestProductionEvaluateUsesRealRouterAndTheSixLaneProductionFixtures|TestStrategyflowFollowOnFixtureBuildsNativePrefixThroughOrdinalSeven' ./internal/strategyflow/ ./internal/weeklyvaluelane/` → 두 패키지 ok.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 위임 한 줄(통과 경로) | `TestProductionEvaluateUsesRealRouterAndTheSixLaneProductionFixtures/US-weekly`(`internal/strategyflow`) | no(사후 작성) | yes |
