# Branch Test Map: `TestAHeldRowIsNotWhispered`

- Source: `internal/obs/a099_round4_test.go`. 시험 코드 — 분기는 단언 갈래.

| Branch | AST anchor | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 131:2 | 단언 갈래 | 자기 자신 | 반증: 변이 L11(줄 삭제) → 이 시험 FAIL(`no engine.alert_claim_held line`), L12(WARN 복귀) → FAIL | PASS(`fbc6df5f`) |
| B2 | if at 134:2 | 단언 갈래 | 자기 자신 | 반증: 변이 L11(줄 삭제) → 이 시험 FAIL(`no engine.alert_claim_held line`), L12(WARN 복귀) → FAIL | PASS(`fbc6df5f`) |
| B3 | else at 136:9 | 단언 갈래 | 자기 자신 | 반증: 변이 L11(줄 삭제) → 이 시험 FAIL(`no engine.alert_claim_held line`), L12(WARN 복귀) → FAIL | PASS(`fbc6df5f`) |
| B4 | if at 136:9 | 단언 갈래 | 자기 자신 | 반증: 변이 L11(줄 삭제) → 이 시험 FAIL(`no engine.alert_claim_held line`), L12(WARN 복귀) → FAIL | PASS(`fbc6df5f`) |
| B5 | if at 140:2 | 단언 갈래 | 자기 자신 | 반증: 변이 L11(줄 삭제) → 이 시험 FAIL(`no engine.alert_claim_held line`), L12(WARN 복귀) → FAIL | PASS(`fbc6df5f`) |
| B6 | if at 143:2 | 단언 갈래 | 자기 자신 | 반증: 변이 L11(줄 삭제) → 이 시험 FAIL(`no engine.alert_claim_held line`), L12(WARN 복귀) → FAIL | PASS(`fbc6df5f`) |
| B7 | if at 148:2 | 단언 갈래 | 자기 자신 | 반증: 변이 L11(줄 삭제) → 이 시험 FAIL(`no engine.alert_claim_held line`), L12(WARN 복귀) → FAIL | PASS(`fbc6df5f`) |
| B8 | if at 154:2 | 단언 갈래 | 자기 자신 | 반증: 변이 L11(줄 삭제) → 이 시험 FAIL(`no engine.alert_claim_held line`), L12(WARN 복귀) → FAIL | PASS(`fbc6df5f`) |
| B9 | if at 158:2 | 단언 갈래 | 자기 자신 | 반증: 변이 L11(줄 삭제) → 이 시험 FAIL(`no engine.alert_claim_held line`), L12(WARN 복귀) → FAIL | PASS(`fbc6df5f`) |
| B10 | if at 161:2 | 단언 갈래 | 자기 자신 | 반증: 변이 L11(줄 삭제) → 이 시험 FAIL(`no engine.alert_claim_held line`), L12(WARN 복귀) → FAIL | PASS(`fbc6df5f`) |
