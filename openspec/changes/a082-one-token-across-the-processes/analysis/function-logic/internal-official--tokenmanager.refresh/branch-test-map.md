# Branch Test Map: `tokenManager.refresh`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| happy | 편집 전: 401 뒤 무조건 교환. 편집 후에도 채택할 것이 없으면 같다 | `TestTokenRefresh` (손대지 않음), `TestARefusedProcessWithNothingToAdoptStillExchanges` | 후자 **yes** | yes |
| B1 | 디스크 토큰이 유효하고 방금 실패한 것과 **다르면** 채택한다 — 네트워크 0, 파일 쓰기 0. 같거나 없으면 교환한다 | `TestARefusedProcessAdoptsTheTokenAnotherProcessAlreadyGot`, `TestTwoProcessesSharingOneCacheFileStopBuyingTokensFromEachOther` | **yes** | yes |
