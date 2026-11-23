# Evidence reconciliation (a195, 1단계)

기준 commit `064f59fd9`. 비교한 계층: CodeGraph 1.6.0(이 worktree 색인) · `tools/logic-map` AST(`../cited-ast/`) · HEAD 파일 직접 읽기·grep · 선택 시험 커버리지(`../coverage/`) · Manager 서술(proposal·design).

| 주장 | CodeGraph | AST | HEAD 직접 | 판정 |
|---|---|---|---|---|
| 운영자 해제는 `QUANTITY_MISMATCH` 외 원인을 거절 | 호출자 1 | `buildReconcileResolveRuntime` B13 range@307 → B14 if@308 → return@309 | 문구 일치 | 일치 |
| 거절은 잠금·스냅숏 뒤 | — | `runEngineReconcileResolve` B7 if@134(잠금) · B11~B19(스냅숏 3) · B21 if@198(비교) · B22 if@201(validate) | 같음 | 일치 |
| `ATTRIBUTION_FAILED` 를 푸는 다른 자리 0 | 해제 쓰기 3자리 | `Observe` B7 if@667 · `orderTerminalReleaseVerdict` B3 if@182 | 쓰기 grep 3 | 일치(자리 표시 정정: `order_terminal_release.go:184` 는 B3 if@182 의 조건 줄) |
| `Tracker.Resolve` 가 `ExpectCause: block.Cause` | — | `Tracker.Resolve` B4 range@1363 · B5 if@1364(건너뜀) | 실제는 `firstNonEmpty(block.Cause, QUANTITY_MISMATCH)`, `ScopeMarket` 없음 | **서술 정정**(결정 불변 — `ATTRIBUTION_FAILED` 행은 scope_market NULL 로만 쓰인다: position_projection.go:409 INSERT 에 열 없음) |
| 계좌 전체 `ATTRIBUTION_FAILED` 를 쓰는 자리 없음(D5) | — | — | 빈 symbol 로 쓰는 6자리 전부 `IDENTIFIER_CONFLICT`(position_projection.go:204 · reservation_release.go:612 · position_campaign.go:1148·1292 · fills.go:1808 · lineage.go:450). `ATTRIBUTION_FAILED` 는 position_projection.go:377 `origin.Symbol` 하나 | 일치(빈 `origin.Symbol` 이 그 자리에 닿을 수 있는지는 미측정 — D5 는 어차피 거절) |
| **D2: 명령의 재조회 일치 = 불확실성 해소** | — | `runEngineReconcileResolve` B21 if@198 `diff.BlocksEntry()` | `BlocksEntry` = `Quantities`·`MissingOrders` 만(compare.go:312). 비교는 **계좌 전체**(Compare 가 계좌∪장부 전 종목·전 미체결). 외부 포지션·외부 주문·평단 차이는 거절 사유가 아님. `HeldBySymbol` 은 CLOSED 인스턴스를 건너뛰어(compare.go:155) 닫힌 종목의 재매수 보유는 `ExternalPos` | **부분 참 — Manager 결정 필요**. `Observe` 의 해제 규칙은 같은 `ExternalPos` 를 「분쟁」으로 본다(`symbolsInDispute` mismatch.go:585-593, 개정 2 D8). 이번 사고 상태(닫힌 뒤 재매수, 미편입)는 명령의 비교를 통과한다 |
| 운영자 경로 시험이 production validate 를 잰다 | — | — | production 클로저를 부르는 시험 0(grep `rt.validate`·`runtime.validate`), 커버리지 302~314 전 블록 hit=0 | **design §3 「하네스 재사용」과 불일치** — 기존 통합 시험(engine_reconcile_test.go:321)은 validate 를 손으로 복제한다 |
| 편집은 validate 한 줄 | impact 4 | — | a166 census shapes.json 이 본문 전체 고정 | **census 적중** — tasks 1.4 의 정지 조건 |

불일치: 위 표의 정정 2줄은 문서에 반영했다(proposal 표 3·4행). D2·시험 하네스·census 3줄은 결정·범위·시험 전략에 닿으므로 문서를 고치지 않고 Manager 에게 올렸다. CodeGraph 와 HEAD 가 다른 자리는 이름 매칭 잡음(`Resolve` 17, `Tracker.Resolve`→`Observe` 오인)과 데이터·텍스트 결속(shapes.json, 파일 텍스트 시험) 미탐지뿐이다.
