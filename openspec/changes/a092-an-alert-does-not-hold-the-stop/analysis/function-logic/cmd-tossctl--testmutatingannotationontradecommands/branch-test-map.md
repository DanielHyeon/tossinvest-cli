# Branch Test Map: `TestMutatingAnnotationOnTradeCommands`

- Source: `cmd/tossctl/help_convention_test.go`. 시험 코드.

| Branch | AST anchor | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | range at 144:2 | 단언 · 픽스처 갈래 | 자기 자신 | 반증: 변이 M21 · M22 에서 표 쪽 기대(`expected mutating=true`)가 FAIL | PASS(`2714e393`) |
| B2 | if at 147:3 | 단언 · 픽스처 갈래 | 자기 자신 | 반증: 변이 M21 · M22 에서 표 쪽 기대(`expected mutating=true`)가 FAIL | PASS(`2714e393`) |
| B3 | if at 150:3 | 단언 · 픽스처 갈래 | 자기 자신 | 반증: 변이 M21 · M22 에서 표 쪽 기대(`expected mutating=true`)가 FAIL | PASS(`2714e393`) |
