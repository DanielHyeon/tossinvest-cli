# Branch Test Map: `adaptStocks`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | range 87 — 행마다 Quote 하나, `securityType` 원문 보존(ETF · ETN · STOCK · enum 밖), 키 없는 행은 출력 키 없음 | `TestA169StocksCarriesTheBrokerSecurityTypeVerbatim` · `TestA169UnknownSecurityTypeIsPassedThroughUnjudged` · `TestA169TheStocksQuoteJSONWithASecurityTypeIsPinned` · `TestA163TheOfficialStocksQuoteJSONIsByteIdenticalAndCarriesNoBrokerTime` · `TestAdaptStocksUnit` · `TestAdaptStocksEmpty` | yes(2.1 — 편집 전 HEAD `566e1cf7c` 에서 TestA169 셋 FAIL: `security_type` 키 없음) | pass |
