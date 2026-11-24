# a200 증거 조정(1단계)

- 기준: base `884be2a74`(= origin/main `9738a0c32` + 등록 문서). CodeGraph **1.6.0**, 이 worktree 에서 `codegraph init` 로 새 색인(`codegraph-callers.txt`).
- CodeGraphContext: **실행 안 함** — 이 1단계는 기존 함수 호출 자리 · 소비자 열거가 목적이고, 그 답은 CodeGraph callers + grep + AST 로 냄. 구현 단계에서 필요하면 돌린다.

| 질의 | CodeGraph | 현재 HEAD(grep · AST) | 결론 |
|---|---|---|---|
| `ProjectPosition` · `ApplyExitFill` 의 생산 호출자 | 시험 함수만 | gateway.go:202 · 204 함수 값 배선 + `runApplyHooks` Project@284 · Exit@298 | CodeGraph 는 함수 값 배선을 호출로 안 셈(사전 증거와 같은 누락). 자리 수는 grep+AST 로 |
| `runApplyHooks` 생산 호출자 | RecordFill · backfillConfirmedStrategyFillTx | fills.go:551 · strategy_dispatch_runtime.go:1348(BUY 만) | 일치. 사전 증거의 「1」 을 「2(둘째는 BUY 만)」 로 정정 |
| `enterReconcileScopeInTx` 호출자 | 5 함수 | 사전 증거 5 자리와 일치 | 일치 |
| `adoptionEpoch` · `computeTradeOutcome` 생산 호출자 | 3 · 4 | grep 같은 수 | 일치 |
| `ExitProposalForIntent` 호출자 | (질의 안 함) | execgw/a166_replay_origin.go:47 하나 | grep |

불일치는 「CodeGraph 가 함수 값 배선을 못 봄」 하나이고, 현재 HEAD 로 해소함(편집 차단 사유 아님).
