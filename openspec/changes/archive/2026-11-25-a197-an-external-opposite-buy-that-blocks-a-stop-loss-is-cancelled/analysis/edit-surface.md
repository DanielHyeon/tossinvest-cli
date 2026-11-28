# a197 편집 표면 · rate limit 계상 자리 (tasks 1.4 · 1.5)

## 0. 재개 준비 갱신 (2026-11-25, HEAD `f18de89ef`, base `b93ed133a`) — 아래 §1.4~§1.5 는 2026-11-23 옛 base 원문

실행 0(grep · AST · git 읽기만). 판정이 바뀐 것만 적는다.

| 항목 | 2026-11-23 | 2026-11-25 현재 | 근거 |
|---|---|---|---|
| `a087_submit_structure_test` · `a163_judgement_pin_test` · `a094_exit_clear_test` | 표 §1.4-1 | **무변** — 세 파일 모두 옛 base 이후 커밋 0. a163 렌더 해시 10 개(`record` 포함, `submit` 미포함) 그대로 | `git log 064f59fd9..HEAD -- <세 파일>` 0 줄 |
| a094 「한 판정」 핀(`TestA094TheClearAndTheGatewayShareOneUnsettledJudgement`, `a094_exit_clear_test.go:560`) | 무관 | **선택지 (C) 에 관련** — `clearTheSymbol` 이 `UnsettledOnSymbol` 을, `checkSymbolFree` 가 `unsettledFor` 를(그리고 `PendingAttempts` 직접 호출 0), `UnsettledOnSymbol` 이 `unsettledFor` 를 부르는지 본다(`:590-598`) | 시험 원문 |
| 새 구조 시험(옛 base 이후) | — | **1 개 추가**: `a198_clear_cancel_log_order_test.go:140` `TestA198TheCancelFailureQueueIsFlushedByTheCycleEnd` — `endExitOrderCycle` 이 `flushClearCancelFailures` 를 부르고 `noteClearCancelFailure` 가 `Warn` 을 부르지 않음(판정 경로 안 쓰기 금지)을 본다. a197 이 새 단계의 로그 줄을 같은 전용 로거에 쓴다면 같은 규칙(주기 끝 쓰기 — a198 D4 재결정)을 따라야 함. `submit`·`alertProposalRefused`·`clearTheSymbol`·`record` 를 굳히는 구조 시험은 여전히 위 셋뿐 | rg `go/parser`·`go/ast` + 원문 |
| a166 census 고정 함수 | 90 개, 엔진 6 | **90 개 그대로**, `NewContext` 여전히 고정(렌더에 `recordGateSettings(… cfg.Engine.ExitSourceAgeReport)` 포함). 옛 base 이후 이 census 를 바꾼 커밋 3(`c29b4f0f8` a195 · `3445403ee`·`fadd2fc4c` a199 — `NewContext`·`buildReconcileResolveRuntime` 항목 갱신) — **Manager 승인 아래 census 갱신의 선례가 생김**. `exitloop.go` · `exitwiring.go` · `a198_*` · `in_doubt_settle.go` · `a094_shared_judgements.go` 는 목록에 없음. `sites.json` 재고 116 그대로 | `shapes.json` · `sites.json` 읽기 |
| `recordGateSettings` | L739-813 | L774-848, 본문동일 | `cited-ast/engine--recordGateSettings.ast.json` |
| 진행 중 change a087 | 활성, `landed-commit.txt` 없음 | **그대로** — `feat/a087-lower-limit-exit` 는 origin/main 대비 ahead 0, `landed-commit.txt` 여전히 없음. a198 종결 기록(§6 마지막 항)대로 `exitloop.go` 해시가 a087 번들과 이미 어긋나 있어 a087 종결 쪽 비용은 a197 과 무관하게 이미 생겼음 | `git rev-list --count origin/main..feat/a087-lower-limit-exit` = 0 |
| 진행 중 change a167 | 등록 단계 | **그대로**(브랜치 0, proposal `:98` FLM 대상 `record` · `submit`) | 디렉터리 · `git branch -a` |
| 엔진 관측기가 도는 기동 | 검증 기동만 | **청산 전용 기동에서도 돎**(a199, `Context.ExitObserver` B2@323 `RunsExitSide()`) — a197 단계가 켜지면 그 기동에서도 열림. 취소는 노출을 늘리지 않아 진입 래치에 걸리지 않음(`checkEntry` `gateway.go:859-861` `!plan.raisesExposure` → nil) | AST · 원문 |

