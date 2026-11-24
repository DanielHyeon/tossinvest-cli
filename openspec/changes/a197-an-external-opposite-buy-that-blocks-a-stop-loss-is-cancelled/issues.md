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

### I1 재개 조건 진행 (2026-11-25)

1. a198 착지 — **충족**: a198 은 2026-11-24 archive `7fec501f1`(origin/main) · 운영 반영(Manager 전달). 이 브랜치는 origin/main `b93ed133a` 를 병합하고 base 를 재고정(`f18de89ef`, Manager 승인 2026-11-25).
2. 선행 2 Manager 결정 — **충족**: (E) 종목 범위 park(Manager 결정 2026-11-25, `review.md` 「Manager 결정 (2026-11-25)」). 근거 `analysis/resume/decision-2-indoubt.md` · `decision-2-addendum.md`.
3. §7 설계 변경 반영 + `openspec validate --strict` — **충족**: proposal · design(D1~D11, §2~§7) · spec delta(exit-policy ADDED 7 · MODIFIED 4, order-execution MODIFIED 1) · tasks 개정, strict valid.
4. 개정본 r2 적대 리뷰 — `review.md` 「proposal-freeze 적대 리뷰 r2」.
- 남은 열린 항목: design O-1(토글 audit 실패 처리) — Manager 결정 대기.

## I2 — flip 전 사람 확인 두 가지 (Manager 결정 2026-11-25, 비차단 — 운영 반영 뒤 토글 flip 의 선행 조건)

- **I-4 계좌 단독 사용**: 「외부」= 엔진 원장에 없는 모든 번호(design D2). 이 계좌에 주문을 내는 다른 자동화 · 다른 주문 원천이 없는지, 있으면 그 매수가 취소돼도 되는지 사람이 확인한다. 코드 무변.
- **I-5 실계좌 PENDING OPEN 행 모양**: 영수증 — 공식 스펙 1.2.24 `docs/migration/openapi.latest.json`(Order 필수 `orderId` · `symbol` · `side` · `quantity` · `currency` · `execution`; OrderExecution 필수 `filledQuantity` 문자열 「체결 내역이 없으면 filledQuantity=0」; `price` 는 문자열 또는 null; 예시 `pendingMixed`), 운영 판독기 `internal/filldetect/payload.go` · `internal/official/orders_reads.go` 가 같은 칸을 문자열로 읽음. 실측은 없다 — `openspec/changes/verify-execution-capability/measurements.md` M31(CLOSED+OPEN 197 건에서 관측 status 는 CANCELED · FILLED · REJECTED 뿐, PENDING 미관측), M25 · M32 는 단건 GET. `parseExternalBuyRow` 와 스펙은 어긋나지 않음(시험 `TestA197TheSpecShapedOpenRowIsRead`, 값은 합성). → flip 전 **읽기 전용 OPEN 조회 1 회** 사람 실측(design §5-1).
