# a157 CodeGraphContext supporting context

- 도구: `codegraphcontext analyze callers` (falkordb, `~/.codegraphcontext/.env`). 원문 `raw/cgc-callers-*.txt`. 각 질의 1 초 안에 끝남.
- **색인 범위가 이 체크아웃이 아니다.** 결과 행은 `/mnt/D/Axipient/workspace/TossOS-worktrees/archive-batch1/…`(옛 worktree) ·
  `/mnt/D/project/axipient/TossOS/…`(옛 경로) · `/tmp/…` 사본들이고, `/mnt/D/Axipient/workspace/TossOS/internal/…` 행은 0 이다.
  줄 번호도 옛 소스의 것(`strategy_first_leg_admission.go:69`, `strategy_entry_supervisor.go:283/371/579`)이다.
- 그래서 이 결과는 **누락 후보 탐색용 advisory** 로만 쓴다. 새 색인(`codegraphcontext index .`)은 전역 DB 를 바꾸고 다른 세션과 공유되므로 돌리지 않았다
  (`not-applicable` — 현재 HEAD 의 hard evidence 는 CodeGraph 1.6.0 + HEAD grep 이 진다).

| 질의 | CGC 후보 | 현재 HEAD 대조 |
|---|---|---|
| `PrecheckQFinalCampaignFirstLeg` 호출자 | `admit`(여러 사본) + 옛 사본의 `TestQFinalCampaignFirstLeg*` | 생산 호출자 `admit` :105 하나 — 일치(`raw/grep-precheck-first-leg-nontest.txt`). 새 후보 없음 |
| `NewStrategyEntrySupervisor` 호출자 | `NewPairedStrategyEntryProductionAssembly`(두 자리) · `NewRefreshingPairedStrategyEntrySupervisor` | 일치 — 생산 조립 셋(`(*Context)` 둘 + 패키지 함수 :648). a157 은 `(*Context).NewPairedStrategyEntryProductionAssembly` 만 편집(spec MODIFIED 가 두 `Context` 생성자를 명시) |
