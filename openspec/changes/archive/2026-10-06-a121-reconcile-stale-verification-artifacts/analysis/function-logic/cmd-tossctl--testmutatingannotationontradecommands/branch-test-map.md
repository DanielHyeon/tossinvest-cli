# Branch Test Map: `TestMutatingAnnotationOnTradeCommands`

- Source: `cmd/tossctl/help_convention_test.go` (96-160)
- AST branches 3. 시험 함수 자체라 「Test」 칸은 그것을 실행하는 대조(변이 원장)다.

| Branch | Anchor | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `range at 150:2` | 잎 명령 전수 | `TestMutatingAnnotationOnTradeCommands` | 해당 없음(시험 편집) | PASS |
| B2 | `if at 153:3` | 집합 안 미표지 | `TestMutatingAnnotationOnTradeCommands` | 변이 C-annotation 에서 발화 | PASS(무변이) |
| B3 | `if at 156:3` | 집합 밖 표지 | `TestMutatingAnnotationOnTradeCommands` | 집합 줄을 빼면 발화(편집 전 base 에 reconcile 이 없던 상태) | PASS(무변이) |
