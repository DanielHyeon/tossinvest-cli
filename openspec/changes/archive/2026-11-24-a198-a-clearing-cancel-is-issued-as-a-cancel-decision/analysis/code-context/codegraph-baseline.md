# a198 CodeGraph baseline (tasks 1.2)

- 도구: codegraph **1.6.0** (`codegraph init -y .` — 이 worktree 새 색인, 2026-11-23). 색인: 파일 3,239 · 노드 55,950 · 엣지 206,351.
- 기준 커밋: `e1eaf720c5fa1622321b98e73a792d2e5762b3d2`(= base-commit.txt, origin/main).
- CodeGraph 는 **함수 단위·이름 기반**으로 답한다. 자리 수는 `grep -rn` 으로 셌고(아래 「자리」 열), 함수 내부 분기는 `analysis/cited-ast/*.ast.json`(tools/logic-map) 이 근거다.

## 질의와 결과

| 질의 | 결과(함수 단위) | 호출 **자리**(grep, HEAD) |
|---|---|---|
| `callers clearTheSymbol` | `ExitObserver.record`(exitloop.go:1314) 1 | 1 — exitloop.go:1395 |
| `callers noteDelay` | `record`·`submit` 2 | 길 치우기 관련 1 — exitloop.go:1402 (나머지 1588 은 submit 의 SymbolInFlight) |
| `callers noteClearFailure` | `record` 1 | 1 — exitloop.go:1403 |
| `callers IssueReduction` | 9 (이름 혼합: RiskGuardian·Issuer·인터페이스·시험) | 엔진의 `o.opts.Issuer.IssueReduction` 2 — exitloop.go:1517(submit, 무편집 D5) · 1684(clearTheSymbol, 편집 대상) |
| (grep) `o.opts.Submit.Cancel(` | — | 1 — exitloop.go:1706 |
| `callers ExitObserver` | 31 (생산 `engineRuntime` 1 + 시험 30) | 생산 1 — cmd/tossctl/engine.go:671; 시험 `ctx.ExitObserver(…Options{` 11 |
| (grep) `ReductionIssuer` | — | 생산: 선언 exitloop.go:131 · 필드 :180 · 단언 exitwiring.go:326; 시험 래퍼 2(a111:66, a129:64) |
| `callers decisionFor` | `Saga.settleStep`(flatten.go:301) · `Saga.sell`(liquidate.go:444) | 취소 자리 1 — flatten.go:360 |
| `callers prepareRequest` / `checkPreimage` / `checkIdempotencyKey` | 각 1 (`submit` / `checkDecision` / `checkDecision`) | gateway.go:509 / guardian.go:290 / guardian.go:306 |
| `callers Prepare` | 19 (생산 `Gateway.submit` 1 + journal 시험 헬퍼) | 생산 1 — gateway.go:548 |
| `callees clearTheSymbol` | LiveOrdersForSymbol · UnsettledOnSymbol · onlyEngineCancelsInFlight · clearTargets · ConfirmedCancelOf · warnEpisode · IssueReduction · floatOf · workingOrderPrice · Cancel · ReleaseClearedExitProposal (+ 구조체) | — |
| `callees noteDelay` | Now · delayBound · alert · label | — |
| `callees noteClearFailure` | recordEpisodeCritical · label · why | — |
| `callees decisionFor` | `execgw.Issuer.IssueReduction`(issue.go:86) · randomID · DecisionTTL | — |
| `affected exitloop.go exitwiring.go` | 기본 판별식: 1 (auth-helper/tests/test_cli.py) | `--filter '*_test.go'`: **1,349** 시험 파일(깊이 5) — 하네스 주석의 판별식 결함 그대로 |

## 확인한 심볼 (정의 자리, HEAD)

| 심볼 | 정의 | AST |
|---|---|---|
| `ExitObserver.clearTheSymbol` | exitloop.go:1649-1743 | cited-ast/ExitObserver_clearTheSymbol.ast.json (분기 17) |
| `ExitObserver.record` | exitloop.go:1314-1455 | ExitObserver_record.ast.json (19) |
| `ReductionIssuer` | exitloop.go:131-133 | (타입) |
| `Context.ExitObserver` | exitwiring.go:319-365 | Context_ExitObserver.ast.json (6) |
| `NewExitObserver` | exitloop.go:319 | NewExitObserver.ast.json |
| `RiskGuardian.IssueReduction` | riskguardian.go:587-634 | RiskGuardian_IssueReduction.ast.json (4) |
| `execgw.Issuer.IssueReduction` | issue.go:86-103 | Issuer_IssueReduction.ast.json (1) |
| `Gateway.Cancel` / `submit` / `prepareRequest` | gateway.go:407-426 / 470-766 / 1011-1095 | Gateway_Cancel / Gateway_submit (57) / Gateway_prepareRequest |
| `Gateway.checkDecision` · `checkPreimage` · `checkIdempotencyKey` · `checkLimits` | guardian.go:235 · 319 · 395 · 427 | 각 .ast.json |
| `PrepareRequest.validate` · `validateDecisionBinding` · `Journal.Prepare` | durability.go:194 · 238 · 373 | 각 .ast.json |
| `DecisionRequest.build` | decision.go:310-391 | DecisionRequest_build.ast.json (B22 = 키 파생) |
| `Saga.settleStep` · `Saga.decisionFor` | flatten.go:301 · 659 | 각 .ast.json |
| `trading.Service.Cancel` · `guard` | service.go:210 · 237 | 각 .ast.json |
| `risk.Evaluate` · `evaluateReduction` | chain.go:131 · 227 | 각 .ast.json |
| `Journal.LiveOrdersForSymbol` · `ConfirmedCancelOf` | fills.go:1854 · exit_proposal_release.go:150 | 각 .ast.json |
| `Gateway.checkSymbolFree` · `unsettledFor` · `checkEntry` · `checkProtection` · `checkReservation` · `withCancelRefusal` · `loadDecision` · `checkDecisionUnspent` | gateway.go · a094_shared_judgements.go · protection.go · cancel_refusal.go · guardian.go | 각 .ast.json |
| `ExitObserver.noteDelay` · `noteClearFailure` · `onlyEngineCancelsInFlight` | exitloop.go:1922 · exit_held_proposal.go:266 · :61 | 각 .ast.json |

## 시험·설정 결속

- 운영 조립: `engineRuntime`(cmd/tossctl/engine.go:655) → `ectx.ExitObserver`(:671). `opts.Issuer = c.Guardian.(ReductionIssuer)`(exitwiring.go:326·339), `opts.Submit = c.Gateway`(:340). `opts.Log` 는 생산에서 설정되지 않는다(engine.go:671-683 리터럴·exitwiring.go 어디에도 없음).
- 취소 정책: `trading.Service.guard`(service.go:237) 가 `policy.Cancel`·`AllowLiveOrderActions` 를 본다. 엔진 기동 interlock clause 3 `checkTradingPolicy`(interlock.go:597, 호출 :513)가 `trading.cancel` 꺼짐을 기동 거부로 막는다. 정책 재적재 경로 없음(trading.Service 에 Set/Update 메서드 0).
- 취소 거절 재전송: 엔진 게이트웨이 `CancelRefusal: execgw.EngineCancelRefusal()`(internal/app/engine/gateway.go:329) — retry matrix 행 31(0회).
