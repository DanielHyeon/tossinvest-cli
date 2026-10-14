# Function Logic Map: `LimitRows`

- Source: `internal/console/settings_limits.go` (61-71, 편집 뒤)
- Qualified: `settingsPage.LimitRows`
- AST evidence: `ast.json`(편집 뒤) · `ast-base.json`(편집 전, base f504cc52, `source_sha256` 021a96457d9b0b17…, 편집 전 분기 수 0) — 현재 branches 0 · returns 1
- Risk scan: `risk-pattern-report.md`

**역할.** 현재 한도 표.

**a136 편집.** 「하루 매수 총액 상한」 행 추가 — 미설정이면 KRW 기본값 적용 · USD 기동 거부를 평이하게.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 인자 | — | 호출자 | — |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| (없음) | 직선 함수 — AST branches 0 | — | — | `TestA136TheScreenSaysWhatAnUnsetDailyBuyCeilingDoes` · `TestUnsetLimitsRenderAsUnsetAndNotAsTheDefault` |

Exact AST return positions: 63

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| — | — | — | AST |

## State mutations and fallbacks

- 설정 파일/audit(해당 시) — 기존과 같음.

## Safety conclusion

- Safe edit boundary: 칸 · 문장 하나 추가 — 기존 분기 무변경.
- High-risk impact: no(표시)
