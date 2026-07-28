# Function Logic Map: `WTSPopular`

- Source: `internal/candidatesrc/candidatesrc.go`
- Change: `fix-chase-veto-measurement`
- AST evidence: `ast.json` (revision `current`, L396–404, 분기 2개)
- Risk scan: `risk-pattern-report.md`

WTS 인기 순위 어댑터의 생성자. 이 change가 `clk clock.Clock`을 더했다 — 두 어댑터가 같은 두
조건(온전함, TTL)을 들어야 하고, 한쪽만 배우면 패널이 절반만 측정 가능해진다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `reader` | `PopularityReader` 한 메서드 | 호출자 | nil이면 `Panel`이 소스를 넣지 않는다 |
| `size` | 1 이상, 그 밖은 30 | 이 함수 | `RankRequested`가 되는 값 |
| `clk` | `clock.Clock` 또는 nil | 호출자 | nil은 `clock.System()` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `size <= 0` | `size = 30` | — | **없음** (아래 주석) |
| B2 | `clk == nil` | `clk = clock.System()` | — | `TestThePopularityRankingReportsNoTradingFigures` 등 nil을 넘기는 테스트 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `clock.System` | 기본 시계 | 오류 없음 | ast.json calls |

## State mutations and fallbacks

- `wtsPopular` 하나를 새로 만든다. `seen`은 nil로 시작하고 그것이 '직전 읽기 없음'이다.
- fallback: `size <= 0` → 30, `clk == nil` → 시스템 시계. 둘 다 주입 기본값이지 미측정의 접기가 아니다.

## Safety conclusion

- Safe edit boundary: 시그니처 가산(`clk`). 본문 로직 무변경.
- High-risk impact: no (읽기 전용 어댑터 생성).
