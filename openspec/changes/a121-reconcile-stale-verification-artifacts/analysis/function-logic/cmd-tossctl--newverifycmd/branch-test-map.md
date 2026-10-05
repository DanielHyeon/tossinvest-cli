# Branch Test Map: `newVerifyCmd`

- Source: `cmd/tossctl/verify.go` (96-129)
- AST branches 0 — 행복 경로 한 행

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 루트 트리에서 `verify` 를 찾고 run·status·report 의 source/mutating 주석을 대조; 전 잎의 `mutating=true` 집합을 대조 | `TestVerifyCommandsAreRegisteredAndAnnotated` (`verify_test.go:234`) · `TestMutatingAnnotationOnTradeCommands` (`help_convention_test.go:96`) · `TestLeafCommandsHaveSourceAnnotation` (`help_convention_test.go:80`) | 해당 없음 — 편집 전 기준선(새 잎 등록 RED 는 task 2.x) | 기존 스위트 통과(base `de147cc2`) |
