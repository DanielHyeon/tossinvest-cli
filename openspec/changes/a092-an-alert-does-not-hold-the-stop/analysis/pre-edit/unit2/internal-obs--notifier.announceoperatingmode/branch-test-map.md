# Branch Test Map: `Notifier.AnnounceOperatingMode`

- Source: `internal/obs/mode.go` (49-74); **편집 전** 측정 — `analysis/harness/branch_coverage.py`를 연결 워크트리(`8c390aa6`)에서
  `./internal/obs` 시험 81개에 하나씩 돌린 커버 프로필(`analysis/harness/coverage-pre-mode.json`).

| Branch | AST anchor | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 50:2 | `n == nil` → nil | `TestAnnouncingWithoutANotifierIsSafe` | 편집 전(기준선) | 블록 50.14-52.3을 시험 1개가 실행, PASS |
| B2 | if at 54:2 | 완화 방향 | `TestARelaxationIsAnnouncedAsARelaxation` · `TestTheTransitionLogLineIsCountable` | 편집 전(기준선) | 블록 54.90-56.3을 시험 2개가 실행, PASS |
