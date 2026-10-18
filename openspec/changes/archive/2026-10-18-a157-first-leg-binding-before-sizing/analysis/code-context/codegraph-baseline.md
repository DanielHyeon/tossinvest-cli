# a157 CodeGraph baseline

- 도구: codegraph **1.6.0**. 질의 시점 HEAD `42cc0dff8`(`codegraph sync .` 직후, 2026-10-18). 원문은 `raw/`.
- `internal/` 소스는 base `7b1748481` 이후 바뀌지 않았다(`git diff --stat 7b1748481 HEAD -- internal/` 비어 있음 — 뒤의 두 커밋은 a157 문서뿐).
  그래서 base 질의(`raw/callers-*` · `raw/impact-*` 의 첫 여섯 파일)와 이번 질의는 같은 소스를 본다.

| 심볼 | 생산 호출자 (CodeGraph) | a157 영향 |
|---|---|---|
| `execgw.RiskGuardian.PrecheckQFinalCampaignFirstLeg` (riskguardian_first_leg.go:49) | 인터페이스 선언 `strategy_first_leg_admission.go:65` 하나로 표시 — 실제 호출 `admit` :105 는 목록에 없음(reconciliation 1) | **편집** — 결속 판정 일곱을 `PrecheckQFinalEntry` 앞으로, 일일 손실 우선 |
| `execgw.RiskGuardian.PrecheckQFinalEntry` (riskguardian_qfinal.go:89) | `PrecheckQFinalCampaignFirstLeg` :49 · `IssueQFinalEntry` :322 | 무편집(파일 sha 로 확인) |
| `execgw.RiskGuardian.IssuePrecheckedQFinalCampaignFirstLeg` (riskguardian_first_leg.go:117) | `admit` :156(인터페이스 경유) | 무편집(같은 파일 — 함수 본문 sha 로 확인) |
| `engine.strategyFirstLegAdmissionBridge.admit` (strategy_first_leg_admission.go:86) | 이름 충돌 — `strategyworker` 의 `admit`(bounded.go) 과 섞이고 `strategyDispatchCycle.dispatch` 누락(`raw/callers-admit.txt`, `raw/impact-admit.txt` 53 심볼 대부분이 다른 `admit`) | **편집** — 일일 손실 두 조건을 판별 함수 호출로(분기 결과 동일) |
| `engine.Context.NewPairedStrategyEntryProductionAssembly` (strategy_entry_supervisor.go:325) | `refreshPairedStrategyEntryProductionAssembly` :611 하나(생산) — 같은 이름의 패키지 함수(:648)와 `NewDormantStrategyEntrySupervisor` 가 섞임 | **편집** — 394 행 감독자 옵션에 알림 둘 + nil 가드 게이트 |
| `engine.Context.strategyDailyLossNotice` · `strategyEntrySkipNotice` | `NewRefreshingPairedStrategyEntrySupervisor` :414 하나 + 시험 각 1 | 무편집 — 호출자가 하나 늘어남 |
| `engine.StrategyEntrySupervisor.blockEntryOnCentralIntegrity` | `runMarket` :926 하나 | 무편집 — 조립 감독자의 `entry` 를 읽게 됨(D5) |

## 영향 시험

- `codegraph affected <riskguardian_first_leg.go · strategy_first_leg_admission.go · strategy_entry_supervisor.go> --filter '*_test.go'` → `raw/affected-filter.txt`(**1,153** 시험 파일).
- 같은 질의를 기본 판별식으로 돌리면 **1** 파일(`auth-helper/tests/test_cli.py`)이다(`raw/affected-default.txt`) — Go `_test.go` 를 못 찾는 알려진 결함(.claude/CLAUDE.md 하네스 주석). 기본값으로 「영향 시험 없음」이라 결론내지 않는다.
- 실제로 돌릴 대상 패키지: `internal/execgw` · `internal/app/engine`(둘 다 `tossos_testseams` 포함) · `cmd/tossctl`(감독자 조립 dormant 시험) · `internal/risk` · `internal/riskbucket` · `internal/journal`.