### 0.1 rate limit 계상 정정(§1.5 의 「대상 매수당 취소 1」)

- 엔진 게이트웨이의 취소는 브로커 어댑터에서 **사전 단건 조회 1**(`GET /api/v1/orders/{id}`, `precheckTimeout` 2s, `internal/app/engine/precheck.go:58·99`)을 먼저 한다 — a198 design §2-4 와 같은 사실. 그래서 a197 의 대상 매수 1 건 = 사전 단건 조회 1 + 취소 POST 1(HTTP 클라이언트 시한 15s, `internal/official/client.go:22`). a198 §2-4 가 적은 대로 토큰 잠금 대기(`token.go:63`, context 로 끊기지 않음) · 401 재전송 · a166 예산 admission 은 그 시한 밖이라 **한 건의 최악 지연은 미확정**이다 — design §7 r1 #4(「단계 전체 시한」)의 전제(아래 `resume/decision-2-indoubt.md` Q5).
- §1.5 의 §3 행(「외부 반대 매수 확인 (a197)」)은 「주기당 최대 1 조회(× 쪽 수) + 대상 매수당 (사전 단건 조회 1 + 취소 1)」로 적어야 한다 — 문서 갱신은 구현 단계(tasks 3.9)에서.


기준 `064f59fd9`. 실행 0(grep · AST · git 읽기만).

## 1.4-1 `submit` · `alertProposalRefused` 를 고정하는 구조 시험 (grep 전수)

`internal/app/engine/*_test.go` 중 `go/parser`·`go/ast`·render 해시를 쓰고 `exitloop.go` 를 여는 파일 5개를 열어 무엇을 고정하는지 봤다.

| 시험 | 고정하는 것 | `submit` | `alertProposalRefused` | a197 편집에 대한 판정 |
|---|---|---|---|---|
| `a087_submit_structure_test.go` `TestA087SubmitUsesOneChoiceAndMarksOnlyAfterPlace` | 이름으로 찾은 첫 `submit` 본문의 **호출 순서·인자**: `exitLimitPrice < applyFloor`, `exitOrderLineFor` 6번째 인자 `choice`, `sellIntent` 3번째 인자 `choice.price`, `Place < markProtectiveAttempt` | 예 | 아니오 | B15 안에 이 다섯 이름의 호출을 더하지 않으면 통과(같은 이름 호출을 더하면 마지막 위치로 덮여 깨질 수 있음) |
| `a163_judgement_pin_test.go` (a) 렌더 해시 | `judge · judgeLadder · judgeRatchet · noteHeldProposal · quoteUsable · record · refreshObservation · requiresFullJudgement · snapshotContext · stableObservationID` 의 주석 없는 렌더 sha256 | **아니오** | 아니오 | `record` 를 고치면 깨짐 — a197 계약은 `record` 무편집이어야 함 |
| `a163_judgement_pin_test.go` (b) 금지 census | 식별자 `SourceTimestamp` · `SourceAgeReport` 의 출현 자리 | — | — | 새 토글이 이 두 식별자를 쓰지 않으면 무관 |
| `a094_exit_clear_test.go` `TestA094TheClearAndTheGatewayShareOneUnsettledJudgement` | `clearTheSymbol` 이 `UnsettledOnSymbol` 을 부름 | 아니오 | 아니오 | 무관 |
| `a094_exit_clear_test.go` `TestA094TheSessionNeverRunsTheRestartRules` | `exitloop.go` · `exit_held_proposal.go` 전체에 `RecoverPending` 선택자 0 | 파일 전체 | 파일 전체 | 새 코드가 `RecoverPending` 을 쓰지 않으면 무관 |
| `a094_exit_clear_test.go` `TestA094TheClearAddsNoBrokerRead` | 3 주기 가격 판독 = 3(「다른 브로커 판독 없음」) | 행동 | — | 토글 꺼짐 하네스에서 새 판독 0 이면 통과 |
| `a090_workingset_exits_test.go` | `workingSet` 의 positions 루프 본문 | 아니오 | 아니오 | 무관 |
| `a111_flat_exit_observation_test.go` | `observe` · `quoteUsable` · `ObserveOnce` 렌더 | 아니오 | 아니오 | 무관 |

`functionBody` 류 공용 헬퍼로 `submit`·`alertProposalRefused` 를 고정하는 시험은 0. 본문 해시로 `submit` 을 고정하는 시험 0.

## 1.4-2 진행 중 change 와의 겹침 (읽기만)

