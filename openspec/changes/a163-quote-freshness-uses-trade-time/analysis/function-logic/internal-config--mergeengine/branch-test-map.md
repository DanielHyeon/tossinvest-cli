# Branch Test Map: `mergeEngine`

편집 전 분기 5 → 편집 뒤 6. 새 B4 가 들어와 옛 B4·B5 는 B5·B6 으로 밀렸다(분기 번호는 위치 — 편집 전·뒤 AST 를 줄 원문으로 대조해 재번호, `../../flm-gen/README.md`).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if 331 — engine 블록 없음 → 기본값 | `TestAConfigWithNoNotificationsBlockIsUnchanged` | n/a | pass |
| B2 | if 334 — autostart 병합 | `TestAutostartDefaultsOffAndLoadsExplicitValues` | n/a | pass |
| B3 | if 338 — exit_policy 병합·거부 | `TestUnknownCommonExitPolicyIsRetainedAsRejected` | n/a | pass |
| B4 | if 349 — (a163) 보고 토글 병합 — automation_gate 없는 파일 포함 | `TestA163TheSourceAgeReportToggleDefaultsOff` | yes(골격: explicit true · true without automation_gate 실패) | pass |
| B5 | if 352 — automation_gate 없음 → 조기 반환(옛 B4) | `TestNotificationsMergeWithoutAnAutomationGate` | n/a | pass |
| B6 | if 356 — gate.enabled 병합(옛 B5) | `TestAutomationGateParsed` | n/a | pass |
