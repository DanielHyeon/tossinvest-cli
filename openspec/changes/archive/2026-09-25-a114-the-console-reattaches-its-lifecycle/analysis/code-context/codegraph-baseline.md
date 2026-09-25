# CodeGraph baseline — a114

- 날짜: 2026-09-25 · 기준 HEAD `634cf3c5` · codegraph **1.6.0**

| 질의 | 결과 |
|---|---|
| `callers runConsole` | 1: `newConsoleCmd`(console.go:121) |
| `callers consolePositionPolicyCommander` | 2: `runConsole` · `TestConsolePositionPolicyCommanderKeepsRuntimeAuthoritySeparate` |
| `callers quarantineClient` | 3: 격리 메서드 셋(exit_quarantine_commander.go:34·42·51) |
| `callers requestCancelled` | 1: `strategyRuntimeAttachment.observe`(httpapi_strategy_attach.go:274) — a114 가 두 번째 호출자가 된다(함수 무편집) |
| `impact consolePositionPolicyCommander` | 52 심볼 — List/Preview/Apply 인터페이스 이름 해소로 engine 서버·ops CLI·internal/console 까지 퍼진다(이름 기반 과대 추정, reconciliation 참조) |
| `affected console.go position_policy_commander.go` 기본 | Go 테스트 **0** (하네스 주석의 `isTestPath` 결함) |
| `affected … --filter '*_test.go'` | **860** 파일 |
