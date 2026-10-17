# Function Logic Map: `ExposureLimits`

- Source: `internal/app/engine/interlock_test.go` (49-49, 편집 뒤)
- Qualified: `stubGuardian.ExposureLimits`
- AST evidence: `ast.json` (revision: base) — 본문 무편집, 같은 파일의 이웃 선언 편집으로 검사기가 교체된 base 함수로 분류해 편집 전 AST 를 요구함.
- Risk scan: `risk-pattern-report.md`

**역할.** 시험 fixture/시험 함수(생산 코드 아님).

**a136 편집.** 같은 파일의 stubGuardian 타입에 dailyBuy 칸과 DailyBuyNotionalLimit 메서드를 더함(이 메서드 본문은 무편집 — 타입 선언 이동으로 위치만 바뀜)

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| — | — | 시험 | — |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| (없음) | 직선 함수 — AST branches 0 | — | — | `go test ./internal/app/engine` (이 함수를 쓰는 기존 시험) |

Exact AST return positions: 49

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| — | fixture 구성 | — | AST |

## State mutations and fallbacks

- 시험 데이터만.

## Safety conclusion

- Safe edit boundary: 칸 추가 — 시험이 단언하는 값 무변경.
- High-risk impact: no(시험 전용, 비례 원칙)
