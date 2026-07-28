# Function Logic Map: `Panel`

- Source: `internal/candidatesrc/candidatesrc.go`
- Change: `fix-chase-veto-measurement`
- AST evidence: `ast.json` (revision `current`, L500–527, 분기 4개)
- Risk scan: `risk-pattern-report.md`

한 시장의 소스 목록을 만든다. 이 change가 **`clk clock.Clock` 인자**를 더해 두 어댑터 모두에
전달한다 — 패널이 만드는 소스가 자기 기억의 나이를 잴 수 있어야 한다.

시장별 멤버십은 기존 결정이고 이 change가 바꾸지 않는다: 스캔은 "응답했는데 심볼을 싣지
않았다"를 심볼이 목록을 떠난 증거로 읽으므로, 구조적으로 그 시장을 볼 수 없는 소스는
present-and-empty가 아니라 **부재**여야 한다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `market` | 정규화된 시장 코드 | 호출자 | `MarketKR`만 WTS를 얻는다 |
| `official` | `RankingReader` 또는 nil | `cmd/tossctl/candidatepanel.go` | nil이면 공식 세 소스 부재 |
| `budget` | `BudgetReader` 또는 nil | 같은 곳 | nil은 예산 미보고 |
| `wts` | `PopularityReader` 또는 nil | 같은 곳 | nil이면 WTS 부재 — typed-nil 방지는 호출부의 `wtsPopularityReader` |
| `clk` | `clock.Clock` 또는 nil | 호출자 | nil은 각 생성자가 시스템 시계로 대체 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `official != nil` | 세 ranking type을 append | — | `TestThePanelWithoutAnOfficialClientIsNotSilentlyWTSOnly` |
| B2 | `for _, typ := range {TradingAmount, TradingVolume, TopGainers}` | 타입당 소스 하나 | — | `TestEveryPanelSourceHasItsOwnID` |
| B3 | `err == nil` | 성공한 것만 append | — | `TestEveryPanelSourceHasItsOwnID`(간접 — 셋 다 있어야 통과) |
| B4 | `wts != nil && market == candidate.MarketKR` | WTS 소스 append | — | `TestTheUSPanelDoesNotIncludeTheKoreanPopularityRanking` |

B3의 false arm은 **도달 불가**다 — 세 type은 이 파일의 컴파일 타임 상수이고 전부
`rankingSourceID`에 있다. `TestEveryPanelSourceHasItsOwnID`가 하나라도 빠지면 실패하므로,
도달 불가라는 주장이 테스트로 서 있다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `OfficialRanking(official, budget, typ, 100, clk)` | 공식 소스 셋 | 오류는 버리고 `TestEveryPanelSourceHasItsOwnID`가 잡는다 | ast.json calls |
| `WTSPopular(wts, 30, clk)` | KR 인기 순위 | 오류 없음 | ast.json calls |
| `strings.ToUpper`/`TrimSpace` | 시장 정규화 | 순수 | ast.json calls |

## State mutations and fallbacks

- 없음 — 슬라이스 하나를 만들어 돌려준다.
- 리터럴 100과 30은 `panelsize_drift_test.go`가 **소스로 읽는** 두 숫자다. 이 두 줄이 이 저장소에서 패널 크기의 정본이다.
- 주문 경로 무접촉 — 받는 것은 두 개의 좁은 read 인터페이스뿐이다.

## Safety conclusion

- Safe edit boundary: 시그니처 가산(`clk`) + 두 생성자 호출에 인자 1개. 멤버십 규칙 무변경.
- High-risk impact: no (배선). `MarketKR` 가드가 풀리면 US 후보가 매 스캔 냉각되고 냉각 시계가 first_seen_at을 지운다 — 기존 위험이며 이 change가 건드리지 않았다.
