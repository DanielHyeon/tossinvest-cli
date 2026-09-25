# CodeGraph baseline — a115

- 날짜: 2026-09-26 · 기준 HEAD `4798d399`(base `8688f74f` 뒤 문서 커밋뿐, 대상 코드 동일) · codegraph **1.6.0**

| 질의 | 결과 |
|---|---|
| `callers resolveStrategyRuntimeReader` | 1: `strategyRuntimeReaderFor`(httpapi.go:257) — 부팅 해석 :270 + 재시도 클로저 :266. 테스트는 호출자 경유(`a108_…_test.go` · `a109_the_daemon_reattaches_…_test.go`) |
| `strategyRuntimeAttachment` 참조 | 3 호출자(생산 1 `strategyRuntimeReaderFor` 인스턴스화 + 테스트 `a109Attachment` 등) · implements: `MultiMarketStrategyRuntimeReader` · `StrategyRuntimeReader` · `StrategyRuntimePresence` — 콘솔 소비 인터페이스를 이미 충족. CodeGraph 의 `MarketScheduleReader` implements 는 이름 해소 오탐(반환 타입 `MarketScheduleReading` 불일치, freeze 리뷰 P2-8) |
| `callers buildMultiMarketStrategyRuntimePage` | 1: internal/console/strategy_runtime.go · 테스트 `strategy_runtime_multimarket_test.go` |
| `callers strategyRuntimeSummary` | 4: `strategyEntries`(settings_tabs.go:223) + 테스트 3(`TestStrategyRuntimeSummary…` 셋) |
| `runConsole` 전략 블록 | console.go:397–407 — dial :399 · dialErr → 경고만, nil 유지 :400–401 · 비부재 stat 오류 → 경고+nil :405–406 · 성공 → client :403 |
| 소비자 nil 판정 자리 | strategy_runtime_multimarket.go:48(`Unwired`)·:49(Read 여부) · settings_tabs.go:291 |
| `affected console.go strategy_runtime_multimarket.go settings_tabs.go --filter '*_test.go'` | **863** 파일(기본 판별식은 Go 테스트 0 — 하네스 주석의 `isTestPath` 결함, a114 와 동일) |
