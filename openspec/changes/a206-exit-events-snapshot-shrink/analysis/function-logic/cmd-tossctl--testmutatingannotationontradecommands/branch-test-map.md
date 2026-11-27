# Branch Test Map: `TestMutatingAnnotationOnTradeCommands`

- Source: `cmd/tossctl/help_convention_test.go` (96-163)
- AST branches 3. 시험 함수 자체라 「Test」 칸은 그것을 실행하는 관측임.

| Branch | Anchor | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `range at 153:2` | 잎 명령 전수 | `TestMutatingAnnotationOnTradeCommands` | 해당 없음(시험 편집) | PASS |
| B2 | `if at 156:3` | 집합 안 미표지 | `TestMutatingAnnotationOnTradeCommands` | 미관측(새 명령은 처음부터 표지를 가짐) | PASS |
| B3 | `if at 159:3` | 집합 밖 표지 | `TestMutatingAnnotationOnTradeCommands` | 집합 줄 추가 전 「unexpected mutating=true」 FAIL | PASS |
