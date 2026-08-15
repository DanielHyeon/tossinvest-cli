# Branch Test Map: `PositionPolicyRuntimeServer.Close`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | nil 수신자 Close | 커버 없음 — 호출부가 항상 non-nil을 defer 한다 | no | no |
| B2 | Close가 descriptor·socket·controlDir을 전부 치운다 | `TestRuntimeOnlyUnixControlCrossNamespaceTopologyAndRouteAbsence` (세 경로의 부재를 확인한다) | no | no |
| B3 | ErrNotExist 관용 | `TestCommandAndRuntimeControlsHaveIndependentLifecycles`가 두 endpoint의 독립적 Close를 잰다 | no | no |
