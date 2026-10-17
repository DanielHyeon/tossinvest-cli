# Function Logic Map: `Validate`

- Source: `internal/config/limits.go` (326-349, 편집 뒤)
- Qualified: `GuardianLimits.Validate`
- AST evidence: `ast.json`(편집 뒤) · `ast-base.json`(편집 전, base f504cc52, `source_sha256` 6817ccd8ebcd4480…, 편집 전 분기 수 6) — 현재 branches 7 · returns 6
- Risk scan: `risk-pattern-report.md`

**역할.** 콘솔 한도 저장 전 검증(엔진이 거부할 블록 거부).

**a136 편집.** DailyBuy 가 있으면 유한 · 0 이상 검사(0 = 미설정).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 인자 | — | 호출자 | — |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `for _, f := range guardianFields() {` (range 327) | — | 분기 본문 참조 | `TestA136TheLimitWriterTouchesTheDailyBuyKeyOnlyWhenAsked` · 기존 Validate 시험 |
| B2 | `switch {` (switch 329) | — | 분기 본문 참조 | `TestA136TheLimitWriterTouchesTheDailyBuyKeyOnlyWhenAsked` · 기존 Validate 시험 |
| B3 | `case math.IsNaN(v) \|\| math.IsInf(v, 0):` (case 330) | — | 분기 본문 참조 | `TestA136TheLimitWriterTouchesTheDailyBuyKeyOnlyWhenAsked` · 기존 Validate 시험 |
| B4 | `case v <= 0:` (case 332) | — | 분기 본문 참조 | `TestA136TheLimitWriterTouchesTheDailyBuyKeyOnlyWhenAsked` · 기존 Validate 시험 |
| B5 | `if l.MaxDailyLossRatio > 1 {` (if 337) | — | 분기 본문 참조 | `TestA136TheLimitWriterTouchesTheDailyBuyKeyOnlyWhenAsked` · 기존 Validate 시험 |
| B6 | `if strings.TrimSpace(l.Currency) == "" {` (if 341) | — | 분기 본문 참조 | `TestA136TheLimitWriterTouchesTheDailyBuyKeyOnlyWhenAsked` · 기존 Validate 시험 |
| B7 | `if l.DailyBuy != nil && (math.IsNaN(*l.DailyBuy) \|\| math.IsInf(*l.DailyBuy, 0) \|\| *l.DailyBuy < 0) {` (if 345) | — | 분기 본문 참조 | `TestA136TheLimitWriterTouchesTheDailyBuyKeyOnlyWhenAsked` · 기존 Validate 시험 |

Exact AST return positions: 331 · 333 · 338 · 342 · 346 · 348

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| — | — | — | AST |

## State mutations and fallbacks

- 설정 파일/audit(해당 시) — 기존과 같음.

## Safety conclusion

- Safe edit boundary: 칸 · 문장 하나 추가 — 기존 분기 무변경.
- High-risk impact: yes(§0.5 한도 설정)
