# Evidence reconciliation — a202

| 항목 | CodeGraph 1.6.0 | 현재 HEAD(base `7bc35e5d1`) 직접 읽기 · AST | 결론 |
|---|---|---|---|
| `TestTheReplacedReaderIsClosed` 의 `attach`·`inFlight`·`wake` | `console_lifecycle_attach.go:95·110·127` 로 해소 | 수신자는 `a109Attachment` 가 돌려주는 `*strategyRuntimeAttachment` — `httpapi_strategy_attach.go:158·151·166` | **불일치 — CodeGraph 가 같은 이름 메서드를 다른 타입으로 해소.** HEAD 가 정본. 편집 대상(시험 대기 조건)에 영향 없음 |
| 닫기 위치 | `closeEvictedStrategyReader` 호출자 `attempt` 하나 | `:210` `a.mu.Unlock()` → `:211` 닫기(잠금 밖) | 일치. 시험이 `!inFlight()` 만 기다리면 `:210-211` 창이 열린다 |
| 자식 수명 | (CodeGraph 범위 밖 — 런타임 동작) | 자식 직접 실행: `READY` 뒤 0.10 s exit 2, `fatal error: all goroutines are asleep - deadlock!`(`../red/shadowsession-child-before.txt`) | 측정이 정본 |
| a166 census 고정 여부 | — | `rg` 대상 파일 · 함수명 in `internal/accountbudgetruntime` → 일치 없음(`a166-census-check.txt`) | 편집이 a166 census 를 깨지 않는다 |
