# Branch Test Map: `ValidateRuntimeControlDirectory`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | leaf 이름이 다른 디렉터리 거부 | `TestDialRuntimeRejectsDescriptorSocketPathInjection` (`internal/positionpolicyrpc/runtime_test.go:60`) — descriptor가 `../command.sock`을 가리키면 socket 경로 검증이 이 이름 절에서 걸린다 | no | no |
| B2 | 안전하지 않은 엔진 디렉터리 | 커버 없음(자매 계열의 `TestPositionPolicyControlRejectsInsecureControlFilesystem`이 같은 `ValidateEngineDirectory`를 잰다) | no | no |
