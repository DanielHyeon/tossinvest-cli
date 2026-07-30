# Branch Test Map: `New`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | StartVerify 없이 New | `TestNewRefusesAConsoleWithNoWayToRunAVerification` | — | yes |
| B2 | Now 미주입 콘솔이 기동한다 | `TestListenBindsTheLoopbackInterface` 외 전 harness 케이스 | — | yes |
| B3 | Out 미주입 콘솔 | 없음 — 이 패키지의 모든 생성이 `Out`을 설정한다(console_test.go:109, restart_test.go:229). 무변경 분기이므로 이 change의 회귀 위험 아님 | — | n/a |
| B4 | Binary 미주입 콘솔이 `binstamp.Self`로 지문을 뜬다 | 전 harness 케이스 + `TestTheStaleEngineBinaryIsWarnedAbout` | — | yes |
