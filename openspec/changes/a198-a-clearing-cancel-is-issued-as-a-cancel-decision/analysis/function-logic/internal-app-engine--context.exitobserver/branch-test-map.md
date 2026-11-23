# Branch Test Map: `Context.ExitObserver` (pre-edit)

덮음 측정: clearTheSymbol 번들과 같은 선택 실행 coverprofile(블록 계수) + grep. 선택 실행에서 B1-B3 블록 계수 0 — 아래 시험은 그 선택 밖이라 **이름 grep 근거, 실행 미측정**.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | nil Context | grep 0 건 (`ErrExitObserverUnavailable` 를 쓰는 시험 2개 중 해당 없음) | n/a | 미덮음 |
| B2 | 게이트 미검증 | TestTheExitObserverIsUnavailableWithoutAVerifiedGate (tracer_test.go:338) | n/a | 미측정(선택 밖) |
| B3 | Guardian 이 발행기 아님 | TestEngineRuntimeConstructionBranchesFailClosedAndAssembleExactSuccess 「B3 exit observer construction」(cmd/tossctl/engine_runtime_branch_test.go:64) | n/a | 미측정(선택 밖) |
| B4 | Names nil | 선택 실행 덮음(344-346 계수 1) | n/a | 기존 green |
| B5 | Notifier 있음 | 선택 실행 덮음(349-354 계수 1) · a092/a094 wiring | n/a | 기존 green |
| B6 | Floor nil | 선택 실행 덮음(361-363 계수 1) | n/a | 기존 green |
| (예정) | 취소 결정 발행기 공급·부재 | a198 §4-7 | 예정 | 예정 |
