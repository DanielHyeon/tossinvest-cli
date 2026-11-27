# Branch Test Map: `TestTheReadOnlyHandleHasNoWriteMethods`

- Source: `internal/journal/readonly_test.go` (116-167)
- AST branches 2. 시험 함수 자체라 「Test」 칸은 그것을 실행하는 관측임.

| Branch | Anchor | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `for at 160:2` | 메서드 전수 | `TestTheReadOnlyHandleHasNoWriteMethods` | 해당 없음(시험 편집) | PASS |
| B2 | `if at 162:3` | 집합 밖 메서드 | `TestTheReadOnlyHandleHasNoWriteMethods` | 줄 추가 전 패키지 시험에서 FAIL(「unlisted method "ExitEvidenceCompactionPlan"」) | PASS |
