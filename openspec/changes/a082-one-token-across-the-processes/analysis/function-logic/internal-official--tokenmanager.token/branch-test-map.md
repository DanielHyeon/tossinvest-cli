# Branch Test Map: `tokenManager.token`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 메모리 캐시 갈래. ① 유효하고 파일도 안 바뀌었으면 교환 없이 그 토큰. ② 유효하지만 **다른 프로세스가 파일을 바꿨으면** 믿지 않고 디스크로 내려간다 | ① `TestTokenExchangeAndCache` (손대지 않음) ② `TestTokenPrefersTheCacheFileWhenAnotherProcessRewroteIt` | ②만 **yes** | yes |
| B2 | 디스크 캐시가 유효하면 채택하고 `m.cache`에 싣는다. 무효면 교환으로 내려간다 | `TestTokenColdLoadFromDiskCache`, `TestTokenExchangeAndCache` (둘 다 손대지 않음) | no (기존 통과) | yes |
