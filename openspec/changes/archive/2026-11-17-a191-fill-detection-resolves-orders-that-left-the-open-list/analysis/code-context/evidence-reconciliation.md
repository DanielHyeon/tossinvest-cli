# Evidence reconciliation (a191, 1단계)

기준 commit `a392308e8`. 비교한 계층: CodeGraph 1.6.0(이 worktree 색인) · `tools/logic-map` AST · 현재 HEAD 파일 직접 읽기 · Manager 서술.

| 주장 | CodeGraph | AST | HEAD 직접 | 판정 |
|---|---|---|---|---|
| `collect` 의 호출 자리는 하나 | `pollLocked` 1 | `pollLocked` call@293 `d.collect` | 같음 | 일치 |
| 목록 이탈 추적 주문은 단건 응답만으로 파생 | — | B13~B19·call@428·@454 | 같음 | 일치(Manager 의 손 읽기를 AST 가 확인) |
| 감지기는 결과를 기록하지 않음 | `Run` 은 `PollOnce` 호출자 | `Run` 분기 3, 로그 call 0 | `_, _ = d.PollOnce(ctx)` | 일치 |
| 감지기 생성 자리는 하나 | `engineRuntime` 1 + 시험 2 | — | engine.go:771 | 일치 |
| 「추적 집합 = terminal 0」 | 호출자만 | Go 부분만(분기 B1~B10) | SQL `f.terminal = 0` | AST 는 SQL 을 못 봄 — **HEAD 직접 읽기로 확정**, design 에 「손으로 읽음」 표시 |
| missing-order 블록은 조정 크레딧 없이 안 풀림 | — | `Observe` B6·B7·B8 | 리터럴 값(Cause·Release) | 분기는 AST, 값은 손으로 — 표시함 |
| `reconcile-resolve` 는 원인 QUANTITY_MISMATCH 만 받음 | 호출자 | `buildReconcileResolveRuntime` B12~B14 | 클로저 본문 | 일치 |
| 영향 시험 수 | 기본 1 / 필터 1336·563·72 | — | — | **기본값 불신**(하네스 주석) — 필터 결과를 씀 |

불일치: 없음(CodeGraph 와 HEAD 가 다른 자리 0). CodeGraphContext·GBrain 은 이 worktree 에서 돌지 않아(`codegraphcontext-context.md`) 교차검증
대상이 아니다 — 결론은 CodeGraph + AST + HEAD 로만 냈다.
