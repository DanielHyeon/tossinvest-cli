# a197 Issues

## I1 — tasks 1.8 [blocked]: design D6 불성립 (2026-11-23)

- **분류**: blocking(계약이 기대는 기존 경로가 생산에서 동작하지 않음) — WORKFLOW 「예외 경로」 ①.
- **막힌 이유**: design D6 은 외부 매수 취소를 「기존 취소 경로(`Issuer.IssueReduction` → 게이트웨이 취소)」로 낸다고 적었다. exit 루프의 생산 발행기는 `*execgw.RiskGuardian`(`exitwiring.go:326,339`)이고, 그 `IssueReduction`(`riskguardian.go:587-634`)은 Side 를 SELL 로만 받고 결정을 `Kind: journal.KindPlace`(`:614`)로 기록해 멱등 키가 붙는다(`journal/decision.go:373-375`). 그 결정으로 `Gateway.Cancel` 을 부르면 `prepareRequest` 가 키를 요청에 옮기고(`gateway.go:1081`) `journal.Prepare` 가 「CANCEL carries no idempotency key」로 거절한다(`journal/durability.go:267-269`) — attempt 0 · 브로커 호출 0.
  - 실측 원문: `analysis/probe/probe-output.txt`(probe 원본 `analysis/probe/zz_a197_probe_test.go.txt`, 버리는 worktree 에서 실행). BUY·SELL 모두 `state="" reason=invalid_mutation_request broker_cancels=0`; 대조군 `execgw.Issuer(Kind CANCEL, Side=주문 방향)` 은 둘 다 `CONFIRMED broker_cancels=1`.
  - 같은 결함이 기존 a094 길 치우기(`ExitObserver.clearTheSymbol` `:1684` → `:1706`)에 있다 — a197 과 무관한 운영 결함.
- **선행 조건**: a198(길 치우기의 취소 결정 발행 수리) 착지 — 취소 결정이 주문 경로를 통과하고 실제 게이트웨이를 거치는 시험이 생긴다. a197 의 외부 취소도 그 발행 경로를 쓴다.
- **재개 조건**(전부):
  1. a198 착지.
  2. design §7 선행 2(r1 #3, 이 단계가 낸 취소의 IN_DOUBT 처리)에 대한 Manager 결정.
  3. design §7 의 재개 시 설계 변경을 proposal · design · spec delta · tasks 에 반영하고 `openspec validate --strict` 통과.
  4. 개정본에 대한 proposal-freeze 적대 리뷰(r2, 다른 계열) — 이번 r1 은 개정 전 문서에 대한 것이다.
- **보존**: 1단계 산출물(`analysis/`)은 a198 의 근거로 쓰인다. 재개 때 base 이후 a198 이 같은 함수를 고쳤으면 AST · FLM 을 다시 뽑는다.
