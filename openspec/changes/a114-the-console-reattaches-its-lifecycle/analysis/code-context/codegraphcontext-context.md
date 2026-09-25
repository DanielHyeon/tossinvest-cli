# CodeGraphContext supporting context — a114

- CGC: a113 에서 같은 워크트리 색인이 Go 심볼을 돌려주지 않음을 확인(`find name` → 0). not-applicable.
- GBrain: `gbrain_project.py search` → exit 75 busy(다른 세션 소유). not-applicable.
- 대체 보조 문맥(HEAD grep): `c.opts.PositionPolicies` 소비자 6곳 — `exit_quarantine.go:88`,
  `portfolio_pages.go:99`, `position_policy.go:225·236·362·393`, `settings_tabs.go:268`. 전부 nil 검사 후
  메서드 호출이며 오류는 화면 값(LoadErr·refuse)으로 간다.
- 원형: `cmd/tossctl/httpapi_strategy_attach.go`(a109 D4) · 테스트 `a109_the_request_path_never_dials_test.go` 등.
