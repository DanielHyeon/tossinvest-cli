# a198 evidence reconciliation (tasks 1.2)

- 기준 커밋 `e1eaf720c` · CodeGraph 1.6.0 · CodeGraphContext not-applicable(codegraphcontext-context.md) · logic-map AST(`analysis/cited-ast/`).

| 항목 | CodeGraph | 현재 HEAD(grep·AST) | 결론 |
|---|---|---|---|
| `clearTheSymbol` 호출자 | `record` 1 함수 | 자리 1(exitloop.go:1395) | 일치 |
| `IssueReduction` 호출자 | 9 — RiskGuardian·Issuer·인터페이스·시험 대역 혼합(이름 기반) | 엔진 생산 자리 2(:1517 submit · :1684 clearTheSymbol) | CodeGraph 는 수신자를 가르지 않음 — 자리는 grep 이 권위. 편집 자리는 :1684 하나 |
| `clearTheSymbol` 피호출 | `Context`(accountbudget/prewire_fence_linux.go:37) 포함 | 그런 호출 없음(`ctx` 인자 이름 혼동) | CodeGraph 거짓 엣지 — 무시 |
| `Cancel` 호출자 | 16(protection·ops·trading 시험 혼합) | 엔진 생산 자리 1(:1706), flatten 1(flatten.go:382 `s.Gateway.Cancel`) | 이름 혼합, 자리는 grep |
| `Context.ExitObserver` 호출자 | 31 | 생산 1(cmd/tossctl/engine.go:671) | 일치 |
| 취소 결정의 「종류」 | (해당 없음) | `journal.Decision` 에 Kind 필드 0, 종류는 DecisionRequest_build B22 의 키 유무로만 남음 | proposal P5 정정의 근거(claim-check.md) |
| 시험 발견 | `affected` 기본 1, `--filter '*_test.go'` 1,349 | — | 하네스 주석의 판별식 결함 재현. 시험 선택은 이름 grep(claim-check·edit-surface) |
| `warnEpisode` 의 싱크 | 피호출 `log` | `o.log` 의 `opts.Log == nil → return`(exitloop.go:1963), 생산 `opts.Log` 미설정 | design D4 가 운영에서 무기록 — Manager 결정 필요(claim-check D4) |

불일치는 전부 CodeGraph 의 이름 기반 해소에서 왔고, 현재 HEAD 로 해소했다. 편집을 막는 미해소 불일치: 없음. 편집 전 Manager 결정이 필요한 사실: D4 싱크, D2 안 선택(edit-surface.md §3), a087 착지 순서(edit-surface.md §4).
