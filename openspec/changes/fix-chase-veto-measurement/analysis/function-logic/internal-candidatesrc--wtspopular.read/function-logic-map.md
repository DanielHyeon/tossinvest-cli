# Function Logic Map: `wtsPopular.Read`

- Source: `internal/candidatesrc/candidatesrc.go`
- Change: `fix-chase-veto-measurement`
- AST evidence: `ast.json` (revision `current`, L448–489, 분기 4개)
- Risk scan: `risk-pattern-report.md`

WTS 인기 순위 한 판을 읽는다. 공식 순위와 같은 두 사실(`RankRequested`, `NewlyListed`)을
이 change가 더했고, 심볼 결정을 `wtsSymbol`로 뽑아냈다 — 기억 집합과 행이 **같은 문자열로**
키잉되어야 하기 때문이다.

시장 가드가 소스 자신에 있는 것은 기존 결정이다. 손으로 만든 패널이나 KR 패널을 US 스캔에
재사용하는 호출자가 한국 행을 US로 파일링하는 것을 막는다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `market` | `candidate.MarketKR`만 | 호출자 | 그 밖은 오류 — 클라이언트를 부르기 전에 거부 |
| `raw.Stocks` | 0..size 행 | WTS | 빈 읽기는 기억을 교체하지 않는다 |
| `total` | `len(raw.Stocks)` | 이 함수 | `RankTotal` |
| `w.size` | 요청 행 수 | `WTSPopular` | `RankRequested` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `m != candidate.MarketKR` | **클라이언트를 부르지 않는다** | `Reading{}, error` | `TestThePopularityRankingRefusesAMarketItCannotSee` |
| B2 | `err != nil` | 없음 — 기억 교체 이전 | `Reading{}, error` | `internal/candidatesrc`의 fake err 경로 |
| B3 | `for _, s := range raw.Stocks` | `rows` append | — | `TestTheWTSPopularityListReportsTheSameTwoFacts` |
| B4 | `symbol == ""` — Symbol도 ProductCode도 없는 행 | 행을 만들지 않고, 그래서 그 읽기는 온전하지 않다 | continue | `TestTheWTSMemoryIsAlsoBuiltFromTheRowsItKeeps` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `w.reader.GetStockRanking(ctx, w.size)` | 인기 순위 | 오류는 래핑 | ast.json calls |
| `w.rememberRead(market, raw.Stocks, w.size)` | 직전 집합 취득 + 조건부 교체(온전 여부는 그 안에서 판정 — issues.md I16) | 오류 없음 | ast.json calls |
| `wtsSymbol(s)` | 행의 식별자 | 순수 | ast.json calls |
| `newlyListed(previous, hadPrevious, symbol)` | 3-상태 | 순수 | ast.json calls |

## State mutations and fallbacks

- `w.seen[market]`을 `rememberRead`가 조건부로 교체.
- `Reading.Budget`은 비운다 — WTS는 rate 헤더를 보내지 않고, 미보고는 0이 아니다(기존 동작).
- 주문 경로 무접촉 — `PopularityReader`는 `GetStockRanking` 한 메서드다.

## Safety conclusion

- Safe edit boundary: 행 리터럴에 필드 2개 가산 + `wtsSymbol` 추출 + `rememberRead` 호출 1줄.
- High-risk impact: no (조회 전용). 이 소스는 가산적이며 세션 만료가 일상이므로 downstream은 부재를 견뎌야 한다.
