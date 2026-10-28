# Function Logic Map: `adaptPrices`

- Source: `internal/official/market_reads.go` (168-183)
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- 기준: 편집 뒤(브랜치 `a163-impl` GREEN 로트, 2026-10-22). 편집 전 번들 문서는 커밋 `78efe1951` · 생성 입력 `../../flm-gen/spec-base.json`
- AST 분기 1 · 반환 1 · defer 0 · go 0
- 위험 등급: High-risk(청산 관측의 유일한 생산 시세 어댑터 — `Client.Prices` 경유 `ExitObserver.observe` B1)
- 출처: `archive/2026-10-22-a163-quote-freshness-uses-trade-time/analysis/function-logic/internal-official--adaptprices` 번들 복사, source sha 대조 일치 — a164 종결 주석(2026-10-28): a164 base `4cd7af19d` 가 그 함수를 고친 형제 착지보다 앞이라 창(base→워킹트리)에 들어온 함수. `sha256sum internal/official/market_reads.go` = `a2c2f01619e0…` 이 번들의 `source_sha256`(revision current)과 같아 번들을 그대로 복사함.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `raw []apiPrice` | 브로커 `/api/v1/prices` 행 0..n | 공식 OpenAPI `PriceResponse` | 디코드 실패는 호출자 `Client.Prices` 가 이미 오류로 반환 — 이 함수는 실패 경로 없음 |
| `p.Timestamp` | 임의 문자열(`""` = null·키 없음) | 브로커 데이터 시각 | 편집 전: 읽지 않음(AST 호출에 없음) |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | range 170 `for _, p := range raw {` | 행마다 `domain.Quote` 하나 추가(행 탈락 0) | 끝에서 `out` 반환 | `TestAdaptPricesUnit` · `TestAdaptPricesEmpty` |

반환 좌표(AST): 182

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `parseDecimal` | `lastPrice` 해독 | 실패 시 0 — 하류 `observe` B4 가 `Last <= 0` 을 거름 | AST |
| `time.Now().UTC()` | 받은 시각 `FetchedAt` | 실패 없음 | AST |

## State mutations and fallbacks

- `out` 슬라이스만 만든다. 상태 없음. 폴백 없음.

## a163 편집

- a163 3.1: 리터럴에 `SourceTimestamp: p.Timestamp` 대입 한 줄. 분기 1 · 반환 1 불변(design D3). 해독·거름·오류 0. — **적용됨**, 편집 전·뒤 AST 분기·반환·defer 수 대조는 `../../flm-gen/README.md`

## Safety conclusion

- Safe edit boundary: 리터럴 필드 대입만. `FetchedAt` 줄 불변.
- High-risk impact: yes — 행 탈락·오류 반환 변이(⑦⑧)는 e2e 2.1~2.3 이 잡는다.
