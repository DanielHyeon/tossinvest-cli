# Branch Test Map: `PositionPolicyCommandServer.Close`

a109는 이 함수를 편집하지 않는다(맵의 Safety conclusion 참조). RED/GREEN 전부 no다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | nil 수신자 Close | 커버 없음 | no | no |
| B2 | Close 후 descriptor 부재 | `TestEngineOwnsAuthenticatedPositionPolicyControlEndpoint` · `TestCommandAndRuntimeControlsHaveIndependentLifecycles` | no | no |
| B3 | Close 후 control 디렉터리 부재 | `TestEngineOwnsAuthenticatedPositionPolicyControlEndpoint` (엔진 디렉터리에 엔트리 0을 확인한다) | no | no |
