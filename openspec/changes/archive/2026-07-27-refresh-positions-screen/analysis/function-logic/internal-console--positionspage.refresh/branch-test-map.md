# Branch Test Map: `positionsPage.Refresh`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | (무분기 happy path) positions 응답에 `content="30"` meta refresh 존재·history에는 없음, 검증 화면 2초 유지 | `TestThePositionsScreenAsksTheBrowserToReloadAtTheCacheTTL` + `TestTheVerificationScreensKeepTheirTwoSecondReload` | yes (meta 부재로 실패; 2초 가드는 hold-the-line) | yes |

RED/GREEN 실행 기록은 `internal-console--positionrow.reason/branch-test-map.md` 하단과 동일 세션.
