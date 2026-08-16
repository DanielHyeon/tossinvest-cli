# Branch Test Map: `PrepareRuntimeSocket`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 잘못된 socket 이름·디렉터리 | `TestDialRuntimeRejectsDescriptorSocketPathInjection`이 같은 `validateRuntimeSocketPath`의 이름 절을 잰다(이 함수 직접 커버는 없음) | no | no |
| B2 | 잔재 없음 → nil | `TestRuntimeOnlyUnixControlCrossNamespaceTopologyAndRouteAbsence`(첫 기동) | no | no |
| B3 | 그 밖의 Lstat 실패 | 커버 없음 — 디렉터리 검증을 통과한 뒤의 Lstat 실패를 만들 안정적인 방법이 없다 | no | no |
| B4 | 정규 파일은 지우지 않고 거부 / pre-chmod 0700 잔재 영구 거부 | `TestPrepareRuntimeSocketNeverDeletesNonSocket` (`internal/positionpolicyrpc/runtime_test.go:67`)가 정규 파일 절을 잰다. 0700 socket 절은 a109 §1.1 RED가 기동에서 관측한다 | yes(§1.1) | yes(§1.4) |
| B5 | 남의 uid 소유 | 커버 없음 — 비root 테스트는 소유자를 바꿀 수 없다 | no | no |
