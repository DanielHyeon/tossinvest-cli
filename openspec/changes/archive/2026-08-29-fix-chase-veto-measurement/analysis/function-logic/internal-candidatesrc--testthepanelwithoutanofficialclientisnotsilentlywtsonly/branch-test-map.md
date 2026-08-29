# Branch Test Map: `TestThePanelWithoutAnOfficialClientIsNotSilentlyWTSOnly`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 공식 없이 WTS만 | 자체 실행 | yes (컴파일) | yes |
| B2 | 클라이언트가 없으면 빈 패널 | 자체 실행 | — (기존 동작) | yes |
