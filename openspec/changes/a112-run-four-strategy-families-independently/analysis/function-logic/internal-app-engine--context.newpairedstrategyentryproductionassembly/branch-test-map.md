# Branch Test Map: `NewPairedStrategyEntryProductionAssembly`

- Source: `internal/app/engine/strategy_entry_supervisor.go` (283-370); file SHA-256 `9d97e59cf36590ade76b3e4804134b9ee66b3af1f8efa409e2f41222d5579b4f`. AST branch positions are authoritative.

- Measurement regime (8.7.2 편집 뒤): 몸통 진입 count. engine tagged suite 바이너리(`-coverpkg=./internal/app/engine,./internal/strategyrouter`, -trimpath 없이)를 `systemd-run … MemoryMax=16G` 안에서 실행, 스위트 PASS; 전체 시험 509 개를 하나씩 돈 per-test 프로파일(하네스 `analysis/harness/a872_pertest_cover.sh` · `a872_attribute.py`). 모든 행에서 시험별 합 == 스위트(ATTRIBUTION MISMATCH 0).

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 284:2 | arm not entered (engine tagged suite, post-edit); no per-test profile entered it |
| B2 | if | 290:2 | arm not entered (engine tagged suite, post-edit); no per-test profile entered it |
| B3 | if | 301:2 | arm not entered (engine tagged suite, post-edit); no per-test profile entered it |
| B4 | if | 311:2 | arm not entered (engine tagged suite, post-edit); no per-test profile entered it |
| B5 | if | 332:2 | arm entered 1x (engine tagged suite, post-edit); `TestTheMarketThatLeadsAWaveAlwaysPublishesIt` |
| B6 | if | 337:3 | arm not entered (engine tagged suite, post-edit); no per-test profile entered it |
| B7 | range | 354:2 | arm entered 2x (engine tagged suite, post-edit); `TestTheMarketThatLeadsAWaveAlwaysPublishesIt` |
| B8 | if | 359:2 | arm not entered (engine tagged suite, post-edit); no per-test profile entered it |
| B9 | if | 366:2 | arm not entered (engine tagged suite, post-edit); no per-test profile entered it |

8.7.2 가 더한 것: `if clk != nil { dispatchCycle.now = clk.Now }` 한 분기(Codex 재리뷰 P2 — nil 인터페이스의 메서드 값은 그 자리에서 panic). 그 뒤 분기들은 한 칸씩 밀렸다. 대입 한 줄은 `TestTheProductionAssemblyGivesTheDispatchCycleTheRealClock` 이 구조로 못 박는다(조립 전체를 행동으로 세우려면 원장·게이트웨이·권한 여덟이 필요하다) — 변이 M23 CAUGHT.

옛 표(8.8.2 까지의 조건-평가 regime 과 편집 전 좌표)는 이 번들의 git 이력에 있다 — 이 파일은 현재 소스만 적는다.

A row states what was measured, not what is intended. An arm recorded as not entered is a coverage gap, not a pass.
