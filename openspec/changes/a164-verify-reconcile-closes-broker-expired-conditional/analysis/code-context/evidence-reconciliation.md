# a164 evidence reconciliation

| 질문 | CodeGraph | CGC | HEAD(AST·git grep) | 결론 |
| --- | --- | --- | --- | --- |
| `terminal()` 호출 자리 | 1(`joinChain`) | 낡음 | 4(record.go 712·722 · projection_reconciled.go 18 · runner.go 1052) | HEAD. 넷째 종결은 술어 한 곳 편집으로 네 자리에 다 닿음 |
| `Reconcile` 생산 호출자 | 8(이름 충돌 포함) | 낡음 | `runVerifyReconcile` 1 | HEAD |
| 미결 HOLD 지점 | — | 낡음 | cmd `validateM0TriggerMode` B9 `:428` · `New` B9 `:346` | AST |
| 전제 판정 순서 | — | 낡음 | cmd: B8 `M0ExactPrerequisites` `:424` 가 B9 `:428` 앞 / `New`: B9 `:346` 이 B15 `:358` 앞 | AST — a164 만으로 cmd 진입은 안 풀림(design D9) |
| Q1 정적 거절 | — | 낡음 | `Reconcile` B8 `:205` → `checkReconcileRetention` B1 `:118` | AST |
| Q6 거절 | — | 낡음 | `checkReconcileAbsence` B5 `:384` | AST |

불일치 둘(CodeGraph 누락·이름 충돌)은 HEAD 로 해소했다. 편집 차단 불일치 없음. 편집 직전 재확인은 tasks 0.9.
