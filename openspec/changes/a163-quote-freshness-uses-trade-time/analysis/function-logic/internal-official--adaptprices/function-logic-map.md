# Function Logic Map: `adaptPrices`

- Source: `internal/official/market_reads.go` (168-180)
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- 기준: 편집 전(base `f7d80c348` = HEAD `1eab9b931` 의 `internal` diff 0)
- AST 분기 1 · 반환 1 · defer 0 · go 0
- 위험 등급: High-risk(청산 관측의 유일한 생산 시세 어댑터 — `Client.Prices` 경유 `ExitObserver.observe` B1)

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `raw []apiPrice` | 브로커 `/api/v1/prices` 행 0..n | 공식 OpenAPI `PriceResponse` | 디코드 실패는 호출자 `Client.Prices` 가 이미 오류로 반환 — 이 함수는 실패 경로 없음 |
| `p.Timestamp` | 임의 문자열(`""` = null·키 없음) | 브로커 데이터 시각 | 편집 전: 읽지 않음(AST 호출에 없음) |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | range 170 `for _, p := range raw {` | 행마다 `domain.Quote` 하나 추가(행 탈락 0) | 끝에서 `out` 반환 | `TestAdaptPricesUnit` · `TestAdaptPricesEmpty` |

반환 좌표(AST): 179

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `parseDecimal` | `lastPrice` 해독 | 실패 시 0 — 하류 `observe` B4 가 `Last <= 0` 을 거름 | AST |
| `time.Now().UTC()` | 받은 시각 `FetchedAt` | 실패 없음 | AST |

## State mutations and fallbacks

- `out` 슬라이스만 만든다. 상태 없음. 폴백 없음.

## a163 편집

- a163 3.1: 리터럴에 `SourceTimestamp: p.Timestamp` 대입 한 줄. 분기 1 · 반환 1 불변(design D3). 해독·거름·오류 0.

## Safety conclusion

- Safe edit boundary: 리터럴 필드 대입만. `FetchedAt` 줄 불변.
- High-risk impact: yes — 행 탈락·오류 반환 변이(⑦⑧)는 e2e 2.1~2.3 이 잡는다.
