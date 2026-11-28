# Branch Test Map: `mergeEngine`

편집 전 좌표(AST 7 분기, `internal/config/engine.go` 344–386).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | engine 블록 없음 | 미측정 | n/a | no |
| B2 | autostart | 미측정 | n/a | no |
| B3 | exit_policy | `TestUnknownCommonExitPolicyIsRetainedAsRejected` | n/a | no |
| B4 | exit_source_age_report | a163 config 시험(미측정 이름) | n/a | no |
| B5 | cancel_external_opposite_buy | a197 config 시험(미측정 이름) | n/a | no |
| B6 | automation_gate 없음 | 미측정 | n/a | no |
| B7 | gate.enabled | 미측정 | n/a | no |

a208 F3: `TestA208FloorOnlyIgnoredReasonFollowsTheCommonPolicy`(HYBRID_50 · RUNNER → 빈 사유, BALANCED · 없음 → 사유, 토글 거짓 → 빈 사유).
