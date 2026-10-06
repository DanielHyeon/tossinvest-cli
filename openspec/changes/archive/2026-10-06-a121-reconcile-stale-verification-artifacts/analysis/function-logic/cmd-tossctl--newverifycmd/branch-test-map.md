# Branch Test Map: `newVerifyCmd`

- Source: `cmd/tossctl/verify.go` (96-130)
- AST branches 0 — 행복 경로 한 행

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 루트 트리에서 `verify` 를 찾고 run·status·report 의 source/mutating 주석을 대조; 전 잎의 `mutating=true` 집합을 대조 | `TestVerifyCommandsAreRegisteredAndAnnotated` (`verify_test.go:234`) · `TestMutatingAnnotationOnTradeCommands` (`help_convention_test.go:96`) · `TestLeafCommandsHaveSourceAnnotation` (`help_convention_test.go:80`) | 해당 없음 — 편집 전 기준선(새 잎 등록 RED 는 task 2.x) | 기존 스위트 통과(base `de147cc2`) |

## GREEN 로트 재추출 (a121 tasks 3.1·3.2, worktree a99a9059 + GREEN 편집)

- 편집: AddCommand 에 `newVerifyReconcileCmd(root)` 한 줄(분기 0 유지, 호출 6→7). GREEN 관측: TestVerifyReconcileIsRegisteredAsAMutatingOfficialCommand·TestMutatingAnnotationOnTradeCommands PASS; 변이 C-register.
- AST 재추출: `go run ./tools/logic-map` — 분기 0→0, 편집 전후 분기 열을 difflib 로 정렬한 결과 equal 뿐(재번호 없음), 앵커는 줄 이동만 반영.
