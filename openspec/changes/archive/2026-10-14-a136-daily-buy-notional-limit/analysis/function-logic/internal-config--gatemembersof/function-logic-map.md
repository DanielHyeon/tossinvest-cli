# Function Logic Map: `gateMembersOf`

- Source: `internal/config/limits_io.go` (191-209, 편집 뒤)
- Qualified: `gateMembersOf`
- AST evidence: `ast.json`(편집 뒤) · `ast-base.json`(편집 전, base f504cc52, `source_sha256` 9d08b616b9a8c621…, 편집 전 분기 수 1) — 현재 branches 2 · returns 2
- Risk scan: `risk-pattern-report.md`

**역할.** 한도 저장이 쓰는 키의 닫힌 목록.

**a136 편집.** DailyBuy 가 있을 때만 `max_daily_buy_notional` 멤버 추가(`enabled` 는 여전히 목록 밖).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 인자 | — | 호출자 | — |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if err != nil {` (if 193) | — | 분기 본문 참조 | `TestA136TheLimitWriterTouchesTheDailyBuyKeyOnlyWhenAsked` · `TestTheSaveCarriesNoEnabledKeyOfItsOwn` |
| B2 | `if l.DailyBuy != nil {` (if 204) | — | 분기 본문 참조 | `TestA136TheLimitWriterTouchesTheDailyBuyKeyOnlyWhenAsked` · `TestTheSaveCarriesNoEnabledKeyOfItsOwn` |

Exact AST return positions: 194 · 208

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| — | — | — | AST |

## State mutations and fallbacks

- 설정 파일/audit(해당 시) — 기존과 같음.

## Safety conclusion

- Safe edit boundary: 칸 · 문장 하나 추가 — 기존 분기 무변경.
- High-risk impact: yes(설정 쓰기)
