# Evidence reconciliation — a203

| 항목 | CodeGraph 1.6.0(primary `e2921fc00`) | 현재 HEAD 직접 읽기 · AST | 결론 |
|---|---|---|---|
| `dispatch` 의 시험 | `a166_budget_test.go` · `a112_sizing_oracle_test.go` | `a166_budget_test.go` 는 `runCommand` 를 직접 부름(dispatch 경유 아님). `a112_sizing_oracle_test.go` 에는 a140 · tools 참조가 없음(`grep` 0) — 이름 해소 우연으로 보임 | 불일치 — HEAD 가 정본. 둘 다 돌려 PASS(`../verify/`). dispatch 갈래 시험은 새 `TestA203DispatchRoutesReplay` |
| a166 census | — | `shapes.json` 의 `main.go\|\|dispatch` 본문 고정 · `sites.json` 은 `tools/` 의 `accountbudgetruntime.Open` · `.NewClient` · `.RoundTrip` 자리를 셈 | 이 change 가 dispatch 1건 갱신 + 새 자리 4건 추가(M2 범위). `go test ./internal/accountbudgetruntime/` PASS |
| `PollClosedBars` 과거 폴 | — | `ValidityAt` 은 시계 역행(폴 < 달력 받은 시각)만 거절 → 재생은 달력 「받은 시각」을 폴 시각으로 둠. `SealBarSeries` 는 `ingested_at ≤ 컷오프` 라 컷오프는 지금(저장소 시계) | claims.md 미검증 항목 → `TestA203PastPollLoadsAWholeSession` 으로 확인(390 봉, 마감 뒤 봉 없으면 389) |
| `NewEvidenceSnapshot` 평가 시각 | — | `validStructuralEvidenceInput` 에 평가 시각 · 봉 나이 조건 없음 | 「마지막 봉 + 2 분 이상」 거절 0 — 같은 시험의 평가 오류 0 |
