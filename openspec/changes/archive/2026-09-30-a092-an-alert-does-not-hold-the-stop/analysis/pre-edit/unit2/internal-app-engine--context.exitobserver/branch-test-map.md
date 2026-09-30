# Branch Test Map: `Context.ExitObserver`

- Source: `internal/app/engine/exitwiring.go` (319-348); **편집 전** 측정 — `branch_coverage.py`, 연결 워크트리 `8c390aa6`, `./internal/app/engine`에서
  `.ExitObserver(`를 부르는 시험 파일 둘(`guardian_production_test.go` · `tracer_test.go`)의 시험 10개(`analysis/harness/coverage-pre-exitobserver.json`).

| Branch | AST anchor | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 320:2 | nil Context | (미실행) | 편집 전(기준선) | 측정 표본의 시험 0개 |
| B2 | if at 323:2 | 게이트 미검증 | `TestTheExitObserverIsUnavailableWithoutAVerifiedGate` | 편집 전(기준선) | 블록 323.28-325.3, PASS |
| B3 | if at 327:2 | Guardian이 발의 불가 | (미실행) | 편집 전(기준선) | 측정 표본의 시험 0개 |
| B4 | if at 338:2 | Names 기본값 | `TestProductionGuardianUsesConfiguredUSDLimitsAndReachesExitObserver` | 편집 전(기준선) | 블록 338.23-340.3, PASS |
| B5 | if at 341:2 | Alerts 기본값 | `TestProductionGuardianUsesConfiguredUSDLimitsAndReachesExitObserver` | 편집 전(기준선) | 블록 341.45-343.3, PASS |
| B6 | if at 344:2 | Floor 기본값 | `TestProductionGuardianUsesConfiguredUSDLimitsAndReachesExitObserver` | 편집 전(기준선) | 블록 344.23-346.3, PASS |
