# CodeGraph hard evidence — a197 (tasks 1.2)

- 도구: CodeGraph **1.6.0**(`codegraph --version`), 격리 worktree `~/tossos-work/a197-impl` 에서 `codegraph init .` 새 색인(2026-11-23, 15.3s; 파일 3,237 · 노드 55,950 · 엣지 206,351).
- 기준 커밋: `064f59fd99c11ea3a19adcdb804bbcdfed2470cb`(= `base-commit.txt`, 편집 0).
- 질의: `codegraph callers -l 100 -j <symbol>`(원문 JSON 은 세션 스크래치, 아래는 비시험 호출자만 옮김).
- **주의**: CodeGraph 는 **함수 단위·이름 단위**로 답한다 — 동명 함수가 섞이고(`submit` = `ExitObserver.submit` + `Gateway.submit`), 한 함수 안의 호출 **자리 수**는 세지 않는다. 자리 수는 `analysis/cited-ast/*.ast.json` 의 `calls` 로 셌다.

| 심볼 | CodeGraph 비시험 호출자(함수) | 자리 수(AST) | 비고 |
|---|---|---|---|
| `ExitObserver.submit` | `ExitObserver.record` (`exitloop.go:1314`) | 1 (`:1453`) | 이름 충돌 3: `Gateway.place/Cancel/Amend` → `Gateway.submit` |
| `ExitObserver.alertProposalRefused` | `ExitObserver.submit` | **2** (`:1530` Guardian 거절, `:1607` default) | |
| `ExitObserver.clearTheSymbol` | `ExitObserver.record` | 1 (`:1395`) | |
| `ExitObserver.record` | (26 건 — 이름 충돌 다수) | 엔진 루프 판정 자리 | a163 렌더 해시가 본문 고정 |
| `ExitObserver.release` | (62 건 — 이름 충돌 다수) | `submit` 안 5 (`:1512` `:1531` `:1548` `:1589` `:1612`) | |
| `isProtective` | `exitLimitPrice`(`exit_lower_limit.go:278`) · `record` · `submit` | `submit` 안 3 (`:1499` `:1553` `:1562`), `record` 안 2 | |
| `classifyMutation` | `Gateway.submit`(`gateway.go:470`) | 1 (`:729`) | |
| `classifyRefusalCode` | `classifyMutation` | 1 (`classify.go:44`) | |
| `Cancel` (`Gateway.Cancel`) | `ExitObserver.clearTheSymbol`(`:1706`) · `flatten.Saga.settleStep`(`flatten.go:382`) · `ops.cancelHandler`(`write_operations.go:194`) · (인터페이스 선언 `ExitSubmitter.Cancel` · `protection` 동명) | 각 1 | |
| `IssueReduction` | `flatten.Saga.decisionFor`(`execgw.Issuer`) · `ExitObserver.submit`(`:1517`) · `ExitObserver.clearTheSymbol`(`:1684`) | 엔진 2 자리 = `RiskGuardian.IssueReduction`(생산) | **두 개의 다른 함수**: `execgw.Issuer.IssueReduction(ReductionRequest{Kind,Side})` vs `RiskGuardian.IssueReduction(ReductionIssuance)` |
| `Journal.LiveOrdersForSymbol` | `noteHeldProposal`(`exit_held_proposal.go:126`) · `clearTheSymbol`(`:1650`) · `Gateway.cancelRefusalLatch`(`cancel_refusal_latch.go:20`) | 각 1 | |
| `Journal.EngineBrokerOrderIDs` | `manual_buy_refresh.go:47`(a136) | 1 | 주문 번호 → 엔진 귀속 집합의 유일한 기존 판독 |
| `execgw.ScanOrders` | `findSuccessors`(`amend_indoubt.go:415`) · `scanBoth`(`indoubt.go:409`) · `lookupClosed`(`filldetect/closed_lookup.go:197`) · `collect`(`filldetect/detect.go:375`) · `flatten.Saga.enumerate`(`flatten.go:495`) · `reconcile.Collector.Collect`(`snapshot.go:257`) | 각 1 | |
| `Context.ExitObserver` | `engineRuntime`(`cmd/tossctl/engine.go:671`) | 1 | |
| `recordGateSettings` | `NewContext`(`engine.go:451`, 호출 `:486`) | 1 | `NewContext` 는 a166 census 고정 함수 |

## 영향 (impact) 요약

- `submit` 편집의 호출 사슬: `ObserveOnce → observe/judge → record → submit` — 엔진 exit 루프 goroutine 하나. 다른 패키지 호출자 0.
- `alertProposalRefused` 시그니처 변경 시 편집 자리 2(둘 다 `submit` 안).
- `Context.ExitObserver` 편집의 호출자: `engineRuntime` 1. `NewContext` · `buildGateway` 는 a166 census(`internal/accountbudgetruntime/testdata/a166-wiring/shapes.json`) 고정.
