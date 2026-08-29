# Branch Test Map: `tokenManager.refresh`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 파일에 거부당한 것과 다른 유효 토큰이 있으면 채택 — 교환 0, 파일 쓰기 0. 형제 goroutine이 먼저 바꿔 둔 경우도 여기서 잡힌다 (`exchange`가 반환 전에 파일을 쓰므로) | `TestARefusedProcessAdoptsTheTokenAnotherProcessAlreadyGot`, `TestARotationThatLandsMidRequestCostsNoToken`, `TestASiblingGoroutineThatAlreadyReplacedTheTokenIsNotOutbought`, `TestTwoProcessesSharingOneCacheFileStopBuyingTokensFromEachOther` | **yes** (M1: 앞의 셋 + 헤드라인) | yes |
| fallthrough | 채택할 것이 없거나 거부당한 것과 같으면 교환 | `TestTokenRefresh`(기존, 손대지 않은 판정), `TestARefusedProcessWithNothingToAdoptStillExchanges` | **yes** (M4) | yes |
