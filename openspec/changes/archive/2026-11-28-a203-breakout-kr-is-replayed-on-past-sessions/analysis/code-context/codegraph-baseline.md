# CodeGraph baseline — a203 (tasks 1.2)

- 기준 commit: `base-commit.txt` = `8877d7e01` (브랜치 `docs/a203-depth-receipt`). CodeGraph **1.6.0**.
- 이 worktree(`~/tossos-work/a203`)에는 CodeGraph 색인이 없음(`codegraph affected` → 「CodeGraph not initialized」). 그래서 질의는 primary
  체크아웃(`main` `e2921fc00`)의 색인으로 함 — 편집 대상 기존 파일(`main.go`)은 base 와 main 에서 같음, 새 파일은 색인에 없음(새 leaf).

| 질의 | 결과 |
|---|---|
| `codegraph affected tools/a140-breakout-signal-shadow/main.go --filter '*_test.go'`(primary) | 2 파일: `tools/a140-breakout-signal-shadow/a166_budget_test.go` · `internal/breakoutlane/a112_sizing_oracle_test.go` |
| 기본 판별식(필터 없음) | 쓰지 않음 — 이 저장소에서 Go 시험을 못 찾음(.claude/CLAUDE.md 하네스 주석) |
| a166 census 대상 | `internal/accountbudgetruntime/testdata/a166-wiring/shapes.json` 의 `tools/a140-breakout-signal-shadow/main.go\|\|dispatch`(본문 고정) · `sites.json`(배선 자리) — 둘 다 이 change 가 갱신(M2) |

편집 표면: `dispatch`(기존 함수 1개, FLM 번들 `analysis/function-logic/tools-a140-breakout-signal-shadow--dispatch/`) + 새 파일 5개
(`a203_replay.go` · `a203_cache.go` · `a203_universe.go` · `a203_quote_proxy.go` · `a203_verdict.go`) + 그 시험 + census 시험 데이터 2 파일.
