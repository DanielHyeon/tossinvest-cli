# Function Logic Map: `wtsSymbol`

- Source: `internal/candidatesrc/candidatesrc.go`
- Change: `fix-chase-veto-measurement`
- AST evidence: `ast.json` (revision `current`, L497–502, 분기 1개)
- Risk scan: `risk-pattern-report.md`

**신규 함수** — 루프 안 두 줄을 뽑아냈다. 뽑아낸 이유가 이 change의 요점이다:
직전 읽기의 집합이 **행이 보고되는 것과 같은 문자열로** 키잉되어야 한다.
`s.Symbol`로 만든 집합에 행은 `s.ProductCode`로 떨어지면, fallback 행은 **매 읽기마다**
신규 진입으로 보고된다 — 색인 방식이 만들어낸 영구 표식이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `s.Symbol` | 있으면 이것 | WTS 응답 | 공백만이면 fallback |
| `s.ProductCode` | Symbol이 없을 때 | WTS 응답 | 둘 다 없으면 빈 문자열 → 호출자가 행을 버린다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `strings.TrimSpace(s.Symbol) != ""` | 없음 | Symbol / 아니면 TrimSpace(ProductCode) | `TestThePopularityRankingFallsBackToTheProductCode` · `TestAWTSRowIdentifiedByItsProductCodeIsNotANewEntrantEveryTime` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strings.TrimSpace` | 양쪽 후보 정규화 | 순수 | ast.json calls |

## State mutations and fallbacks

- 없음. 순수 함수.
- fallback은 여기 하나뿐이고, 그것이 이 함수가 존재하는 이유다 — `Read`와 `rememberRead`가 같은 fallback을 각각 구현하면 어긋난다.

## Safety conclusion

- Safe edit boundary: 신규 함수(추출). 동작은 base의 인라인 두 줄과 같다.
- High-risk impact: no (문자열 선택). 두 호출부가 갈라지면 fallback 행 전부가 영구 `신규 진입`이 된다.
