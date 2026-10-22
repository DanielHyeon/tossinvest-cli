# CodeGraph baseline — a169 2.1 (`securityType` 노출)

- 도구: codegraph 1.6.0, 색인 = `make sdd-sync` 로 `/tmp/tossos-a169-20261023` 에서 재생성(HEAD `566e1cf7c`, 2026-10-23).
- `codegraph callers adaptStocks` → 4: `Client.Stocks`(stock_reads.go:44) · `TestA163TheOfficialStocksQuoteJSONIsByteIdenticalAndCarriesNoBrokerTime` ·
  `TestAdaptStocksUnit` · `TestAdaptStocksEmpty`.
- `codegraph callers Stocks` → 3: `readOperations`(internal/ops/read_operations.go:78 — `ops call stocks`) · `cmd/tossctl/instrument_names.go:36`
  (콘솔 종목명, `Name` 만 읽음) · `tools/a128-jev-shadow-probe/collect.go:29`(메타, `Name` · `MarketCode` · `Currency` 만 PublicState 로 복사).
- `codegraph affected internal/official/stock_reads.go internal/domain/models.go --filter '*_test.go'` → 1,224 건(stock_reads.go 단독 1,052).
  `domain` 은 거의 모든 패키지가 import 하므로 파일 단위 영향은 과대다 — 심볼 단위(위 callers)로 좁혀 시험을 골랐다.
