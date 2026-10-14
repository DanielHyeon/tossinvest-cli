# Function Logic Map: `reservePlan.checkLimits`

- Source: `internal/journal/reservations.go` (675-711, base 7b073f55)
- Qualified: `reservePlan.checkLimits`
- 위치: `analysis/ast-evidence/` — **proposal 단계 근거(게이트 번들 아님)**. 편집하게 되는 함수는 구현 착수 때 `analysis/function-logic/` 에 정식 번들을 만든다.
- AST evidence: `ast.json` (`source_sha256` 933d37c95a6212fe…) — AST branches 7 · returns 6
- Risk scan: `risk-pattern-report.md`
- 용도: a136 spec 의 경계 주장 「총계 한도는 **도달 시 거부(≥)**」가 예약 트랜잭션의 기존 실장과 같다는 근거. 새 일일 매수 합산도 같은 경계를 쓴다(design D4).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `p.newTotals` | 종류별 이번 요청 합 | `build` | — |
| `held` | 계좌의 HELD 합 (`heldTotals`) | journal tx 안 | 통화 불일치 오류 |
| `p.usage` · `p.limits` | 같은 통화 | `build` | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `range p.newTotals` (676) | 지역 합 | — | B2~B7 |
| B2 | `held[kind]` 있음 (681) | 합에 더함 | — | `TestConcurrentDecisionsCannotBothTakeTheLastSlot` |
| B3 | held 통화 ≠ limit 통화 (682) | — | `ErrInvalidRequest` | 없음(기존 부채 — `TestMixedCurrenciesAreRefused` 는 build B24 를 잰다) |
| B4 | held 덧셈 실패 (689) | — | 오류 | 없음(기존 부채) |
| B5 | 새 합 덧셈 실패 (695) | — | 오류 | 없음(기존 부채) |
| B6 | 비교 실패 (701) | — | 오류 | 없음(기존 부채) |
| B7 | `cmp >= 0` (704) — **usage + held + new ≥ limit** | — | `ErrReservationLimitExceeded` | `TestReachingTheAggregateLimitIsRefused` |

Exact AST return positions: 683 · 690 · 696 · 702 · 705 · 710

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `riskcalc.AddDecimal` · `CompareDecimal` | 정확 decimal | 오류 반환 | AST calls |

## State mutations and fallbacks

- 없음. fallback 없음.

## Safety conclusion

- Safe edit boundary: a136 추천안에서 무편집. 새 합산은 이 경계(≥)를 따라야 하며 다르게 만들면 「총계 경계 ≥」 spec 문장과 갈린다.
- High-risk impact: no (무편집)
