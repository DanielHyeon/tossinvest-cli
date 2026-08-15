# Branch Test Map: `PositionPolicyRuntimeServer.Close`

구현 후 재최신화(§1.7). 분기 5개는 편집 **후**의 AST다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | nil 수신자 Close | 커버 없음 — 호출부가 항상 non-nil을 defer 한다 | no | no |
| B2 | listener를 Close가 직접 닫는다 | **결정적 커버 없음** — late unlink는 경합이다(a108도 300라운드 중 3회). 뮤테이션 M26의 자매 형태가 생존한 것이 그 사실의 측정이다(mutation-ledger-t1.md §B) | no | no |
| B3 | 이미 Shutdown이 닫았으면 `net.ErrClosed`는 성공 | `TestRuntimeOnlyUnixControlCrossNamespaceTopologyAndRouteAbsence`가 Close의 오류 없음을 확인한다 | no | yes(§1.4) |
| B4 | Close가 descriptor·socket·controlDir을 전부 치운다 | 같은 테스트가 세 경로의 부재를 확인한다 | no | no |
| B5 | ErrNotExist 관용 | `TestCommandAndRuntimeControlsHaveIndependentLifecycles`가 두 endpoint의 독립적 Close를 잰다 | no | no |
