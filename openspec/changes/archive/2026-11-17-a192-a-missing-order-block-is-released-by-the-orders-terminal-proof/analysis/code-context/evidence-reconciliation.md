# Evidence reconciliation (a192, 1단계)

기준 commit `6277ec34e`. 비교한 계층: CodeGraph 1.6.0(이 worktree 색인) · `tools/logic-map` AST · 현재 HEAD 파일 직접 읽기 · Manager 서술.

| 주장 | CodeGraph | AST | HEAD 직접 | 판정 |
|---|---|---|---|---|
| `Observe` 의 production 호출 자리는 하나 | 이름 매칭 49(잡음) | `RunOnce` call@453 | `rg` 1 | 일치(CodeGraph 이름 잡음은 HEAD 로 좁힘) |
| missing-order 는 첫 blocking 비교에서 블록 | `blocksFor` 호출자 `Observe` 1 | `Observe` B3·B17, `blocksFor` B2 | 리터럴 값 | 분기는 AST, 값은 손 — 표시함 |
| missing-order 만 있는 종목의 자동 해제 0 | `AdjustmentApplied` 호출자 production 1(수렴) | `ConvergeQuantities` B1, `Observe` B8 | `SET released_at` 2곳 | 일치(+ 단서: 별도 수량 수렴이면 풀림) |
| 청산은 블록에 안 막힌다 | `ExitAllowed` production 호출자 0 | `ConfirmedFloor` B2, `applyFloor` B2 | — | **Manager 서술과 불일치** — 막히지는 않으나 하한(왕복 2·실패 시 0)이 붙음. design §8 |
| 30s 대사 간격 | `Due` 호출자 0 | `Run` call@403 Sleep(period) | 상수 60s | **Manager 서술 보충** — 운영 주기는 60s |
| 영구 행이 승격 분쟁을 전부 기록 | — | `thresholdKey` B2·`promotionDisputeKeyLess` B2 | 배열 리터럴(kind 첫 원소) | 하나만 기록, missing-order 가 먼저 — Q8 반증 ① |
| 결속을 기존 열로 유도 가능 | — | `EnterReconcile` B8, `enterReconcileScopeInTx` B3 | DDL·SQL | 불가 — Q7 (b) 기각 |
| 영향 시험 수 | 기본 1 / 필터 1317·278·24 | — | — | **기본값 불신**(하네스 주석) — 필터 결과를 씀 |

불일치: 위 표의 두 줄(Manager 서술)은 코드와 HEAD 로 해소했다(design §8). CodeGraph 와 HEAD 가 다른 자리는 이름 매칭 잡음뿐이며 HEAD `rg` 로 좁혔다.
CodeGraphContext·GBrain 은 이 worktree 에서 돌지 않았다(`codegraphcontext-context.md`) — 결론은 CodeGraph + AST + HEAD 로만 냈다.