| change | 상태 | 같은 함수 편집 |
|---|---|---|
| a087 `a087-a-protective-exit-is-a-market-order` | 활성, 미완료 체크 9(4.4 · 5.4 · 5.5 활성화 · 6.x 종결). 브랜치 `feat/a087-lower-limit-exit` = main 대비 **ahead 0**(구현 착지됨), `docs/a087-d7` ahead 17 — `git log main..docs/a087-d7 -- internal/app/engine/exitloop.go` 결과 0 | 지금 진행 중 편집 0. **단 a087 은 `landed-commit.txt` 가 없다** → a087 의 gate ⑤ 창 끝이 워킹트리이므로 a197 이 먼저 착지하면 a197 이 고친 `submit`·`alertProposalRefused` 의 FLM 을 a087 종결 때 a087 이 요구받는다(형제 착지 몫). 5.5 활성화는 `lowerLimitBoard` non-nil 배선(설계상 `submit` 본문 밖) |
| a167 `a167-a-protective-remainder-is-re-priced` | 등록 단계(proposal 만, tasks 미착수 19), 브랜치·worktree 없음 | 지금 0. 계획상 FLM 대상에 `ExitObserver.record` · `submit` 포함(proposal `:98`) → **장래 같은 함수** |

## 1.4-3 a166 census 적중

- `internal/accountbudgetruntime/testdata/a166-wiring/shapes.json` 고정 함수 90 중 엔진 관련 6: `cmd/tossctl/engine_assembly.go||assembleEngine` · `cmd/tossctl/engine_reconcile.go||buildReconcileResolveRuntime` · `internal/app/engine/engine.go||NewContext` · `…||NewOrderPath` · `internal/app/engine/gateway.go||buildGateway` · `…||newGatewayReplay`.
- `exitloop.go` · `exitwiring.go`(`Context.ExitObserver`) · `cmd/tossctl/engine.go`(`engineRuntime`) · `interlock.go`(`recordGateSettings`) · `config/engine.go` 는 **목록에 없음**.
- `sites.json` 재고(116): 비시험 Go 전체에서 선택자 이름 `Do` · `RoundTrip` · `NewClient` · `Budget`(accountbudget 의 것 제외) · `official.New*` · `accountbudget.Open*/NewTransport` · `net/http` 의 Get/Post… 를 센다. 새 코드가 `sync.Once.Do` 를 포함해 이 이름을 쓰면 재고 수가 바뀌어 census 가 깨진다.
- **적중 하나**: a163 선례대로 토글을 audit 하려면 `NewContext` 의 `recordGateSettings(…)` 인자 줄 편집이 필요 → 고정 함수 편집 → **멈춤 조건(Manager 결정)**.

## 1.4-4 기존 exit 토글 선례 → `cited-ast/README.md` (f)

## 1.5 rate limit 계상 자리

- 문서: `openspec/changes/archive/2026-07-26-harden-execution-base/retry-matrix.md` — §1 「endpoint × method × 오류 클래스」 표와 §3 「rate limit 예산 계상 (§0.4)」 표. a191 이 아카이브 뒤에도 이 파일에 행을 더한 선례가 있다(§1 `/api/v1/orders?status=CLOSED` 행, §3 「종결 목록 GET (a191)」 행). 수치 근거 문서는 `openspec/changes/verify-execution-capability/analysis/retry-matrix-slo-2026-10-22.md`.
- a197 이 더할 행(형식은 a191 행과 같게):
  - §1: `/api/v1/orders?status=OPEN&symbol=…` GET — 용도 「a197 손절을 막는 외부 매수 확인」, 재시도 0, 시한 3s(D4), 실패 = 취소 0 · freshness 영향 없음.
  - §1: `/api/v1/orders/{id}/cancel` POST — 기존 「엔진 청산 전 취소」 행에 a197 을 더함(409 already-processing 확정 종결, 재시도 0).
  - §3: 「외부 반대 매수 확인 (a197)」 — 토글 켜짐 ∧ 보호 손절 반대 미체결 거절이 난 주기만, 주기(5s)당 최대 1 조회(× 쪽 수, OPEN 은 실측상 1 요청 — slo `:47` M7) + 대상 매수당 취소 1. 정상 0.
- 그룹: a166 계좌 공통 예산 하나(350ms 최소 간격, `accountbudget/budget_linux.go:27`)이며 우선순위 클래스는 Normal/Critical 둘(`:33-34`). 보호 손절 경로의 선례는 Critical(`exitloop.go:872` · `:1393` · `:1554`). 읽기는 `Retrier` 를 거치지 않는 것이 a191 선례(`QueryOpenOrders` freshness 를 찍지 않음).
