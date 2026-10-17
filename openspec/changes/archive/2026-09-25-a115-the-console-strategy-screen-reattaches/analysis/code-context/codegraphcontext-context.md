# CodeGraphContext supporting context — a115

- CGC: a113·a114 에서 같은 워크트리 색인이 Go 심볼을 돌려주지 않음을 확인(`find name` → 0). not-applicable.
- GBrain: `gbrain_project.py search` → busy(owner pid 1564542, 다른 세션 소유, 2026-09-26). not-applicable.
- 대체 보조 문맥(codegraph_explore 소스 열람): `c.opts.StrategyRuntime` 소비자는 두 함수뿐 —
  `buildMultiMarketStrategyRuntimePage`(:48–57)와 `strategyRuntimeSummary`(:291–300). 둘 다 nil 검사 후
  Read+Validate, 오류는 화면 값(LoadErr / 「읽지 못함」)으로 간다. a114 의 lifecycle 소비자 6곳과 달리
  전략 reader 소비자는 이 둘이 전부다.
- 원형: `cmd/tossctl/httpapi_strategy_attach.go`(a109 D4) · `console_lifecycle_attach.go`(a114) ·
  테스트 `a109_the_request_path_never_dials_test.go`.
