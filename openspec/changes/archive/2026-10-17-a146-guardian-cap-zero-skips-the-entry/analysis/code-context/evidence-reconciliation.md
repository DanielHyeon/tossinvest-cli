# a146 evidence reconciliation

| 항목 | CodeGraph | HEAD 읽기 · AST · 시험 | 결론 |
|---|---|---|---|
| `admit` 생산 호출자 | 이름 충돌 — `RunBounded` · `coordinateMarketProposals` 등 다른 `admit` 과 섞이고 `dispatch` 누락 | grep: `strategy_dispatch_cycle.go:151 cycle.firstLeg.admit(ctx, result)` 하나(`raw/grep-bridge-admit-nontest.txt`) | **불일치** — CodeGraph 가 필드 경유 메서드 호출을 놓침. HEAD 가 권위: 생산 호출자 1 |
| 잠금의 저장 위치 | `latchMarket` 호출자 `runMarket` | AST `latchmarket` 1139 · 1164, `Faults()` 생산 독자 0(`raw/grep-Faults-readers.txt`), durable 잠금 writer 는 `persistMarketLatches` 뿐 | 감독자 잠금은 메모리 — a139 review 의 「원장 잠금」 서술 정정(design F6) |
| 생산 감독자의 worker 종류 | `NewRefreshingPairedStrategyEntrySupervisor` 호출자 `engineRuntime` | AST B3(408) — `RefreshesAuthority: true` 만, `effective = ` 대입은 `false` 하나(`raw/grep-production-supervisor.txt`) | 생산은 refreshOnly 갈래(design F7) — 특성 시험 두 종류로 둘 다 잼 |
| 상한 0 → 잠금 사슬 | callers 사슬 | 특성 시험 4 건 편집 전 HEAD 통과(`analysis/red/characterization-head.txt`) | 일치(effective) / 정정(생산 refreshOnly 는 삼킴 + 뒤 범위 굶음) |
| 범위 거절 → 잠금 | — | AST `delivereachstrategyhandoff` B3 · 92, 특성 시험, spec 618-621 | 명시된 동작 — a146 밖(design D7) |
