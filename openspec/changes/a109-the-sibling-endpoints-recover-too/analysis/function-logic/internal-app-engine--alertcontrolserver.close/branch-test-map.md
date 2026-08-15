# Branch Test Map: `AlertControlServer.Close`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | nil 수신자 Close | 커버 없음 — 호출부가 항상 non-nil을 defer 한다 | no | no |
| B2 | Close가 descriptor·socket·controlDir을 전부 치운다 | `TestTheAlertSocketIsPrivateToThisUser` (`a098_the_operator_socket_is_the_engines_test.go:113`가 세 경로의 부재를 확인한다) | no | no |
| B3 | 이미 사라진 경로(ErrNotExist)를 오류로 읽지 않는다 | `TestAlertControlServerStartsOverALeftover`의 `t.Cleanup` 이중 Close가 같은 관용에 의존한다 | no | no |
