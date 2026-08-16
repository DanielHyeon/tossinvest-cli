# Branch Test Map: `ValidateRuntimeSocket`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | descriptor가 다른 socket을 가리킴 | `TestDialRuntimeRejectsDescriptorSocketPathInjection` (`internal/positionpolicyrpc/runtime_test.go:60`) | no | no |
| B2 | socket 파일 부재 | `TestCommandAndRuntimeControlsHaveIndependentLifecycles`가 Close 후 부재를 확인하는 인접 경로를 잰다(직접 커버는 없음) | no | no |
| B3 | 0700 socket 거부(정확-0600 유지) | a109 §1.5의 클라이언트 검증 핀이 이 절을 직접 잰다(뮤테이션 M-C1이 완화 누출을 검사한다) | no | yes |
