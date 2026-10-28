# Function Logic Map: `adaptStocks`

- Source: `internal/official/stock_reads.go` (85-103)
- AST evidence: `ast.json` (`revision: current`, 편집 뒤). 편집 전 AST 는 `../../ast-evidence/internal-official-stock_reads.go--adaptStocks.json`
  (81-97, `source_sha256` `fa091d96…`) — base `2bbee345f` 에서 다시 뽑아 그 파일과 동일함을 확인(2026-10-23).
- Risk scan: `risk-pattern-report.md`
- AST 분기 1 · 반환 1 · defer 0 · go 0 — 편집 전 · 뒤 같음.
- 위험 등급: Normal — 읽기 전용 메타 어댑터. 주문 · 손절 · 사이징 · 인증 · 원장 경로 밖(호출자: `Client.Stocks` →
- 출처: `a169-breakout-universes-are-compared-in-shadow/analysis/function-logic/internal-official--adaptstocks` 번들 복사, source sha 대조 일치 — a164 종결 주석(2026-10-28): a164 base `4cd7af19d` 가 그 함수를 고친 형제 착지보다 앞이라 창(base→워킹트리)에 들어온 함수. `sha256sum internal/official/stock_reads.go` = `011792c585da…` 이 번들의 `source_sha256`(revision current)과 같아 번들을 그대로 복사함.
  `ops call stocks` · 콘솔 종목명 `instrument_names.go` · a128 도구 메타). 비례 원칙상 변이 원장 · 다중 적대 리뷰 불필요.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `raw []apiStockInfo` | 브로커 `/api/v1/stocks` 행 0..n | 공식 OpenAPI `StockInfo` | 디코드 실패는 호출자 `Client.Stocks` 가 오류로 반환 — 이 함수는 실패 경로 없음 |
| `s.SecurityType` | 임의 문자열(`""` = 키 없음 · null) | OpenAPI enum `STOCK · FOREIGN_STOCK · DEPOSITARY_RECEIPT · INFRASTRUCTURE_FUND · REIT · ETF · FOREIGN_ETF · ETN · STOCK_WARRANTS` | 해석 · 거름 없음 — enum 밖 값도 원문 그대로(분류는 소비자) |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | range 87 `for _, s := range raw {` | 행마다 `domain.Quote` 하나 추가(행 탈락 0), `SecurityType` 원문 대입 | 끝(102)에서 `out` 반환 | `TestA169StocksCarriesTheBrokerSecurityTypeVerbatim` · `TestA169TheStocksQuoteJSONWithASecurityTypeIsPinned` · `TestAdaptStocksUnit` · `TestAdaptStocksEmpty` |

반환 좌표(AST): 102

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `make` · `len` · `append` | 결과 슬라이스 | 실패 없음 | AST |
| `time.Now().UTC()` | 받은 시각 `FetchedAt` | 실패 없음 | AST |

live config binding 없음.

## State mutations and fallbacks

- `out` 슬라이스만 만든다. 상태 없음. 폴백 없음.

## a169 편집

- a169 2.1: 리터럴에 `SecurityType: s.SecurityType` 대입 한 줄 + `domain.Quote.SecurityType string \`json:"security_type,omitempty"\``.
  분기 · 반환 수 불변. 해독 · 거름 · 오류 0.

## Safety conclusion

- Safe edit boundary: 리터럴 필드 대입만. 다른 필드 · `FetchedAt` 줄 불변.
- High-risk impact: no — 값을 채우지 않는 생산자(WTS `GetQuote` · `/prices`)의 JSON 은 키가 없어 바이트 동일
  (`TestA163TheWTSQuoteJSONIsByteIdenticalAndCarriesNoBrokerTime` · `TestA169ThePricesQuoteJSONCarriesNoSecurityType`),
  값 없는 `/stocks` 행도 편집 전 골든과 동일(`TestA163TheOfficialStocksQuoteJSONIsByteIdenticalAndCarriesNoBrokerTime`).
